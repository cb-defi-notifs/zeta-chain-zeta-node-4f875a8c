package bitcoin

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/btcutil"
	"github.com/ethereum/go-ethereum"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/zeta-chain/zetacore/common"
	"github.com/zeta-chain/zetacore/zetaclient"
	"github.com/zeta-chain/zetacore/zetaclient/adapters/observer"
	"github.com/zeta-chain/zetacore/zetaclient/adapters/signer"
	"github.com/zeta-chain/zetacore/zetaclient/config"
	"github.com/zeta-chain/zetacore/zetaclient/model"
)

var _ observer.ChainObserver = (*BitcoinChainObserver)(nil)

type BitcoinChainObserver struct {
	chain                  common.Chain
	endpoint               string
	ticker                 *time.Ticker
	BtcClient              *rpcclient.Client
	zetaClient             *zetaclient.ZetaCoreBridge
	Tss                    signer.TSSSigner
	lastBlock              uint64
	confCount              uint64
	BlockTime              uint64
	txWatchList            map[btcutil.Tx]string
	mu                     *sync.Mutex
	db                     *leveldb.DB
	sampleLogger           *zerolog.Logger
	metrics                *metricsPkg.Metrics
	outTXConfirmedReceipts map[int]*wire.TxOut
	outTxChan              chan model.OutTx
	ZetaPriceQuerier       zetaclient.ZetaPriceQuerier
	stop                   chan struct{}
	fileLogger             *zerolog.Logger // for critical info
	logger                 zerolog.Logger
}

func NewBitcoinChainObserver() *BitcoinChainObserver {
	return &BitcoinChainObserver{}
}

func NewBitcoinChainObserver(chain common.Chain, bridge *zetaclient.ZetaCoreBridge, tss signer.TSSSigner, dbpath string, metrics *metricsPkg.Metrics) (*BitcoinChainObserver, error) {
	ob := BitcoinChainObserver{}
	ob.stop = make(chan struct{})
	ob.chain = chain
	ob.mu = &sync.Mutex{}
	sampled := log.Sample(&zerolog.BasicSampler{N: 10})
	ob.sampleLogger = &sampled
	ob.logger = log.With().Str("chain", chain.String()).Logger()
	ob.zetaClient = bridge
	ob.txWatchList = make(map[btcutil.Tx]string)
	ob.Tss = tss
	ob.metrics = metrics
	ob.outTXConfirmedReceipts = make(map[int]wire.TxOut)
	ob.outTxChan = make(chan model.OutTx, 100)
	ob.endpoint = config.Chains[chain.String()].Endpoint
	logFile, err := os.OpenFile(ob.chain.String()+"_debug.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		// Can we log an error before we have our logger? :)
		log.Error().Err(err).Msgf("there was an error creating a logFile chain %s", ob.chain.String())
	}
	fileLogger := zerolog.New(logFile).With().Logger()
	ob.fileLogger = &fileLogger

	// initialize the Client
	ob.logger.Info().Msgf("Chain %s endpoint %s", ob.chain, ob.endpoint)

	connConfig := &rpcclient.ConnConfig{
		Host:     ob.endpoint,
		Endpoint: "http",
	}
	client, err := rpcclient.New(connConfig, nil)
	if err != nil {
		ob.logger.Error().Err(err).Msg("btc client Dial")
		return nil, err
	}
	ob.BtcClient = client

	// create metric counters
	err = ob.RegisterPromCounter("rpc_getLogs_count", "Number of getLogs")
	if err != nil {
		return nil, err
	}
	err = ob.RegisterPromCounter("rpc_getBlockByNumber_count", "Number of getBlockByNumber")
	if err != nil {
		return nil, err
	}
	err = ob.RegisterPromGauge(metricsPkg.PendingTxs, "Number of pending transactions")
	if err != nil {
		return nil, err
	}

	// TODO : check from wich contract if we will get BTC prices from
	/*
		uniswapV3ABI, err := abi.JSON(strings.NewReader(config.UNISWAPV3POOL))
		if err != nil {
			return nil, err
		}
		uniswapV2ABI, err := abi.JSON(strings.NewReader(config.PANCAKEPOOL))
		if err != nil {
			return nil, err
		}

		uniswapv3Querier, uniswapv2Querier, dummyQuerior := ob.GetPriceQueriers(chain.String(), uniswapV3ABI, uniswapV2ABI)
		ob.SetChainDetails(chain, uniswapv3Querier, uniswapv2Querier)
		if os.Getenv("DUMMY_PRICE") != "" {
	*/
	ob.logger.Info().Msg("Using dummy price of 1:1")
	ob.ZetaPriceQuerier = dummyQuerior
	/*} */

	if dbpath != "" {
		err := ob.BuildBlockIndex(dbpath, chain.String())
		if err != nil {
			return nil, err
		}
		ob.BuildReceiptsMap()

	}
	ob.logger.Info().Msgf("%s: start scanning from block %d", chain, ob.GetLastBlock())

	return &ob, nil
}

