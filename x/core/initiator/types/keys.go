package types

import "fmt"

const SubModuleName = "initiator"

const (
	KeyInitiateTxStatePrefix uint8 = iota
)

// KeyPrefixBytes return the key prefix bytes from a URL string format
func KeyPrefixBytes(prefix uint8) []byte {
	return []byte(fmt.Sprintf("%d/", prefix))
}

func KeyInitiateTxState() []byte {
	return KeyPrefixBytes(KeyInitiateTxStatePrefix)
}
