package types

import (
	"fmt"

	crosstypes "github.com/datachainlab/cross/x/core/types"
)

const (
	KeyCoordinatorStatePrefix uint8 = iota
	KeyContractTransactionStatePrefix
	KeyContractResultPrefix
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

func KeyContractResult(txID crosstypes.TxID, txIndex crosstypes.TxIndex) []byte {
	return append(
		append(
			KeyPrefixBytes(KeyContractResultPrefix),
			txID[:]...,
		),
		txIndex,
	)
}