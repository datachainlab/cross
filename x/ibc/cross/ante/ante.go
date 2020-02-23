package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/datachainlab/cross/x/ibc/cross"
)

// ProofVerificationDecorator handles messages that contains application specific packet types,
// including MsgPacket, MsgAcknowledgement, MsgTimeout.
// MsgUpdateClients are also handled here to perform atomic multimsg transaction
type ProofVerificationDecorator struct {
	clientKeeper  client.Keeper
	channelKeeper channel.Keeper
}

// NewProofVerificationDecorator constructs new ProofverificationDecorator
func NewProofVerificationDecorator(clientKeeper client.Keeper, channelKeeper channel.Keeper) ProofVerificationDecorator {
	return ProofVerificationDecorator{
		clientKeeper:  clientKeeper,
		channelKeeper: channelKeeper,
	}
}

// AnteHandle executes MsgUpdateClient, MsgPacket, MsgAcknowledgement, MsgTimeout.
// The packet execution messages are then passed to the respective application handlers.
func (pvr ProofVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		var err error
		switch msg := msg.(type) {
		case clientexported.MsgUpdateClient:
			err = pvr.clientKeeper.UpdateClient(ctx, msg.GetClientID(), msg.GetHeader())
		case channel.MsgPacket:
			_, err = pvr.channelKeeper.RecvPacket(ctx, msg.Packet, msg.Proof, msg.ProofHeight)
		// case channel.MsgAcknowledgement:
		// 	_, err = pvr.channelKeeper.AcknowledgePacket(ctx, msg.Packet, msg.Acknowledgement, msg.Proof, msg.ProofHeight)
		// case channel.MsgTimeout:
		// 	_, err = pvr.channelKeeper.TimeoutPacket(ctx, msg.Packet, msg.Proof, msg.ProofHeight, msg.NextSequenceRecv)
		case cross.MultiplePackets:
			for _, msg := range msg.Packets() {
				_, err = pvr.channelKeeper.RecvPacket(ctx, msg.Packet, msg.Proof, msg.ProofHeight)
				if err != nil {
					break
				}
			}
		}

		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}