func (ob *BtcChainObserver) Start() {
	go ob.ExternalChainWatcher() // Observes external Chains for incoming trasnactions
	go ob.WatchGasPrice()        // Observes external Chains for Gas prices and posts to core
	go ob.WatchExchangeRate()    // Observers ZetaPriceQuerier for Zeta prices and posts to core
	go ob.observeOutTx()
}

func (ob *BtcChainObserver) Stop() {
	ob.logger.Info().Msgf("ob %s is stopping", ob.chain)
	close(ob.stop) // this notifies all goroutines to stop

	ob.logger.Info().Msg("closing ob.db")
	err := ob.db.Close()
	if err != nil {
		ob.logger.Error().Err(err).Msg("error closing db")
	}

	ob.logger.Info().Msgf("%s observer stopped", ob.chain)
}

// returns: isIncluded, isConfirmed, Error
// If isConfirmed, it also post to ZetaCore
func (ob *BtcChainObserver) IsSendOutTxProcessed(sendHash string, nonce int) (bool, bool, error) {
	ob.mu.Lock()
	receipt, found := ob.outTXConfirmedReceipts[nonce]
	ob.mu.Unlock()
	sendID := fmt.Sprintf("%s/%d", ob.chain.String(), nonce)
	logger := ob.logger.With().Str("sendID", sendID).Logger()
	if found && receipt.Status == 1 {
		logs := receipt.Logs
		for _, vLog := range logs {
			receivedLog, err := ob.Connector.ConnectorFilterer.ParseZetaReceived(*vLog)
			if err == nil {
				logger.Info().Msgf("Found (outTx) sendHash %s on chain %s txhash %s", sendHash, ob.chain, vLog.TxHash.Hex())
				if vLog.BlockNumber+ob.confCount < ob.GetLastBlock() {
					logger.Info().Msg("Confirmed! Sending PostConfirmation to zetacore...")
					if len(vLog.Topics) != 4 {
						logger.Error().Msgf("wrong number of topics in log %d", len(vLog.Topics))
						return false, false, fmt.Errorf("wrong number of topics in log %d", len(vLog.Topics))
					}
					sendhash := vLog.Topics[3].Hex()
					//var rxAddress string = ethcommon.HexToAddress(vLog.Topics[1].Hex()).Hex()
					mMint := receivedLog.ZetaValue.String()
					zetaHash, err := ob.zetaClient.PostReceiveConfirmation(
						sendhash,
						vLog.TxHash.Hex(),
						vLog.BlockNumber,
						mMint,
						common.ReceiveStatus_Success,
						ob.chain.String(),
						nonce,
					)
					if err != nil {
						logger.Error().Err(err).Msg("error posting confirmation to meta core")
						continue
					}
					logger.Info().Msgf("Zeta tx hash: %s\n", zetaHash)
					return true, true, nil
				}
				logger.Info().Msgf("Included; %d blocks before confirmed! chain %s nonce %d", int(vLog.BlockNumber+ob.confCount)-int(ob.GetLastBlock()), ob.chain, nonce)
				return true, false, nil
			}
			revertedLog, err := ob.Connector.ConnectorFilterer.ParseZetaReverted(*vLog)
			if err == nil {
				logger.Info().Msgf("Found (revertTx) sendHash %s on chain %s txhash %s", sendHash, ob.chain, vLog.TxHash.Hex())
				if vLog.BlockNumber+ob.confCount < ob.GetLastBlock() {
					logger.Info().Msg("Confirmed! Sending PostConfirmation to zetacore...")
					if len(vLog.Topics) != 3 {
						logger.Error().Msgf("wrong number of topics in log %d", len(vLog.Topics))
						return false, false, fmt.Errorf("wrong number of topics in log %d", len(vLog.Topics))
					}
					sendhash := vLog.Topics[2].Hex()
					mMint := revertedLog.RemainingZetaValue.String()
					metaHash, err := ob.zetaClient.PostReceiveConfirmation(
						sendhash,
						vLog.TxHash.Hex(),
						vLog.BlockNumber,
						mMint,
						common.ReceiveStatus_Success,
						ob.chain.String(),
						nonce,
					)
					if err != nil {
						logger.Err(err).Msg("error posting confirmation to meta core")
						continue
					}
					logger.Info().Msgf("Zeta tx hash: %s", metaHash)
					return true, true, nil
				}
				logger.Info().Msgf("Included; %d blocks before confirmed! chain %s nonce %d", int(vLog.BlockNumber+ob.confCount)-int(ob.GetLastBlock()), ob.chain, nonce)
				return true, false, nil
			}
		}
	} else if found && receipt.Status == 0 {
		//FIXME: check nonce here by getTransaction RPC
		logger.Info().Msgf("Found (failed tx) sendHash %s on chain %s txhash %s", sendHash, ob.chain, receipt.TxHash.Hex())
		zetaTxHash, err := ob.zetaClient.PostReceiveConfirmation(sendHash, receipt.TxHash.Hex(), receipt.BlockNumber.Uint64(), "", common.ReceiveStatus_Failed, ob.chain.String(), nonce)
		if err != nil {
			logger.Error().Err(err).Msgf("PostReceiveConfirmation error in WatchTxHashWithTimeout; zeta tx hash %s", zetaTxHash)
		}
		logger.Info().Msgf("Zeta tx hash: %s", zetaTxHash)
		return true, true, nil
	}

	return false, false, nil
}

