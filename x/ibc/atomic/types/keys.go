package types

const (
	// ModuleName defines the module name
	ModuleName = "atomic"

	// Version defines the current version the Atomic
	// module supports
	Version = "atomic-1"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_capability"

	// PortID defines the portID of this module
	PortID = "atomic"
)
