package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/contract/types"
	"github.com/datachainlab/cross/x/ibc/cross"
	crosstypes "github.com/datachainlab/cross/x/ibc/cross/types"
)

var _ cross.ContractHandler = (*contractHandler)(nil)

type contract struct {
	methods map[string]Method
}

func (c contract) CallMethod(ctx Context, store cross.Store, method string) ([]byte, error) {
	m, ok := c.methods[method]
	if !ok {
		return nil, fmt.Errorf("method '%v' not found", method)
	}
	return m.F(ctx, store)
}

type Method struct {
	Name string
	F    func(ctx Context, store cross.Store) ([]byte, error)
}

func NewContract(methods []Method) Contract {
	mm := make(map[string]Method)
	for _, m := range methods {
		mm[m.Name] = m
	}
	return &contract{methods: mm}
}

type Contract interface {
	CallMethod(ctx Context, store cross.Store, method string) ([]byte, error)
}

type StateProvider = func(sdk.KVStore, cross.StateConstraintType) cross.State

type contractHandler struct {
	keeper          Keeper
	routes          map[string]Contract
	stateProvider   StateProvider
	channelResolver crosstypes.ChannelResolver
}

var _ cross.ContractHandler = (*contractHandler)(nil)

func NewContractHandler(k Keeper, stateProvider StateProvider, channelResolver crosstypes.ChannelResolver) *contractHandler {
	return &contractHandler{keeper: k, routes: make(map[string]Contract), stateProvider: stateProvider, channelResolver: channelResolver}
}

func (h *contractHandler) Handle(ctx sdk.Context, callInfo cross.ContractCallInfo, rtInfo cross.ContractRuntimeInfo) (state cross.State, res cross.ContractHandlerResult, err error) {
	info, err := types.DecodeContractCallInfo(callInfo)
	if err != nil {
		return nil, nil, err
	}
	st, err := h.GetState(ctx, callInfo, rtInfo)
	if err != nil {
		return nil, nil, err
	}
	route, ok := h.routes[info.ID]
	if !ok {
		return nil, nil, fmt.Errorf("route for '%v' not found", info.ID)
	}
	signers, ok := cross.SignersFromContext(ctx)
	if !ok {
		return nil, nil, fmt.Errorf("signer is not set")
	}
	c := NewContext(signers, info.Args, rtInfo)
	defer func() {
		if e := recover(); e != nil {
			if e2, ok := e.(error); ok {
				err = fmt.Errorf("error=%v object=%#v", e2.Error(), e)
			} else {
				err = fmt.Errorf("type=%T object=%#v", e, e)
			}
		}
	}()

	v, err := route.CallMethod(c, st, info.Method)
	if err != nil {
		return nil, nil, err
	}
	return st, types.NewContractHandlerResult(v, c.EventManager().Events()), nil
}

func (h *contractHandler) GetState(ctx sdk.Context, callInfo cross.ContractCallInfo, rtInfo cross.ContractRuntimeInfo) (cross.State, error) {
	info, err := types.DecodeContractCallInfo(callInfo)
	if err != nil {
		return nil, err
	}
	return h.stateProvider(h.keeper.GetContractStateStore(ctx, []byte(info.ID)), rtInfo.StateConstraintType), nil
}

func (h *contractHandler) OnCommit(ctx sdk.Context, result cross.ContractHandlerResult) cross.ContractHandlerResult {
	return result
}

func (h *contractHandler) AddRoute(id string, c Contract) {
	if _, ok := h.routes[id]; ok {
		panic("this route id already exists")
	}
	h.routes[id] = c
}

type Context interface {
	Signers() []sdk.AccAddress
	Args() [][]byte
	EventManager() *sdk.EventManager

	runtimeInfo() cross.ContractRuntimeInfo
}

type ccontext struct {
	signers      []sdk.AccAddress
	args         [][]byte
	eventManager *sdk.EventManager
	rtInfo       cross.ContractRuntimeInfo
}

func NewContext(signers []sdk.AccAddress, args [][]byte, rtInfo cross.ContractRuntimeInfo) Context {
	return &ccontext{signers: signers, args: args, eventManager: sdk.NewEventManager(), rtInfo: rtInfo}
}

func (c ccontext) Signers() []sdk.AccAddress {
	return c.signers
}

func (c ccontext) Args() [][]byte {
	return c.args
}

func (c ccontext) EventManager() *sdk.EventManager {
	return c.eventManager
}

func (c ccontext) runtimeInfo() cross.ContractRuntimeInfo {
	return c.rtInfo
}

// CallExternalFunc calls a contract function on external chain
func CallExternalFunc(ctx Context, id crosstypes.ChainID, callInfo types.ContractCallInfo, signers []sdk.AccAddress) []byte {
	rs := ctx.runtimeInfo().ExternalObjectResolver
	key := cross.MakeObjectKey(callInfo.Bytes(), signers)
	obj, err := rs.Resolve(id, key)
	if err != nil {
		panic(err)
	}
	v, err := obj.Evaluate(key)
	if err != nil {
		panic(err)
	}
	return v
}
