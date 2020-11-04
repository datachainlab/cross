package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
	"github.com/datachainlab/cross/x/utils"
)

type Keeper struct {
	cdc      codec.Marshaler
	storeKey sdk.StoreKey

	channelKeeper crosstypes.ChannelKeeper
	portKeeper    crosstypes.PortKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper
	commitStore   crosstypes.CommitStore

	contractHandler  crosstypes.ContractHandler
	resolverProvider crosstypes.ObjectResolverProvider
	channelResolver  crosstypes.ChannelResolver
}

func NewKeeper(
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	channelKeeper crosstypes.ChannelKeeper,
	portKeeper crosstypes.PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	contractHandler crosstypes.ContractHandler,
	commitStore crosstypes.CommitStore,
) Keeper {
	return Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		channelKeeper:   channelKeeper,
		portKeeper:      portKeeper,
		scopedKeeper:    scopedKeeper,
		commitStore:     commitStore,
		contractHandler: contractHandler,
	}
}

func (k Keeper) ChannelKeeper() crosstypes.ChannelKeeper {
	return k.channelKeeper
}

func (k Keeper) ChannelResolver() crosstypes.ChannelResolver {
	return k.channelResolver
}

func (k Keeper) PrepareCommit(
	ctx sdk.Context,
	txID crosstypes.TxID,
	txIndex crosstypes.TxIndex,
	tx crosstypes.ContractTransaction,
	links []crosstypes.Object,
) error {
	res, err := k.processTransaction(ctx, txIndex, tx, links, crosstypes.AtomicMode)
	if err != nil {
		return err
	}
	k.SetContractResult(ctx, txID, txIndex, *res)
	return k.commitStore.Precommit(ctx, makeContractTransactionID(txID, txIndex))
}

func (k Keeper) processTransaction(
	ctx sdk.Context,
	txIndex crosstypes.TxIndex,
	tx crosstypes.ContractTransaction,
	links []crosstypes.Object,
	commitMode crosstypes.CommitMode,
) (res *crosstypes.ContractHandlerResult, err error) {
	// TODO resolverProvider can be moved into contract package?
	rs, err := k.resolverProvider(links)
	if err != nil {
		return nil, err
	}

	// Setup a context
	goCtx := crosstypes.SetupContractContext(
		sdk.WrapSDKContext(ctx),
		crosstypes.ContractRuntimeInfo{
			CommitMode:             commitMode,
			StateConstraintType:    tx.StateConstraint.Type,
			ExternalObjectResolver: rs,
		},
	)
	ops, err := k.contractHandler(
		goCtx,
		tx.CallInfo,
	)
	if err != nil {
		return nil, err
	}

	if rv := tx.ReturnValue; !rv.IsNil() && !rv.Equal(crosstypes.NewReturnValue(res.Data)) {
		return nil, fmt.Errorf("unexpected return-value: expected='%X' actual='%X'", *rv, res.Data)
	}

	if !tx.StateConstraint.Ops.Equal(ops) {
		return nil, fmt.Errorf("unexpected ops: actual(%v) != expected(%v)", ops.String(), tx.StateConstraint.Ops.String())
	}

	return res, nil
}

func (k Keeper) SendPacket(
	ctx sdk.Context,
	packetSender packets.PacketSender,
	payload packets.PacketDataPayload,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	pd := packets.NewPacketData(nil, utils.MustMarshalJSONAny(k.cdc, payload))
	bz, err := packets.MarshalPacketData(pd)
	if err != nil {
		return err
	}

	// get the next sequence
	seq, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.ErrSequenceSendNotFound
	}
	packet := channeltypes.NewPacket(
		bz,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	if err := packetSender.SendPacket(ctx, channelCap, packets.NewOutgoingPacket(packet, pd, payload)); err != nil {
		return err
	}
	return nil
}
