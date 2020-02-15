package types

const (
	// ModuleName is the name of the module
	ModuleName = "contract"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the msg router key for the IBC module
	RouterKey string = ModuleName
)
