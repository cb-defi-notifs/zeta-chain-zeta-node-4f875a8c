package zetaclient

import (
	"encoding/base64"
	"encoding/hex"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zeta-chain/zetacore/common"
	"github.com/zeta-chain/zetacore/x/zetacore/types"
	"github.com/zeta-chain/zetacore/zetaclient/config"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func (co *CoreObserver) ShepherdManager() {
	numShepherds := 0
	for {
		select {
		case send := <-co.sendNew:
			if _, ok := co.shepherds[send.Index]; !ok {
				log.Info().Msgf("shepherd manager: new send %s", send.Index)
				co.shepherds[send.Index] = true
				log.Info().Msg("waiting on a signer slot...")
				<-co.signerSlots
				log.Info().Msg("got a signer slot! spawn shepherd")
				go co.shepherdSend(send)
				numShepherds++
				log.Info().Msgf("new shepherd: %d shepherds in total", numShepherds)
			}
		case send := <-co.sendDone:
			delete(co.shepherds, send.Index)
			numShepherds--
			log.Info().Msgf("remove shepherd: %d shepherds left", numShepherds)
		}
	}
}

// Once this function receives a Send, it will make sure that the send is processed and confirmed
// on external chains and ZetaCore.
// FIXME: make sure that ZetaCore is updated when the Send cannot be processed.
func (co *CoreObserver) shepherdSend(send *types.Send) {
	startTime := time.Now()
	confirmDone := make(chan bool, 1)
	coreSendDone := make(chan bool, 1)
	numQueries := 0
	keysignCount := 0

	defer func() {
		elapsedTime := time.Since(startTime)
		if keysignCount > 0 {
			log.Info().Msgf("shepherd stopped: numQueries %d; elapsed time %s; keysignCount %d", numQueries, elapsedTime, keysignCount)
			co.fileLogger.Info().Msgf("shepherd stopped: numQueries %d; elapsed time %s; keysignCount %d", numQueries, elapsedTime, keysignCount)
		}
		co.signerSlots <- true
		co.sendDone <- send
		confirmDone <- true
		coreSendDone <- true
	}()

	myid := co.bridge.keys.GetSignerInfo().GetAddress().String()
	amount, ok := new(big.Int).SetString(send.ZetaMint, 0)
	if !ok {
		log.Error().Msg("error converting MBurnt to big.Int")
		return
	}

	var to ethcommon.Address
	var err error
	var toChain common.Chain
	if send.Status == types.SendStatus_PendingRevert {
		to = ethcommon.HexToAddress(send.Sender)
		toChain, err = common.ParseChain(send.SenderChain)
		log.Info().Msgf("Abort: reverting inbound")
	} else if send.Status == types.SendStatus_PendingOutbound {
		to = ethcommon.HexToAddress(send.Receiver)
		toChain, err = common.ParseChain(send.ReceiverChain)
	}
	if err != nil {
		log.Error().Err(err).Msg("ParseChain fail; skip")
		return
	}

	// Early return if the send is already processed
	included, confirmed, _ := co.clientMap[toChain].IsSendOutTxProcessed(send.Index, int(send.Nonce))
	if included || confirmed {
		log.Info().Msgf("sendHash %s already processed; exit signer", send.Index)
		return
	}

	signer := co.signerMap[toChain]
	message, err := base64.StdEncoding.DecodeString(send.Message)
	if err != nil {
		log.Err(err).Msgf("decode send.Message %s error", send.Message)
	}

	gasLimit := send.GasLimit
	if gasLimit < 50_000 {
		gasLimit = 50_000
	}

	log.Info().Msgf("chain %s minting %d to %s, nonce %d, finalized %d", toChain, amount, to.Hex(), send.Nonce, send.FinalizedMetaHeight)
	sendHash, err := hex.DecodeString(send.Index[2:]) // remove the leading 0x
	if err != nil || len(sendHash) != 32 {
		log.Err(err).Msgf("decode sendHash %s error", send.Index)
		return
	}
	var sendhash [32]byte
	copy(sendhash[:32], sendHash[:32])
	gasprice, ok := new(big.Int).SetString(send.GasPrice, 10)
	if !ok {
		log.Err(err).Msgf("cannot convert gas price  %s ", send.GasPrice)
		return
	}
	var tx *ethtypes.Transaction

	signloopDone := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-confirmDone:
				return
			default:
				included, confirmed, err := co.clientMap[toChain].IsSendOutTxProcessed(send.Index, int(send.Nonce))
				if err != nil {
					numQueries++
				}
				if included || confirmed {
					log.Info().Msgf("sendHash %s included; kill this shepherd", send.Index)
					signloopDone <- true
					return
				}
				time.Sleep(12 * time.Second)
			}
		}
	}()

	// watch ZetaCore /zeta-chain/send/<sendHash> endpoint; send coreSendDone when the state of the send is updated;
	// e.g. pendingOutbound->outboundMined; or pendingOutbound->pendingRevert
	go func() {
		for {
			select {
			case <-coreSendDone:
				return
			default:
				newSend, err := co.bridge.GetSendByHash(send.Index)
				if err != nil || send == nil {
					log.Info().Msgf("sendHash %s cannot be found in ZetaCore; kill the shepherd", send.Index)
					signloopDone <- true
				}
				if newSend.Status != send.Status {
					log.Info().Msgf("sendHash %s status changed to %s from %s; kill the shepherd", send.Index, newSend.Status, send.Status)
					signloopDone <- true
				}
				time.Sleep(12 * time.Second)
			}
		}
	}()

	var signInterval int64 = 72 //  gap between two keysigns
	var startDelay int64 = 12   // maximum seconds delay before the first keysign
	startTimeUnix := (send.LastUpdateTimestamp + startDelay) / startDelay * startDelay
	//  if all zetaclients start within startDelay of each other,
	// and their clocks are roughly in sync, then they would arrive at the same time after sleep.
	time.Sleep(time.Unix(startTimeUnix, 0).Sub(time.Now()))
