package types_test

import (
	"github.com/stretchr/testify/require"
	"github.com/zeta-chain/zetacore/x/crosschain/types"
	"testing"
)

// FIXME: make it work
func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{

				//ZetaConversionRateList: []types.ZetaConversionRate{
				//	{
				//		Index: "0",
				//	},
				//	{
				//		Index: "1",
				//	},
				//},
				OutTxTrackerList: []types.OutTxTracker{
					{
						Index: "0",
					},
					{
						Index: "1",
					},
				},
				InTxHashToCctxList: []types.InTxHashToCctx{
					{
						InTxHash: "0",
					},
					{
						InTxHash: "1",
					},
				},
				PermissionFlags: &types.PermissionFlags{
					IsInboundEnabled: true,
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated outTxTracker",
			genState: &types.GenesisState{
				OutTxTrackerList: []types.OutTxTracker{
					{
						Index: "0",
					},
					{
						Index: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated inTxHashToCctx",
			genState: &types.GenesisState{
				InTxHashToCctxList: []types.InTxHashToCctx{
					{
						InTxHash: "0",
					},
					{
						InTxHash: "0",
					},
				},
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
