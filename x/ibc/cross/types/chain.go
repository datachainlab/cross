package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ChainID represents an ID of chain that contains a contract function to be called
type ChainID interface {
	Type() string
	Equal(ChainID) bool
	String() string
}

// Type implements ChainID.Type
func (ci ChannelInfo) Type() string {
	return "channelinfo"
}

// String implements ChainID.String
func (ci ChannelInfo) String() string {
	return fmt.Sprintf("%v/%v/%v", ci.Type(), ci.Channel, ci.Port)
}

// Equal implements ChainID.Equal
func (ci ChannelInfo) Equal(other ChainID) bool {
	return ci == other
}

// ChannelResolver defines the interface of resolver resolves chainID to ChannelInfo
type ChannelResolver interface {
	Resolve(ctx sdk.Context, chainID ChainID) (*ChannelInfo, error)
	Capabilities() ChannelResolverCapabilities
}

// ChannelResolverCapabilities defines the capabilities for the ChannelResolver
type ChannelResolverCapabilities interface {
	// CrossChainCalls returns true if support for cross-chain calls is enabled.
	CrossChainCalls() bool
}

type channelResolverCapabilities struct {
	crossChainCalls bool
}

// CrossChainCalls implements ChannelResolverCapabilities.CrossChainCalls
func (c channelResolverCapabilities) CrossChainCalls() bool {
	return c.crossChainCalls
}

// ChannelInfoResolver just returns a given ChannelInfo as is.
type ChannelInfoResolver struct{}

var _ ChannelResolver = (*ChannelInfoResolver)(nil)

// Resolve implements ChannelResolver.ResResolve
func (r ChannelInfoResolver) Resolve(ctx sdk.Context, chainID ChainID) (*ChannelInfo, error) {
	ci, ok := chainID.(ChannelInfo)
	if !ok {
		return nil, fmt.Errorf("cannot resolve '%v'", chainID)
	}
	return &ci, nil
}

// Capabilities implements ChannelResolver.Capabilities
func (r ChannelInfoResolver) Capabilities() ChannelResolverCapabilities {
	return channelResolverCapabilities{crossChainCalls: false}
}
