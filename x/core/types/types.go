package types

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

type TxID = []byte

type TxIndex = uint8

type TxIndexSlice = []TxIndex

type AccountID []byte

func (id AccountID) AccAddress() sdk.AccAddress {
	return sdk.AccAddress(id)
}

// Account definition

// NewAccount creates a new instance of Account
func NewAccount(chainID ChainID, id AccountID) Account {
	var anyChainID *codectypes.Any
	if chainID != nil {
		var err error
		anyChainID, err = PackChainID(chainID)
		if err != nil {
			panic(err)
		}
	}
	return Account{ChainId: anyChainID, Id: id}
}

// NewLocalAccount creates a new instance of Account
func NewLocalAccount(id AccountID) Account {
	return NewAccount(nil, id)
}

// IsLocalAccount returns a boolean value whether the account is LocalAccount.
func (acc Account) IsLocalAccount() bool {
	return acc.ChainId == nil
}

// GetChainID returns ChainID
func (acc Account) GetChainID(m codec.Marshaler) ChainID {
	chainID, err := UnpackChainID(m, *acc.ChainId)
	if err != nil {
		panic(err)
	}
	return chainID
}

func (tx ContractTransaction) GetChainID(m codec.Marshaler) (ChainID, error) {
	var chainID ChainID
	if err := m.UnpackAny(tx.ChainId, &chainID); err != nil {
		return nil, err
	}
	return chainID, nil
}

func (tx ContractTransaction) ValidateBasic() error {
	return nil
}

func NewReturnValue(v []byte) *ReturnValue {
	rv := ReturnValue{Value: v}
	return &rv
}

func (rv *ReturnValue) IsNil() bool {
	if rv == nil {
		return true
	}
	return false
}

func (rv *ReturnValue) Equal(other *ReturnValue) bool {
	if rv.IsNil() && other.IsNil() {
		return true
	} else if rv.IsNil() && !other.IsNil() {
		return false
	} else if !rv.IsNil() && other.IsNil() {
		return false
	} else {
		return bytes.Equal(rv.Value, other.Value)
	}
}

func (lk Link) ValidateBasic() error {
	if lk.SrcIndex > math.MaxUint8 {
		return fmt.Errorf("src_index value is overflow: %v", lk.SrcIndex)
	}
	return nil
}

func (lk Link) GetSrcIndex() TxIndex {
	return TxIndex(lk.SrcIndex)
}

type ContractCallInfo []byte

type ContractRuntimeInfo struct {
	CommitMode             CommitMode
	StateConstraintType    StateConstraintType
	ExternalObjectResolver ObjectResolver
}

type StateConstraintType = uint32

const (
	NoStateConstraint         StateConstraintType = iota // NoStateConstraint indicates that no constraints on the state before and after the precommit is performed
	ExactMatchStateConstraint                            // ExactMatchStateConstraint indicates the constraint on state state before and after the precommit is performed
	PreStateConstraint                                   // PreStateConstraint indicates the constraint on state before the precommit is performed
	PostStateConstraint                                  // PostStateConstraint indicates the constraint on state after the precommit is performed
)

func NewStateConstraint(tp StateConstraintType, opItems []OP) StateConstraint {
	ops, err := PackOPs(opItems)
	if err != nil {
		panic(err)
	}
	return StateConstraint{Type: tp, Ops: ops}
}

type OP interface {
	proto.Message
}

// ChainID represents an ID of chain that contains a contract function to be called
type ChainID interface {
	proto.Message
	Type() string
	Equal(ChainID) bool
	String() string
}

var _ ChainID = (*ChannelInfo)(nil)

// Type implements ChainID.Type
func (ci ChannelInfo) Type() string {
	return "channelinfo"
}

func (ci *ChannelInfo) Equal(other ChainID) bool {
	return ci == other
}

// GetOurChainID returns our chainID
func GetOurChainID() ChainID {
	return &ChannelInfo{}
}

// ChainResolver defines the interface of resolver resolves chainID to ChannelInfo
type ChainResolver interface {
	Resolve(ctx sdk.Context, chainID ChainID) (*ChannelInfo, error)
	ConvertChainID(ctx sdk.Context, calleeID ChainID, callerID ChainID) (calleeOnCaller ChainID, err error)
	Capabilities() ChainResolverCapabilities
}

// ChainResolverCapabilities defines the capabilities for the ChainResolver
type ChainResolverCapabilities interface {
	// CrossChainCalls returns true if support for cross-chain calls is enabled.
	CrossChainCalls() bool
}

