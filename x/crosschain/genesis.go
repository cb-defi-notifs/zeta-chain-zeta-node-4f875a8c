package crosschain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zeta-chain/zetacore/x/crosschain/keeper"
	"github.com/zeta-chain/zetacore/x/crosschain/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the outTxTracker
	for _, elem := range genState.OutTxTrackerList {
		k.SetOutTxTracker(ctx, elem)
	}
	// Set all the inTxHashToCctx
	for _, elem := range genState.InTxHashToCctxList {
		k.SetInTxHashToCctx(ctx, elem)
	}
	// Set if defined
	if genState.PermissionFlags != nil {
		k.SetPermissionFlags(ctx, *genState.PermissionFlags)
	} else {
		k.SetPermissionFlags(ctx, types.PermissionFlags{IsInboundEnabled: true})
	}
	// this line is used by starport scaffolding # genesis/module/init
	// Set if defined
	if genState.Keygen != nil {
		k.SetKeygen(ctx, *genState.Keygen)
	}

	// Set all the gasPrice
	for _, elem := range genState.GasPriceList {
		k.SetGasPrice(ctx, *elem)
	}

	// Set all the chainNonces
	for _, elem := range genState.ChainNoncesList {
		k.SetChainNonces(ctx, *elem)
	}

	// Set all the lastBlockHeight
	for _, elem := range genState.LastBlockHeightList {
		k.SetLastBlockHeight(ctx, *elem)
	}

	// Set all the send
	for _, elem := range genState.CrossChainTxs {
		k.SetCrossChainTx(ctx, *elem)
	}

	// Set all the nodeAccount
	for _, elem := range genState.NodeAccountList {
		k.SetNodeAccount(ctx, *elem)
	}

	if genState.Tss != nil {
		k.SetTSS(ctx, *genState.Tss)
	}

}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.OutTxTrackerList = k.GetAllOutTxTracker(ctx)
	genesis.InTxHashToCctxList = k.GetAllInTxHashToCctx(ctx)
	// Get all permissionFlags
	permissionFlags, found := k.GetPermissionFlags(ctx)
	if found {
		genesis.PermissionFlags = &permissionFlags
	}
	// this line is used by starport scaffolding # genesis/module/export
	// Get all keygen
	keygen, found := k.GetKeygen(ctx)
	if found {
		genesis.Keygen = &keygen
	}

	// Get all tSSVoter
	// TODO : ADD for single TSS

	// Get all gasPrice
	gasPriceList := k.GetAllGasPrice(ctx)
	for _, elem := range gasPriceList {
		elem := elem
		genesis.GasPriceList = append(genesis.GasPriceList, &elem)
	}

	// Get all chainNonces
	chainNoncesList := k.GetAllChainNonces(ctx)
	for _, elem := range chainNoncesList {
		elem := elem
		genesis.ChainNoncesList = append(genesis.ChainNoncesList, &elem)
	}

	// Get all lastBlockHeight
	lastBlockHeightList := k.GetAllLastBlockHeight(ctx)
	for _, elem := range lastBlockHeightList {
		elem := elem
		genesis.LastBlockHeightList = append(genesis.LastBlockHeightList, &elem)
	}

	// Get all send
	sendList := k.GetAllCrossChainTx(ctx)
	for _, elem := range sendList {
		e := elem
		genesis.CrossChainTxs = append(genesis.CrossChainTxs, &e)
	}

	// Get all nodeAccount
	nodeAccountList := k.GetAllNodeAccount(ctx)
	for _, elem := range nodeAccountList {
		e := elem
		genesis.NodeAccountList = append(genesis.NodeAccountList, &e)
	}
	return genesis
}

// TODO : Verify genesis import and export
