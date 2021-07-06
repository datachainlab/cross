package router

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	"github.com/datachainlab/cross/x/packets"
)

type Router interface {
	AddRoute(typeName string, h PacketHandler)
	GetRoute(typeName string) (PacketHandler, bool)
}

type PacketHandler interface {
	HandlePacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
		ip packets.IncomingPacket,
	) (*sdk.Result, *packets.PacketAcknowledgementData, error)
	HandleACK(
		ctx sdk.Context,
		packet channeltypes.Packet,
		ip packets.IncomingPacket,
		ipa packets.IncomingPacketAcknowledgement,
	) (*sdk.Result, error)
}

type router struct {
	routes map[string]PacketHandler
	sealed bool
}

func NewRouter() Router {
	return &router{
		routes: make(map[string]PacketHandler),
	}
}

func (r *router) AddRoute(typeName string, h PacketHandler) {
	if r.sealed {
		panic(fmt.Sprintf("router sealed; cannot register %s route callbacks", typeName))
	}
	if _, ok := r.routes[typeName]; ok {
		panic(fmt.Sprintf("Route type '%v' already exist", typeName))
	}
	r.routes[typeName] = h
}

func (r *router) GetRoute(typeName string) (PacketHandler, bool) {
	route, ok := r.routes[typeName]
	return route, ok
}

// Seal prevents the Router from any subsequent route handlers to be registered.
// Seal will panic if called more than once.
func (r *router) Seal() {
	if r.sealed {
		panic("router already sealed")
	}
	r.sealed = true
}

// Sealed returns a boolean signifying if the Router is sealed or not.
func (r router) Sealed() bool {
	return r.sealed
}
