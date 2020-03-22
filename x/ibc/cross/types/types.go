package types

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	TxID    = [32]byte
	TxIndex = uint8
)

const MaxContractTransactoinNum = math.MaxUint8

type ContractCallResult struct {
	ChainID  string           `json:"chain_id" yaml:"chain_id"`
	Height   int64            `json:"height" yaml:"height"`
	Signers  []sdk.AccAddress `json:"signers" yaml:"signers"`
	Contract []byte           `json:"contract" yaml:"contract"`
	OPs      []OP             `json:"ops" yaml:"ops"`
}

func (r ContractCallResult) String() string {
	// TODO make this more readable
	return fmt.Sprintf("%#v", r)
}
