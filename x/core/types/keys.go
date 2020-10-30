package types

import "fmt"

const (
	// ModuleName defines the module name
	ModuleName = "cross"

	// Version defines the current version the Cross
	// module supports
	Version = "cross-1"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_capability"

	// PortID defines the portID of this module
	PortID = "cross"
)

const (
	CommitProtocolSimple uint32 = iota // Simple commit
	CommitProtocolTPC                  // Two-phase commit
)

const (
	KeyCoordinatorPrefix uint8 = iota
	KeyTxPrefix
	KeyContractResultPrefix
	KeyUnacknowledgedPacketPrefix
)

// KeyPrefixBytes return the key prefix bytes from a URL string format
func KeyPrefixBytes(prefix uint8) []byte {
	return []byte(fmt.Sprintf("%d/", prefix))
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
