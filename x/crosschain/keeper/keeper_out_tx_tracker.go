package keeper

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/zeta-chain/zetacore/x/crosschain/types"
	zetaObserverTypes "github.com/zeta-chain/zetacore/x/observer/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getOutTrackerIndex(chainID int64, nonce uint64) string {
	return fmt.Sprintf("%d-%d", chainID, nonce)
}

// SetOutTxTracker set a specific outTxTracker in the store from its index
func (k Keeper) SetOutTxTracker(ctx sdk.Context, outTxTracker types.OutTxTracker) {
	outTxTracker.Index = getOutTrackerIndex(outTxTracker.ChainId, outTxTracker.Nonce)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OutTxTrackerKeyPrefix))
	b := k.cdc.MustMarshal(&outTxTracker)
	store.Set(types.OutTxTrackerKey(
		outTxTracker.Index,
	), b)
}

// GetOutTxTracker returns a outTxTracker from its index
func (k Keeper) GetOutTxTracker(
	ctx sdk.Context,
	chainID int64,
	nonce uint64,

) (val types.OutTxTracker, found bool) {
	index := getOutTrackerIndex(chainID, nonce)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OutTxTrackerKeyPrefix))

	b := store.Get(types.OutTxTrackerKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveOutTxTracker removes a outTxTracker from the store
func (k Keeper) RemoveOutTxTracker(
	ctx sdk.Context,
	chainID int64,
	nonce uint64,

) {
	index := getOutTrackerIndex(chainID, nonce)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OutTxTrackerKeyPrefix))
	store.Delete(types.OutTxTrackerKey(
		index,
	))
}

// GetAllOutTxTracker returns all outTxTracker
func (k Keeper) GetAllOutTxTracker(ctx sdk.Context) (list []types.OutTxTracker) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OutTxTrackerKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.OutTxTracker
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// Queries

func (k Keeper) OutTxTrackerAll(c context.Context, req *types.QueryAllOutTxTrackerRequest) (*types.QueryAllOutTxTrackerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var outTxTrackers []types.OutTxTracker
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	outTxTrackerStore := prefix.NewStore(store, types.KeyPrefix(types.OutTxTrackerKeyPrefix))
	pageRes, err := query.Paginate(outTxTrackerStore, req.Pagination, func(key []byte, value []byte) error {
		var outTxTracker types.OutTxTracker
		if err := k.cdc.Unmarshal(value, &outTxTracker); err != nil {
			return err
		}

		outTxTrackers = append(outTxTrackers, outTxTracker)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAllOutTxTrackerResponse{OutTxTracker: outTxTrackers, Pagination: pageRes}, nil
}

func (k Keeper) OutTxTrackerAllByChain(c context.Context, req *types.QueryAllOutTxTrackerByChainRequest) (*types.QueryAllOutTxTrackerByChainResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var outTxTrackers []types.OutTxTracker
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	outTxTrackerStore := prefix.NewStore(store, types.KeyPrefix(types.OutTxTrackerKeyPrefix))

	pageRes, err := query.Paginate(outTxTrackerStore, req.Pagination, func(key []byte, value []byte) error {
		var outTxTracker types.OutTxTracker
		if err := k.cdc.Unmarshal(value, &outTxTracker); err != nil {
			return err
		}
		if outTxTracker.ChainId == req.Chain {
			outTxTrackers = append(outTxTrackers, outTxTracker)
		}
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllOutTxTrackerByChainResponse{OutTxTracker: outTxTrackers, Pagination: pageRes}, nil
}

func (k Keeper) OutTxTracker(c context.Context, req *types.QueryGetOutTxTrackerRequest) (*types.QueryGetOutTxTrackerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	val, found := k.GetOutTxTracker(
		ctx,
		req.ChainID,
		req.Nonce,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetOutTxTrackerResponse{OutTxTracker: val}, nil
}

// Messages

// The AddToOutTxTracker function adds a transaction hash and signer to the
// outgoing transaction tracker for a given chain and nonce. It first checks if
// the message creator is a bonded validator or an admin key. It then retrieves
// a chain based on the chain ID and checks if it is supported. After that, it
// retrieves a tracker based on the chain ID and the nonce. If the tracker does
// not exist, it creates a new tracker with the provided hash and signer. If the
// tracker exists, it checks if the hash already exists in the tracker. If not,
// it adds the hash and signer to the tracker. The function returns an empty
// MsgAddToOutTxTrackerResponse and no error.
func (k msgServer) AddToOutTxTracker(goCtx context.Context, msg *types.MsgAddToOutTxTracker) (*types.MsgAddToOutTxTrackerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validators := k.StakingKeeper.GetAllValidators(ctx)
	if !IsBondedValidator(msg.Creator, validators) && msg.Creator != types.AdminKey {
		return nil, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, fmt.Sprintf("signer %s is not a bonded validator", msg.Creator))
	}

	chain := k.zetaObserverKeeper.GetParams(ctx).GetChainFromChainID(msg.ChainId)
	if chain == nil {
		return nil, sdkerrors.Wrap(zetaObserverTypes.ErrSupportedChains, fmt.Sprintf("Chain is not supported %d", msg.ChainId))
	}

	tracker, found := k.GetOutTxTracker(ctx, msg.ChainId, msg.Nonce)
	hash := types.TxHashList{
		TxHash:   msg.TxHash,
		TxSigner: msg.Creator,
	}
	if !found {
		k.SetOutTxTracker(ctx, types.OutTxTracker{
			Index:    "",
			ChainId:  chain.ChainId,
			Nonce:    msg.Nonce,
			HashList: []*types.TxHashList{&hash},
		})
		return &types.MsgAddToOutTxTrackerResponse{}, nil
	}

	var isDup = false
	for _, hash := range tracker.HashList {
		if strings.EqualFold(hash.TxHash, msg.TxHash) {
			isDup = true
			break
		}
	}
	if !isDup {
		tracker.HashList = append(tracker.HashList, &hash)
		k.SetOutTxTracker(ctx, tracker)
	}
	return &types.MsgAddToOutTxTrackerResponse{}, nil
}

// The RemoveFromOutTxTracker function removes an outgoing transaction tracker
// for a given chain and nonce. It first checks if the message creator is a
// bonded validator or an admin key. It then removes the tracker with the
// provided chain ID and nonce using the RemoveOutTxTracker function. The
// function returns an empty MsgRemoveFromOutTxTrackerResponse and no error.
func (k msgServer) RemoveFromOutTxTracker(goCtx context.Context, msg *types.MsgRemoveFromOutTxTracker) (*types.MsgRemoveFromOutTxTrackerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validators := k.StakingKeeper.GetAllValidators(ctx)
	if !IsBondedValidator(msg.Creator, validators) && msg.Creator != types.AdminKey {
		return nil, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, fmt.Sprintf("signer %s is not a bonded validator", msg.Creator))
	}

	k.RemoveOutTxTracker(ctx, msg.ChainId, msg.Nonce)
	return &types.MsgRemoveFromOutTxTrackerResponse{}, nil
}
