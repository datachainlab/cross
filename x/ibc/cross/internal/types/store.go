package types

type State interface {
	Store
	Committer
}

type Store interface {
	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte

	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool

	// Set sets the key. Panics on nil key or value.
	Set(key, value []byte)

	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)
}

type Committer interface {
	Precommit(id []byte) error
	Commit(id []byte) error
	CommitImmediately() error
	Discard(id []byte) error
	OPs() OPs
}
