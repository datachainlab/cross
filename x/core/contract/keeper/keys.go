package keeper

import (
	txtypes "github.com/datachainlab/cross/x/core/tx/types"
	"github.com/datachainlab/cross/x/utils"
)

func makeContractTransactionID(txID txtypes.TxID, txIndex txtypes.TxIndex) []byte {
	size := len(txID)
	bz := make([]byte, size+4)
	copy(bz[:size], txID[:])
	copy(bz[size:], utils.Uint32ToBigEndian(txIndex))
	return bz
}
