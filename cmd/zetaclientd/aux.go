package main

import (
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/zeta-chain/zetacore/common"
	"github.com/zeta-chain/zetacore/common/cosmos"
	"github.com/zeta-chain/zetacore/zetaclient"
	"github.com/zeta-chain/zetacore/zetaclient/config"
	"github.com/zeta-chain/zetacore/zetaclient/metrics"
)

func CreateAuthzSigner(granter string, grantee sdk.AccAddress) {
	zetaclient.SetupAuthZSignerList(granter, grantee)
}

func CreateZetaBridge(chainHomeFolder string, config *config.Config) (*zetaclient.ZetaCoreBridge, error) {
	signerName := config.ValidatorName
	signerPass := "password"
	chainIP := config.ZetaCoreURL
	kb, err := zetaclient.GetKeyringKeybase([]common.KeyType{common.ZetaClientGranteeKey, common.TssSignerKey}, chainHomeFolder, signerName, signerPass)
	if err != nil {
		return nil, err
	}
	granterAddreess, err := cosmos.AccAddressFromBech32(config.AuthzGranter)
	if err != nil {
		return nil, err
	}
	k := zetaclient.NewKeysWithKeybase(kb, granterAddreess, signerName, signerPass)
	authzSignerName := zetaclient.GetGranteeKeyName(common.ZetaClientGranteeKey, signerName)
	bridge, err := zetaclient.NewZetaCoreBridge(k, chainIP, authzSignerName)
	if err != nil {
		return nil, err
	}
	return bridge, nil
}

func CreateSignerMap(tss zetaclient.TSSSigner) (map[common.Chain]zetaclient.ChainSigner, error) {
	signerMap := make(map[common.Chain]zetaclient.ChainSigner)
	for _, chain := range config.ChainsEnabled {
		if chain.IsEVMChain() {
			mpiAddress := ethcommon.HexToAddress(config.ChainConfigs[chain.ChainName.String()].ConnectorContractAddress)
			erc20CustodyAddress := ethcommon.HexToAddress(config.ChainConfigs[chain.ChainName.String()].ERC20CustodyContractAddress)
			signer, err := zetaclient.NewEVMSigner(chain, config.ChainConfigs[chain.ChainName.String()].Endpoint, tss, config.ConnectorAbiString, config.ERC20CustodyAbiString, mpiAddress, erc20CustodyAddress)
			if err != nil {
				log.Fatal().Err(err).Msgf("%s: NewEVMSigner Ethereum error ", chain.String())
				return nil, err
			}
			signerMap[chain] = signer
		} else if chain.IsBitcoinChain() {
			// FIXME: move the construction of rpcclient to somewhere else
			connCfg := &rpcclient.ConnConfig{
				Host:         config.BitcoinConfig.RPCEndpoint,
				User:         config.BitcoinConfig.RPCUsername,
				Pass:         config.BitcoinConfig.RPCPassword,
				HTTPPostMode: true,
				DisableTLS:   true,
				Params:       config.BitcoinConfig.RPCParams,
			}
			client, err := rpcclient.New(connCfg, nil)
			if err != nil {
				return nil, fmt.Errorf("error creating rpc client: %s", err)
			}
			signer, err := zetaclient.NewBTCSigner(tss, client)
			if err != nil {
				log.Fatal().Err(err).Msgf("%s: NewBitcoinSigner Bitcoin error ", chain.String())
				return nil, err
			}
			signerMap[chain] = signer
		}
	}

	return signerMap, nil
}

func CreateChainClientMap(bridge *zetaclient.ZetaCoreBridge, tss zetaclient.TSSSigner, dbpath string, metrics *metrics.Metrics) (map[common.Chain]zetaclient.ChainClient, error) {
	clientMap := make(map[common.Chain]zetaclient.ChainClient)
	for _, chain := range config.ChainsEnabled {
		log.Info().Msgf("starting observer for : %s ", chain.String())
		var co zetaclient.ChainClient
		var err error
		if chain.IsEVMChain() {
			co, err = zetaclient.NewEVMChainClient(chain, bridge, tss, dbpath, metrics)
		} else {
			co, err = zetaclient.NewBitcoinClient(chain, bridge, tss, dbpath, metrics)
		}
		if err != nil {
			log.Err(err).Msgf("%s NewEVMChainClient", chain.String())
			return nil, err
		}
		clientMap[chain] = co
	}

	return clientMap, nil
}
