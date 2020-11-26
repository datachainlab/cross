package keeper

import txtypes "github.com/datachainlab/cross/x/core/tx/types"

func makeContractTransactionID(txID txtypes.TxID, txIndex txtypes.TxIndex) []byte {
	size := len(txID)
	bz := make([]byte, size+1)
	copy(bz[:size], txID[:])
	bz[size] = txIndex
	return bz
}
