package types

import "fmt"

const (
	// ModuleName is the name of the module
	ModuleName = "cross"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the msg router key for the IBC module
	RouterKey string = ModuleName
)

const (
	KeyCoordinatorPrefix uint8 = iota + 1
	KeyTxPrefix
)

// KeyPrefixBytes return the key prefix bytes from a URL string format
func KeyPrefixBytes(prefix uint8) []byte {
	return []byte(fmt.Sprintf("%d/", prefix))
}

func KeyTx(txID TxID, txIndex TxIndex) []byte {
	return append(
		append(
			KeyPrefixBytes(KeyTxPrefix),
			txID[:]...,
		),
		txIndex,
	)
}

func KeyCoordinator(txID TxID) []byte {
	return append(
		KeyPrefixBytes(KeyCoordinatorPrefix),
		txID[:]...,
	)
}