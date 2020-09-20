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

// HTTPObjectResolver resolves a given ChainID and call info to HTTPObject
type HTTPObjectResolver struct {
	Router ServerRouter
}

// ServerRouter manages routes to contract functions on http server
type ServerRouter interface {
	Route(id types.ChainID, bz []byte) (cross.Object, error)
}

var _ cross.ObjectResolver = (*HTTPObjectResolver)(nil)

// NewHTTPObjectResolver returns a new HTTPObjectResolver
func NewHTTPObjectResolver(router ServerRouter) HTTPObjectResolver {
	return HTTPObjectResolver{Router: router}
}

// Resolve implements Resolver.Resolve
func (rs HTTPObjectResolver) Resolve(id types.ChainID, bz []byte) (cross.Object, error) {
	return rs.Router.Route(id, bz)
}

var _ cross.Object = (*HTTPObject)(nil)

// HTTPServerInfo is an info of http server
type HTTPServerInfo struct {
	ChainID types.ChainID
	Address string
}

// HTTPObject is an Object that wraps HTTPServerInfo
type HTTPObject struct {
	K          []byte
	ServerInfo HTTPServerInfo
}

// Type implements Object.Type
func (HTTPObject) Type() cross.ObjectType {
	return ObjectTypeHTTP
}

// ChainID implements Object.ChainID
func (o HTTPObject) ChainID() types.ChainID {
	return o.ServerInfo.ChainID
}

// Key implements Object.Key
func (o HTTPObject) Key() []byte {
	return o.K
}

// Evaluate implements Object.Evaluate
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

// HTTPServerRouter manages the routes to each http servers that have contract functions
type HTTPServerRouter struct {
	servers map[string]map[string]HTTPServerInfo
}

var _ ServerRouter = (*HTTPServerRouter)(nil)

// NewHTTPServerRouter returns a new HTTPServerRouter
func NewHTTPServerRouter() *HTTPServerRouter {
	return &HTTPServerRouter{servers: make(map[string]map[string]HTTPServerInfo)}
}

// AddRoute adds a new route to the http server
func (r *HTTPServerRouter) AddRoute(chainID types.ChainID, id string, info HTTPServerInfo) {
	cidStr := chainID.String()
	if r.servers[cidStr] == nil {
		r.servers[cidStr] = make(map[string]HTTPServerInfo)
	}
	r.servers[cidStr][id] = info
}

// Route returns HTTPObject that indicates the object on matched http server
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