type chainResolverCapabilities struct {
	crossChainCalls bool
}

// CrossChainCalls implements ChainResolverCapabilities.CrossChainCalls
func (c chainResolverCapabilities) CrossChainCalls() bool {
	return c.crossChainCalls
}

// ChannelInfoResolver just returns a given ChannelInfo as is.
type ChannelInfoResolver struct {
	channelKeeper ChannelKeeper
}

// NewChannelInfoResolver creates a new instance of ChannelInfoResolver
func NewChannelInfoResolver(channelKeeper ChannelKeeper) ChannelInfoResolver {
	return ChannelInfoResolver{
		channelKeeper: channelKeeper,
	}
}

var _ ChainResolver = (*ChannelInfoResolver)(nil)

// Resolve implements ChainResolver.Resolve
func (r ChannelInfoResolver) Resolve(ctx sdk.Context, chainID ChainID) (*ChannelInfo, error) {
	ci, ok := chainID.(*ChannelInfo)
	if !ok {
		return nil, fmt.Errorf("cannot resolve '%v'", chainID)
	}
	return ci, nil
}

// ConvertChainID returns a chainID of callee in caller's context
func (r ChannelInfoResolver) ConvertChainID(ctx sdk.Context, callee ChainID, caller ChainID) (ChainID, error) {
	ours := GetOurChainID()
	ourChannelInfo, ok := ours.(*ChannelInfo)
	if !ok {
		return nil, errors.New("our chainID must be *ChannelInfo type")
	}
	calleeChannelInfo, ok := callee.(*ChannelInfo)
	if !ok {
		return nil, errors.New("callee's chainID must be *ChannelInfo type")
	}
	callerChannelInfo, ok := caller.(*ChannelInfo)
	if !ok {
		return nil, errors.New("caller's chainID must be *ChannelInfo type")
	}
	isLocalCallee := calleeChannelInfo == ourChannelInfo
	isLocalCaller := callerChannelInfo == ourChannelInfo

	if !isLocalCallee && !isLocalCaller {
		return nil, fmt.Errorf("either callee or caller must be our chain")
	} else if !isLocalCallee {
		return calleeChannelInfo, nil
	} else if !isLocalCaller {
		calleeChannel, found := r.channelKeeper.GetChannel(ctx, calleeChannelInfo.Channel, calleeChannelInfo.Port)
		if !found {
			return nil, fmt.Errorf("channel '%v' not found", calleeChannel.String())
		}
		return &ChannelInfo{Port: calleeChannel.GetCounterparty().GetPortID(), Channel: calleeChannel.GetCounterparty().GetChannelID()}, nil
	} else {
		return calleeChannelInfo, nil
	}
}

// Capabilities implements ChainResolver.Capabilities
func (r ChannelInfoResolver) Capabilities() ChainResolverCapabilities {
	return chainResolverCapabilities{crossChainCalls: false}
}

func NewContractTransactionInfo(tx ContractTransaction, linkObjects []Object) ContractTransactionInfo {
	anyObjects, err := PackObjects(linkObjects)
	if err != nil {
		panic(err)
	}
	return ContractTransactionInfo{
		Tx:      tx,
		Objects: anyObjects,
	}
}

func (ti ContractTransactionInfo) ValidateBasic() error {
	if err := ti.Tx.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

func (ti ContractTransactionInfo) UnpackObjects(m codec.Marshaler) []Object {
	objects, err := UnpackObjects(m, ti.Objects)
	if err != nil {
		panic(err)
	}
	return objects
}

// GetEvents converts Events to sdk.Events
func (res ContractCallResult) GetEvents() sdk.Events {
	events := make(sdk.Events, 0, len(res.Events))
	for _, ev := range res.Events {
		attrs := make([]sdk.Attribute, 0, len(ev.Attributes))
		for _, attr := range ev.Attributes {
			attrs = append(attrs, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
		}
		events = append(events, sdk.NewEvent(ev.Type, attrs...))
	}
	return events
}

// NewInitiateTxState creates an new instance of InitiateTxState
func NewInitiateTxState(remainingSigners []Account) InitiateTxState {
	var status InitiateTxStatus
	if len(remainingSigners) == 0 {
		status = INITIATE_TX_STATUS_VERIFIED
	} else {
		status = INITIATE_TX_STATUS_PENDING
	}
	return InitiateTxState{
		Status:           status,
		RemainingSigners: remainingSigners,
	}
}
