package keeper

import (
	crosstypes "github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/utils"
)

func makeContractTransactionID(txID crosstypes.TxID, txIndex crosstypes.TxIndex) []byte {
	size := len(txID)
	bz := make([]byte, size+4)
	copy(bz[:size], txID[:])
	copy(bz[size:], utils.Uint32ToBigEndian(txIndex))
	return bz
}
