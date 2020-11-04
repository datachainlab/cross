package keeper

import (
	crosstypes "github.com/datachainlab/cross/x/core/types"
)

func makeContractTransactionID(txID crosstypes.TxID, txIndex crosstypes.TxIndex) []byte {
	size := len(txID)
	bz := make([]byte, size+1)
	copy(bz[:size], txID[:])
	bz[size] = txIndex
	return bz
}
