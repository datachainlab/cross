package types

import "fmt"

const (
	KeyTxAuthStatePrefix uint8 = iota
)

// KeyPrefixBytes return the key prefix bytes from a URL string format
func KeyPrefixBytes(prefix uint8) []byte {
	return []byte(fmt.Sprintf("%d/", prefix))
}

func KeyTxAuthState() []byte {
	return KeyPrefixBytes(KeyTxAuthStatePrefix)
}
