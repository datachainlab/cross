package types

type ContractCallInfo struct {
	ID     string
	Method string
	Args   [][]byte
}

func NewContractCallInfo(id, method string, args [][]byte) ContractCallInfo {
	return ContractCallInfo{
		ID:     id,
		Method: method,
		Args:   args,
	}
}

func (ci ContractCallInfo) Bytes() []byte {
	bz, err := EncodeContractSignature(ci)
	if err != nil {
		panic(err)
	}
	return bz
}

func EncodeContractSignature(c ContractCallInfo) ([]byte, error) {
	return ModuleCdc.MarshalBinaryLengthPrefixed(c)
}

func DecodeContractSignature(bz []byte) (*ContractCallInfo, error) {
	var c ContractCallInfo
	err := ModuleCdc.UnmarshalBinaryLengthPrefixed(bz, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
