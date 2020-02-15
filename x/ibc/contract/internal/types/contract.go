package types

type ContractInfo struct {
	ID     string
	Method string
	Args   [][]byte
}

func NewContractInfo(id, method string, args [][]byte) ContractInfo {
	return ContractInfo{
		ID:     id,
		Method: method,
		Args:   args,
	}
}

func (ci ContractInfo) Bytes() []byte {
	bz, err := EncodeContractSignature(ci)
	if err != nil {
		panic(err)
	}
	return bz
}

func EncodeContractSignature(c ContractInfo) ([]byte, error) {
	return ModuleCdc.MarshalBinaryLengthPrefixed(c)
}

func DecodeContractSignature(bz []byte) (*ContractInfo, error) {
	var c ContractInfo
	err := ModuleCdc.UnmarshalBinaryLengthPrefixed(bz, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