// FIXME: there's a chance that a txhash in OutTxChan may not deliver when Stop() is called
// observeOutTx periodically checks all the txhash in potential outbound txs
func (ob *BtcChainObserver) observeOutTx() {
	logger := ob.logger
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			trackers, err := ob.zetaClient.GetAllOutTxTrackerByChain(ob.chain)
			if err != nil {
				return
			}
			outTimeout := time.After(90 * time.Second)
		TRACKERLOOP:
			for _, tracker := range trackers {
				nonceInt, err := strconv.Atoi(tracker.Nonce)
				if err != nil {
					return
				}
			TXHASHLOOP:
				for _, txHash := range tracker.HashList {
					inTimeout := time.After(1000 * time.Millisecond)
					select {
					case <-outTimeout:
						logger.Warn().Msgf("observeOutTx timeout on nonce %d", nonceInt)
						break TRACKERLOOP
					default:
						receipt, err := ob.queryTxByHash(txHash.TxHash, int64(nonceInt))
						if err == nil && receipt != nil { // confirmed
							ob.mu.Lock()
							ob.outTXConfirmedReceipts[nonceInt] = receipt
							value, err := receipt.MarshalJSON()
							if err != nil {
								logger.Error().Err(err).Msgf("receipt marshal error %s", receipt.TxHash.Hex())
							}
							ob.mu.Unlock()
							err = ob.db.Put([]byte(model.NonceTxKeyPrefix+fmt.Sprintf("%d", nonceInt)), value, nil)
							if err != nil {
								logger.Error().Err(err).Msgf("PurgeTxHashWatchList: error putting nonce %d tx hashes %s to db", nonceInt, receipt.TxHash.Hex())
							}
							break TXHASHLOOP
						}
						<-inTimeout
					}
				}
			}
		case <-ob.stop:
			logger.Info().Msg("observeOutTx: stopped")
			return
		}
	}
}

// return the status of txHash
// receipt nil, err non-nil: txHash not found
// receipt nil, err nil: txHash receipt recorded, but may not be confirmed
// receipt non-nil, err nil: txHash confirmed
func (ob *BtcChainObserver) queryTxByHash(txHash string, nonce int64) (*ethtypes.Receipt, error) {
	logger := ob.logger.With().Str("txHash", txHash).Int64("nonce", nonce).Logger()
	if ob.outTXConfirmedReceipts[int(nonce)] != nil {
		return nil, fmt.Errorf("queryTxByHash: txHash %s receipts already recorded", txHash)
	}
	ctxt, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	receipt, err := ob.EvmClient.TransactionReceipt(ctxt, ethcommon.HexToHash(txHash))
	if err != nil {
		if err != ethereum.NotFound {
			logger.Warn().Err(err).Msg("TransactionReceipt")
		}
		return nil, err
	} else if receipt.BlockNumber.Uint64()+ob.confCount > ob.GetLastBlock() {
		log.Warn().Msgf("included but not confirmed: receipt block %d, current block %d", receipt.BlockNumber, ob.GetLastBlock())
		return nil, fmt.Errorf("included but not confirmed")
	} else { // confirmed outbound tx
		return receipt, nil
	}
}

func (ob *BtcChainObserver) setLastBlock(block uint64) {
	atomic.StoreUint64(&ob.lastBlock, block)
}

func (ob *BtcChainObserver) GetLastBlock() uint64 {
	return atomic.LoadUint64(&ob.lastBlock)
}

func (ob *BtcChainObserver) BlockTimeSeconds() uint64 {
	return ob.BlockTime
}

func (ob *BtcChainObserver) Chain() *common.Chain {
	return &ob.chain
}

func (ob *BtcChainObserver) ChainClient() *rpcclient.Client {
	return ob.BtcClient
}

func (ob *BtcChainObserver) ConfirmationsCount() uint64 {
	return ob.confCount
}

func (ob *BtcChainObserver) CriticalLog() *zerolog.Logger {
	return ob.fileLogger
}

func (ob *BtcChainObserver) Log() zerolog.Logger {
	return ob.logger
}

func (ob *BtcChainObserver) Endpoint() string {
	return ob.endpoint
}

func (ob *BtcChainObserver) LastBlock() uint64 {
	return ob.lastBlock
}

func (ob *BtcChainObserver) OutTxChan() chan model.OutTx {
	return ob.outTxChan
}

func (ob *BtcChainObserver) TSSSigner() signer.TSSSigner {
	return ob.Tss
}

func (ob *BtcChainObserver) Ticker() *time.Ticker {
	return ob.ticker
}

func (ob *BtcChainObserver) TxWatchList() map[string]string {
	watchList := make(map[string]string, len(ob.txWatchList))
	for k, v := range ob.txWatchList {
		watchList[k.Hex()] = v
	}
	return watchList
}
