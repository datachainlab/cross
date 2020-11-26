package types

import "fmt"

const SubModuleName = "initiator"

const (
	KeyInitiateTxPrefix uint8 = iota
	KeyInitiateTxStatePrefix
)

// KeyPrefixBytes return the key prefix bytes from a URL string format
func KeyPrefixBytes(prefix uint8) []byte {
	return []byte(fmt.Sprintf("%d/", prefix))
}

func KeyInitiateTx() []byte {
	return KeyPrefixBytes(KeyInitiateTxPrefix)
}

func KeyInitiateTxState() []byte {
	return KeyPrefixBytes(KeyInitiateTxStatePrefix)
}
