package core

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/datachainlab/cross/x/core/keeper"
	"github.com/datachainlab/cross/x/core/types"
)

// NewHandler ...
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.MsgInitiate:
			// return handleMsgInitiate(ctx, k)
			panic("not implemented error")
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// func handleMsgInitiate(ctx sdk.Context, k keeper.Keeper, packetSender types.PacketSender,  msg *types.MsgInitiate) (*sdk.Result, error) {
// 	var data []byte
// 	switch msg.CommitProtocol {
// 	case types.COMMIT_PROTOCOL_SIMPLE:
// 		txID, err := k.SimpleKeeper().SendCall(ctx, packetSender, contractHandler, msg, msg.ContractTransactions)
// 		if err != nil {
// 			return nil, sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
// 		}
// 		data = txID[:]
// 	case types.COMMIT_PROTOCOL_TPC:
// 		txID, err := k.TPCKeeper().MulticastPacketPrepare(ctx, packetSender, msg.Sender, msg, msg.ContractTransactions)
// 		if err != nil {
// 			return nil, sdkerrors.Wrap(types.ErrFailedInitiateTx, err.Error())
// 		}
// 		data = txID[:]
// 	default:
// 		return nil, fmt.Errorf("unknown Commit protocol '%v'", msg.CommitProtocol)
// 	}

// 	return &sdk.Result{Data: data, Events: ctx.EventManager().ABCIEvents()}, nil
// }
