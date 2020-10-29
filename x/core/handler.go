package core

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/datachainlab/cross/x/core/keeper"
	"github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
)

// NewHandler ...
func NewHandler(k keeper.Keeper, packetMiddleware packets.PacketMiddleware) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		ctx, ps, err := packetMiddleware.HandleMsg(ctx, msg, packets.NewBasicPacketSender(k.ChannelKeeper()))
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "failed to handle request: %v", err)
		}
		switch msg := msg.(type) {
		case *types.MsgInitiate:
			return handleMsgInitiate(ctx, k, ps, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

func handleMsgInitiate(ctx sdk.Context, k keeper.Keeper, packetSender packets.PacketSender, msg *types.MsgInitiate) (*sdk.Result, error) {
	panic("not implemented error")
}
