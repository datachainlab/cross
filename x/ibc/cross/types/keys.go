package types

import "fmt"

const (
	// ModuleName is the name of the module
	ModuleName = "cross"

	// Version defines the current version the Cross
	// module supports
	Version = "cross-1"

	// PortID that Cross module binds to
	PortID = "cross"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the msg router key for the IBC module
	RouterKey string = ModuleName

	// QuerierRoute is the querier route for Cross
	QuerierRoute = ModuleName
)

const (
	TypeInitiate      = "cross_initiate"
	TypePrepare       = "cross_prepare"
	TypePrepareResult = "cross_prepare_result"
	TypeCommit        = "cross_commit"
	TypeAckCommit     = "cross_ack_commit"
)

const (
	KeyCoordinatorPrefix uint8 = iota
	KeyTxPrefix
	KeyContractResultPrefix
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

func KeyContractResult(txID TxID, txIndex TxIndex) []byte {
	return append(
		append(
			KeyPrefixBytes(KeyContractResultPrefix),
			txID[:]...,
		),
		txIndex,
	)
}
