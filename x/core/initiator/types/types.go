package types

import (
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/codec"
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	xcctypes "github.com/datachainlab/cross/x/core/xcc/types"
)

func (tx ContractTransaction) GetCrossChainChannel(m codec.Codec) (xcctypes.XCC, error) {
	var xcc xcctypes.XCC
	if err := m.UnpackAny(tx.CrossChainChannel, &xcc); err != nil {
		return nil, err
	}
	return xcc, nil
}

func (tx ContractTransaction) ValidateBasic() error {
	return nil
}

func (lk Link) ValidateBasic() error {
	if lk.SrcIndex > math.MaxUint8 {
		return fmt.Errorf("src_index value is overflow: %v", lk.SrcIndex)
	}
	return nil
}

func (lk Link) GetSrcIndex() txtypes.TxIndex {
	return txtypes.TxIndex(lk.SrcIndex)
}

// NewInitiateTxState creates an new instance of InitiateTxState
func NewInitiateTxState(msg MsgInitiateTx) InitiateTxState {
	return InitiateTxState{
		Status: INITIATE_TX_STATUS_PENDING,
		Msg:    msg,
	}
}

func (state InitiateTxState) IsVerified() bool {
	return state.Status == INITIATE_TX_STATUS_VERIFIED
}
