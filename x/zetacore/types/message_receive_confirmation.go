package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zeta-chain/zetacore/common"
)

var _ sdk.Msg = &MsgReceiveConfirmation{}

func NewMsgReceiveConfirmation(creator string, sendHash string, outTxHash string, outBlockHeight uint64, mMint string, status common.ReceiveStatus, chain string, nonce uint64) *MsgReceiveConfirmation {
	return &MsgReceiveConfirmation{
		Creator:        creator,
		SendHash:       sendHash,
		OutTxHash:      outTxHash,
		OutBlockHeight: outBlockHeight,
		MMint:          mMint,
		Status:         status,
		Chain:          chain,
		OutTxNonce:     nonce,
	}
}

func (msg *MsgReceiveConfirmation) Route() string {
	return RouterKey
}

func (msg *MsgReceiveConfirmation) Type() string {
	return "ReceiveConfirmation"
}

func (msg *MsgReceiveConfirmation) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgReceiveConfirmation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgReceiveConfirmation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}

func (msg *MsgReceiveConfirmation) Digest() string {
	m := *msg
	m.Creator = ""
	hash := crypto.Keccak256Hash([]byte(m.String()))
	return hash.Hex()
}