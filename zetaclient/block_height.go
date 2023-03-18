package zetaclient

import (
	"context"
	//"github.com/Meta-Protocol/zetacore/common"
	"github.com/zeta-chain/zetacore/x/crosschain/types"
)

// GetBlockHeight returns the current height for metachain blocks
// FIXME: deprecate this in favor of tendermint RPC?
func (b *ZetaCoreBridge) GetBlockHeight() (uint64, error) {
	client := types.NewQueryClient(b.grpcConn)
	height, err := client.LastMetaHeight(
		context.Background(),
		&types.QueryLastMetaHeightRequest{},
	)
	if err != nil {
		return 0, err
	}

	//fmt.Printf("block height: %d\n", height.Height)
	return height.Height, nil
}

//func (b *ZetaCoreBridge) GetLastBlockObserved(chain common.Chain) (uint64, error) {
//	Client := types.NewQueryClient(b.grpcConn)
//	last_obs, err := Client.LastBlockObserved(
//		context.Background(),
//		&types.QueryGetLastBlockObservedRequest{
//			Index: chain.String(),
//		},
//	)
//	if err != nil {
//		return 0, err
//	}
//
//	observed := last_obs.LastBlockObserved
//	fmt.Printf("last observed block height on chain %s: %d\n",
//		observed.Chain,
//		observed.Height)
//	return observed.Height, nil
//}
