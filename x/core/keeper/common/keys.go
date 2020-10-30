package common

import "github.com/datachainlab/cross/x/core/types"

func makeContractTransactionID(txID types.TxID, txIndex types.TxIndex) []byte {
	size := len(txID)
	bz := make([]byte, size+1)
	copy(bz[:size], txID[:])
	bz[size] = txIndex
	return bz
}
