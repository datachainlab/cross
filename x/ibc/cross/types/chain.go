package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ChainID represents an ID of chain that contains a contract function to be called
type ChainID interface {
	Equal(ChainID) bool
}

func (ci ChannelInfo) Equal(other ChainID) bool {
	return ci == other
}

type ChannelResolver interface {
	Resolve(ctx sdk.Context, chainID ChainID) (*ChannelInfo, error)
}

type ChannelInfoResolver struct{}

var _ ChannelResolver = (*ChannelInfoResolver)(nil)

func (r ChannelInfoResolver) Resolve(ctx sdk.Context, chainID ChainID) (*ChannelInfo, error) {
	ci, ok := chainID.(ChannelInfo)
	if !ok {
		return nil, fmt.Errorf("cannot resolve '%v'", chainID)
	}
	return &ci, nil
}

// TODO move this into other package

// var _ ChainID = (*DNSChainID)(nil)

// type DNSChainID struct {
// 	DomainName string
// }

// func (c DNSChainID) Equal(chainID ChainID) bool {
// 	return c == chainID
// }

// type DNSResolver struct {
// 	primaryDNSID string
// }

// var _ ChannelResolver = (*DNSResolver)(nil)

// func NewDNSResolver(primaryDNSID string) DNSResolver {
// 	return DNSResolver{primaryDNSID: primaryDNSID}
// }

// func (r DNSResolver) SetupContextWithReceivingPacket(ctx sdk.Context, packetData []byte) (sdk.Context, error) {
// 	// TODO implement this
// 	// parse header and get DNS-ID from its
// 	return ctx.WithValue("dns", "xxxx"), nil
// }

// func (r DNSResolver) MatchContext(ctx sdk.Context) bool {
// 	return r.primaryDNSID == ctx.Value("dns").(string)
// }

// func (r DNSResolver) Resolve(ctx sdk.Context, chainID ChainID) (*ChannelInfo, error) {
// 	// use primaryDNSID to resolve chainID
// 	return nil, fmt.Errorf("not implemented error")
// }
