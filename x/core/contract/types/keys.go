package types

import (
	"fmt"

	crosstypes "github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/utils"
)

const (
	ModuleName = "cross-contract"
)

const (
	KeyContractCallResultPrefix uint8 = iota
)

// KeyPrefixBytes return the key prefix bytes from a URL string format
func KeyPrefixBytes(prefix uint8) []byte {
	return []byte(fmt.Sprintf("%d/", prefix))
}

func KeyContractCallResult(txID crosstypes.TxID, txIndex crosstypes.TxIndex) []byte {
	return append(
		append(
			KeyPrefixBytes(KeyContractCallResultPrefix),
			txID[:]...,
		),
		utils.Uint32ToBigEndian(txIndex)...,
	)
}
