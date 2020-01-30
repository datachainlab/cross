package types

type OP interface {
	Equal(OP) bool
}

type OPs []OP

func (ops OPs) Equal(other OPs) bool {
	if len(ops) != len(other) {
		return false
	}
	for i, op := range ops {
		if !op.Equal(other[i]) {
			return false
		}
	}
	return true
}
