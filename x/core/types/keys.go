package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
	// MemStoreKey = "mem_capability"

	// PortID defines the portID of this module
	PortID = "cross"
)

var (
	InitiatorKeyPrefix     = []byte("initiator")
	AuthKeyPrefix          = []byte("auth")
	AtomicKeyPrefix        = []byte("atomic")
	ContractManagerPrefix  = []byte("cmanager")
	ContractStoreKeyPrefix = []byte("cstore")
)

type PrefixStoreKey struct {
	StoreKey sdk.StoreKey
	Prefix   []byte
}

var _ sdk.StoreKey = (*PrefixStoreKey)(nil)

func NewPrefixStoreKey(storeKey sdk.StoreKey, prefix []byte) *PrefixStoreKey {
	return &PrefixStoreKey{
		StoreKey: storeKey,
		Prefix:   prefix,
	}
}

func (sk *PrefixStoreKey) Name() string {
	return sk.StoreKey.Name()
}

func (sk *PrefixStoreKey) String() string {
	return fmt.Sprintf("PrefixStoreKey{%p, %s, %v}", sk, sk.Name(), sk.Prefix)
}
