package types

const (
	QueryCoordinatorStatus = "coordinator_status"
)

type QueryCoordinatorStatusRequest struct {
	TxID TxID `json:"tx_id" yaml:"tx_id"`
}

type QueryCoordinatorStatusResponse struct {
	TxID            TxID            `json:"tx_id" yaml:"tx_id"`
	CoordinatorInfo CoordinatorInfo `json:"coordinator_info" yaml:"coordinator_info"`
	Completed       bool            `json:"completed" yaml:"completed"`
}
