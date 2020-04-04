package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/datachainlab/cross/x/ibc/contract/types"
	"github.com/datachainlab/cross/x/ibc/cross"
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

type StateProvider = func(sdk.KVStore) cross.State

type contractHandler struct {
	keeper        Keeper
	routes        map[string]Contract
	stateProvider StateProvider
}

var _ cross.ContractHandler = (*contractHandler)(nil)

func NewContractHandler(k Keeper, stateProvider StateProvider) *contractHandler {
	return &contractHandler{keeper: k, routes: make(map[string]Contract), stateProvider: stateProvider}
}

func (h *contractHandler) Handle(ctx sdk.Context, contract []byte) (state cross.State, res cross.ContractHandlerResult, err error) {
	info, err := types.DecodeContractSignature(contract)
	if err != nil {
		return nil, nil, err
	}
	st, err := h.GetState(ctx, contract)
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
	c := NewContext(signers, info.Args)
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

func (h *contractHandler) GetState(ctx sdk.Context, contract []byte) (cross.State, error) {
	info, err := types.DecodeContractSignature(contract)
	if err != nil {
		return nil, err
	}
	return h.stateProvider(h.keeper.GetContractStateStore(ctx, []byte(info.ID))), nil
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
}

type ccontext struct {
	signers      []sdk.AccAddress
	args         [][]byte
	eventManager *sdk.EventManager
}

func NewContext(signers []sdk.AccAddress, args [][]byte) Context {
	return &ccontext{signers: signers, args: args, eventManager: sdk.NewEventManager()}
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
