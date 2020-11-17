package types

import (
	"bytes"
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
func NewAccount(xcc CrossChainChannel, id AccountID) Account {
	var anyCrossChainChannel *codectypes.Any
	if xcc != nil {
		var err error
		anyCrossChainChannel, err = PackCrossChainChannel(xcc)
		if err != nil {
			panic(err)
		}
	}
	return Account{CrossChainChannel: anyCrossChainChannel, Id: id}
}

// GetCrossChainChannel returns CrossChainChannel
func (acc Account) GetCrossChainChannel(m codec.Marshaler) CrossChainChannel {
	xcc, err := UnpackCrossChainChannel(m, *acc.CrossChainChannel)
	if err != nil {
		panic(err)
	}
	return xcc
}

func (tx ContractTransaction) GetCrossChainChannel(m codec.Marshaler) (CrossChainChannel, error) {
	var xcc CrossChainChannel
	if err := m.UnpackAny(tx.CrossChainChannel, &xcc); err != nil {
		return nil, err
	}
	return xcc, nil
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

// CrossChainChannel represents a channel of chain that contains a contract function to be called
type CrossChainChannel interface {
	proto.Message
	Type() string
	Equal(CrossChainChannel) bool
	String() string
}

var _ CrossChainChannel = (*ChannelInfo)(nil)

// Type implements CrossChainChannel.Type
func (ci ChannelInfo) Type() string {
	return "channelinfo"
}

func (ci *ChannelInfo) Equal(other CrossChainChannel) bool {
	oi, ok := other.(*ChannelInfo)
	if !ok {
		return false
	}
	return ci.Port == oi.Port && ci.Channel == oi.Channel
}

// CrossChainChannelResolver defines the interface of resolver resolves cross-chain channel to ChannelInfo
type CrossChainChannelResolver interface {
	ResolveCrossChainChannel(ctx sdk.Context, xcc CrossChainChannel) (*ChannelInfo, error)
	ResolveChannel(ctx sdk.Context, channel *ChannelInfo) (CrossChainChannel, error)
	ConvertCrossChainChannel(ctx sdk.Context, calleeXCC CrossChainChannel, callerXCC CrossChainChannel) (calleeXCCOnCaller CrossChainChannel, err error)
	GetSelfCrossChainChannel(ctx sdk.Context) CrossChainChannel
	IsSelfCrossChainChannel(ctx sdk.Context, xcc CrossChainChannel) bool
	Capabilities() CrossChainChannelResolverCapabilities
}

// CrossChainChannelResolverCapabilities defines the capabilities for the ChainResolver
type CrossChainChannelResolverCapabilities interface {
	// CrossChainCalls returns true if support for cross-chain calls is enabled.
	CrossChainCalls(ctx sdk.Context, commitProtocol CommitProtocol) bool
}

type crossChainChannelResolverCapabilities struct {
	crossChainCalls map[CommitProtocol]bool
}

// CrossChainCalls implements ChainResolverCapabilities.CrossChainCalls
func (c crossChainChannelResolverCapabilities) CrossChainCalls(ctx sdk.Context, cp CommitProtocol) bool {
	return c.crossChainCalls[cp]
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

var _ CrossChainChannelResolver = (*ChannelInfoResolver)(nil)

// ResolveCrossChainChannel implements CrossChainChannelResolver.ResolveCrossChainChannel
func (r ChannelInfoResolver) ResolveCrossChainChannel(ctx sdk.Context, xcc CrossChainChannel) (*ChannelInfo, error) {
	ci, ok := xcc.(*ChannelInfo)
	if !ok {
		return nil, fmt.Errorf("cannot resolve '%v'", xcc)
	}
	return ci, nil
}

// ResolveChannel implements CrossChainChannelResolver.ResolveChannel
func (r ChannelInfoResolver) ResolveChannel(ctx sdk.Context, channel *ChannelInfo) (CrossChainChannel, error) {
	// check if given channel exists in channelKeeper
	_, found := r.channelKeeper.GetChannel(ctx, channel.Port, channel.Channel)
	if !found {
		return nil, fmt.Errorf("channel '%v' not found", channel.String())
	}
	return channel, nil
}

// ConvertCrossChainChannel returns a xcc of callee in caller's context
func (r ChannelInfoResolver) ConvertCrossChainChannel(ctx sdk.Context, calleeXCC CrossChainChannel, callerXCC CrossChainChannel) (CrossChainChannel, error) {
	isLocalCallee := r.IsSelfCrossChainChannel(ctx, calleeXCC)
	isLocalCaller := r.IsSelfCrossChainChannel(ctx, callerXCC)

	if !isLocalCallee && !isLocalCaller {
		return nil, fmt.Errorf("either callee or caller must be self xcc")
	} else if !isLocalCallee && isLocalCaller {
		return calleeXCC, nil
	} else if !isLocalCaller && isLocalCallee {
		callerChannelInfo, err := r.ResolveCrossChainChannel(ctx, callerXCC)
		if err != nil {
			return nil, err
		}
		callerChannel, found := r.channelKeeper.GetChannel(ctx, callerChannelInfo.Port, callerChannelInfo.Channel)
		if !found {
			return nil, fmt.Errorf("channel '%v' not found", callerChannelInfo.String())
		}
		return &ChannelInfo{Port: callerChannel.GetCounterparty().GetPortID(), Channel: callerChannel.GetCounterparty().GetChannelID()}, nil
	} else {
		panic("unreachable")
	}
}

// GetSelfCrossChainChannel implements CrossChainChannelResolver.GetSelfCrossChainChannel
func (ChannelInfoResolver) GetSelfCrossChainChannel(ctx sdk.Context) CrossChainChannel {
	return &ChannelInfo{}
}

// IsSelfCrossChainChannel implements CrossChainChannelResolver.IsSelfCrossChainChannel
func (r ChannelInfoResolver) IsSelfCrossChainChannel(ctx sdk.Context, xcc CrossChainChannel) bool {
	return xcc.Equal(r.GetSelfCrossChainChannel(ctx))
}

// Capabilities implements CrossChainChannelResolver.Capabilities
func (r ChannelInfoResolver) Capabilities() CrossChainChannelResolverCapabilities {
	return crossChainChannelResolverCapabilities{
		crossChainCalls: map[CommitProtocol]bool{
			COMMIT_PROTOCOL_SIMPLE: true,
			COMMIT_PROTOCOL_TPC:    false,
		},
	}
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
