package types

import (
	"fmt"

	txtypes "github.com/datachainlab/cross/x/core/tx/types"
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

func KeyCoordinatorState(txID txtypes.TxID) []byte {
	return append(
		KeyPrefixBytes(KeyCoordinatorStatePrefix),
		txID[:]...,
	)
}

func KeyContractTransactionState(txID txtypes.TxID, txIndex txtypes.TxIndex) []byte {
	return append(
		append(
			KeyPrefixBytes(KeyContractTransactionStatePrefix),
			txID[:]...,
		),
		txIndex,
	)
}