SIGNLOOP:
	for {
		select {
		case <-signloopDone:
			log.Info().Msg("breaking SignOutBoundTx loop: outbound already processed")
			break SIGNLOOP
		default:
			included, confirmed, _ := co.clientMap[toChain].IsSendOutTxProcessed(send.Index, int(send.Nonce))
			if included || confirmed {
				log.Info().Msgf("sendHash %s already confirmed; skip it", send.Index)
				break SIGNLOOP
			}
			timeSinceStart := time.Since(time.Unix(startTimeUnix, 0))
			srcChainID := config.Chains[send.SenderChain].ChainID
			if send.Status == types.SendStatus_PendingRevert {
				log.Info().Msgf("SignRevertTx: %s => %s, nonce %d, time since start %d", send.SenderChain, toChain, send.Nonce, send.Index, timeSinceStart)
				toChainID := config.Chains[send.ReceiverChain].ChainID
				tx, err = signer.SignRevertTx(ethcommon.HexToAddress(send.Sender), srcChainID, to.Bytes(), toChainID, amount, gasLimit, message, sendhash, send.Nonce, gasprice)
			} else if send.Status == types.SendStatus_PendingOutbound {
				log.Info().Msgf("SignOutboundTx: %s => %s, nonce %d, time since start %d", send.SenderChain, toChain, send.Nonce, send.Index, timeSinceStart)
				tx, err = signer.SignOutboundTx(ethcommon.HexToAddress(send.Sender), srcChainID, to, amount, gasLimit, message, sendhash, send.Nonce, gasprice)
			}
			if err != nil {
				log.Warn().Err(err).Msgf("SignOutboundTx error: nonce %d chain %s", send.Nonce, send.ReceiverChain)
				continue
			}
			cnt, err := co.GetPromCounter(OUTBOUND_TX_SIGN_COUNT)
			if err != nil {
				log.Error().Err(err).Msgf("GetPromCounter error")
			} else {
				cnt.Inc()
			}
			if tx != nil {
				outTxHash := tx.Hash().Hex()
				log.Info().Msgf("on chain %s nonce %d, sendHash: %s, outTxHash %s signer %s", signer.chain, send.Nonce, send.Index[:6], outTxHash, myid)
				if myid == send.Signers[send.Broadcaster] || myid == send.Signers[int(send.Broadcaster+1)%len(send.Signers)] {
					backOff := 1000 * time.Millisecond
					// retry loop: 1s, 2s, 4s, 8s, 16s in case of RPC error
					for i := 0; i < 5; i++ {
						log.Info().Msgf("broadcasting tx %s to chain %s: nonce %d, retry %d", outTxHash, toChain, send.Nonce, i)
						// #nosec G404 randomness is not a security issue here
						time.Sleep(time.Duration(rand.Intn(1500)) * time.Millisecond) //random delay to avoid sychronized broadcast
						err := signer.Broadcast(tx)
						if err != nil {
							retry := HandlerBroadcastError(err, co.fileLogger, strconv.FormatUint(send.Nonce, 10), toChain.String(), outTxHash)
							if !retry {
								break
							}
							backOff *= 2
							continue
						}
						log.Info().Msgf("Broadcast success: nonce %d chain %s outTxHash %s", send.Nonce, toChain, outTxHash)
						co.fileLogger.Info().Msgf("Broadcast success: nonce %d chain %s outTxHash %s", send.Nonce, toChain, outTxHash)
						zetaHash, err := co.bridge.AddTxHashToWatchlist(toChain.String(), tx.Nonce(), outTxHash)
						if err != nil {
							log.Err(err).Msgf("Unable to add to tracker on ZetaCore: nonce %d chain %s outTxHash %s", send.Nonce, toChain, outTxHash)
							break
						}
						log.Info().Msgf("Broadcast to core successful %s", zetaHash)
					}
				}
				co.fileLogger.Info().Msgf("Keysign: %s => %s, nonce %d, outTxHash %s; keysignCount %d", send.SenderChain, toChain, send.Nonce, outTxHash, keysignCount)
				keysignCount++
			}
		}

		// wake up at the next multiple 72s since startTimeUnit
		wakeTime := startTimeUnix + (time.Now().Unix()-startTimeUnix+signInterval)/signInterval*signInterval
		time.Sleep(time.Unix(wakeTime, 0).Sub(time.Now()))
	}
}

