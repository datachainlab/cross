package types

import (
	"fmt"

	crosstypes "github.com/datachainlab/cross/x/core/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "cross-atomic"
)

const (
	KeyCoordinatorStatePrefix uint8 = iota
	KeyContractTransactionStatePrefix
	KeyContractCallResultPrefix
)

// KeyPrefixBytes return the key prefix bytes from a URL string format
func KeyPrefixBytes(prefix uint8) []byte {
	return []byte(fmt.Sprintf("%d/", prefix))
}

func KeyCoordinatorState(txID crosstypes.TxID) []byte {
	return append(
		KeyPrefixBytes(KeyCoordinatorStatePrefix),
		txID[:]...,
	)
}

func KeyContractTransactionState(txID crosstypes.TxID, txIndex crosstypes.TxIndex) []byte {
	return append(
		append(
			KeyPrefixBytes(KeyContractTransactionStatePrefix),
			txID[:]...,
		),
		txIndex,
	)
}

func KeyContractCallResult(txID crosstypes.TxID, txIndex crosstypes.TxIndex) []byte {
	return append(
		append(
			KeyPrefixBytes(KeyContractCallResultPrefix),
			txID[:]...,
		),
		txIndex,
	)
}
