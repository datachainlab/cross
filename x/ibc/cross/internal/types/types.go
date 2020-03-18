package types

import "math"

type (
	TxID    = [32]byte
	TxIndex = uint8
)

const MaxContractTransactoinNum = math.MaxUint8
