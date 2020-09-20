package types

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/datachainlab/cross/x/ibc/cross"
	"github.com/datachainlab/cross/x/ibc/cross/types"
)

const (
	ObjectTypeHTTP cross.ObjectType = 10
)

type HTTPResolver struct {
	Router ServerRouter
}

type ServerRouter interface {
	Route(id types.ChainID, bz []byte) (cross.Object, error)
}

var _ cross.ObjectResolver = (*HTTPResolver)(nil)

func NewHTTPResolver(router ServerRouter) HTTPResolver {
	return HTTPResolver{Router: router}
}

func (rs HTTPResolver) Resolve(id types.ChainID, bz []byte) (cross.Object, error) {
	return rs.Router.Route(id, bz)
}

var _ cross.Object = (*HTTPObject)(nil)

type HTTPServerInfo struct {
	Address string
}

type HTTPObject struct {
	K          []byte
	ServerInfo *HTTPServerInfo
}

func (HTTPObject) Type() cross.ObjectType {
	return ObjectTypeHTTP
}

func (HTTPObject) ChainID() types.ChainID {
	panic("not implemented error")
}

func (o HTTPObject) Key() []byte {
	return o.K
}

func (o HTTPObject) Evaluate(_ []byte) ([]byte, error) {
	// TODO build correct URL and set it
	// should we retry to request when it's failed?
	resp, err := http.DefaultClient.Get("http://" + o.ServerInfo.Address + "/cross/contract/call")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type HTTPServerRouter struct {
	servers map[string]map[string]*HTTPServerInfo
}

var _ ServerRouter = (*HTTPServerRouter)(nil)

func NewHTTPServerRouter() *HTTPServerRouter {
	return &HTTPServerRouter{servers: make(map[string]map[string]*HTTPServerInfo)}
}

func (r *HTTPServerRouter) AddRoute(chainID types.ChainID, id string, info HTTPServerInfo) {
	cidStr := chainID.String()
	if r.servers[cidStr] == nil {
		r.servers[cidStr] = make(map[string]*HTTPServerInfo)
	}
	r.servers[cidStr][id] = &info
}

func (r HTTPServerRouter) Route(chainID types.ChainID, bz []byte) (cross.Object, error) {
	call, err := DecodeContractCallInfo(bz)
	if err != nil {
		return nil, err
	}
	m := r.servers[chainID.String()]
	if m == nil {
		return nil, fmt.Errorf("chainID '%v' not found", chainID.String())
	}
	info, ok := m[call.ID]
	if !ok {
		return nil, fmt.Errorf("id '%v' not found", call.ID)
	}
	return &HTTPObject{K: bz, ServerInfo: info}, nil
}
