package types

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/datachainlab/cross/x/ibc/cross"
)

const (
	ObjectTypeHTTP cross.ObjectType = 10
)

type HTTPResolver struct {
	Router ServerRouter
}

type ServerRouter interface {
	Route(bz []byte) (cross.Object, error)
}

var _ cross.ObjectResolver = (*HTTPResolver)(nil)

func NewHTTPResolver(router ServerRouter) HTTPResolver {
	return HTTPResolver{Router: router}
}

func (rs HTTPResolver) Resolve(bz []byte) (cross.Object, error) {
	return rs.Router.Route(bz)
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
	servers map[string]*HTTPServerInfo
}

var _ ServerRouter = (*HTTPServerRouter)(nil)

func NewHTTPServerRouter() *HTTPServerRouter {
	return &HTTPServerRouter{servers: make(map[string]*HTTPServerInfo)}
}

func (r *HTTPServerRouter) AddRoute(id string, info HTTPServerInfo) {
	r.servers[id] = &info
}

func (r HTTPServerRouter) Route(bz []byte) (cross.Object, error) {
	call, err := DecodeContractCallInfo(bz)
	if err != nil {
		return nil, err
	}
	info, ok := r.servers[call.ID]
	if !ok {
		return nil, fmt.Errorf("id '%v' not found", call.ID)
	}
	return &HTTPObject{K: bz, ServerInfo: info}, nil
}