// return whether we should retry the broadcast
func HandlerBroadcastError(err error, logger *zerolog.Logger, nonce, toChain, outTxHash string) bool {
	if strings.Contains(err.Error(), "nonce too low") {
		log.Warn().Err(err).Msgf("nonce too low! this might be a unnecessary keysign. increase re-try interval and awaits outTx confirmation")
		logger.Err(err).Msgf("Broadcast nonce too low: nonce %d chain %s outTxHash %s; increase re-try interval", nonce, toChain, outTxHash)
		return false
	}
	if strings.Contains(err.Error(), "replacement transaction underpriced") {
		log.Warn().Err(err).Msgf("Broadcast replacement: nonce %d chain %s outTxHash %s", nonce, toChain, outTxHash)
		logger.Err(err).Msgf("Broadcast replacement: nonce %d chain %s outTxHash %s", nonce, toChain, outTxHash)
		return false
	} else if strings.Contains(err.Error(), "already known") { // this is error code from QuickNode
		log.Warn().Err(err).Msgf("Broadcast duplicates: nonce %d chain %s outTxHash %s", nonce, toChain, outTxHash)
		logger.Err(err).Msgf("Broadcast duplicates: nonce %d chain %s outTxHash %s", nonce, toChain, outTxHash)
		return false
	} // most likely an RPC error, such as timeout or being rate limited. Exp backoff retry

	log.Error().Err(err).Msgf("Broadcast error: nonce %d chain %s outTxHash %s; retring...", nonce, toChain, outTxHash)
	logger.Err(err).Msgf("Broadcast error: nonce %d chain %s outTxHash %s; retrying...", nonce, toChain, outTxHash)
	return true
}
