package types

import "fmt"

const (
	// ModuleName is the name of the module
	ModuleName = "crossccc"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the msg router key for the IBC module
	RouterKey string = ModuleName
)

const (
	KeyTxPrefix int = iota + 1
	KeyCoordinatorPrefix
)

// KeyPrefixBytes return the key prefix bytes from a URL string format
func KeyPrefixBytes(prefix int) []byte {
	return []byte(fmt.Sprintf("%d/", prefix))
}

func KeyTx(txID []byte) []byte {
	return append(
		KeyPrefixBytes(KeyTxPrefix),
		txID...,
	)
}

func KeyCoordinator(txID []byte) []byte {
	return append(
		KeyPrefixBytes(KeyCoordinatorPrefix),
		txID...,
	)
}
