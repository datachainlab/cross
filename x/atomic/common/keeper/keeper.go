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
	"github.com/datachainlab/cross/x/atomic/common/types"
	crosstypes "github.com/datachainlab/cross/x/core/types"
	"github.com/datachainlab/cross/x/packets"
	"github.com/datachainlab/cross/x/utils"
)

type Keeper struct {
	cdc       codec.Marshaler
	storeKey  sdk.StoreKey
	keyPrefix []byte

	channelKeeper crosstypes.ChannelKeeper
	portKeeper    crosstypes.PortKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper
	commitStore   crosstypes.CommitStore

	contractModule          crosstypes.ContractModule
	contractHandleDecorator crosstypes.ContractHandleDecorator
	resolverProvider        crosstypes.ObjectResolverProvider
	xccResolver             crosstypes.CrossChainChannelResolver
}

func NewKeeper(
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	keyPrefix []byte,
	channelKeeper crosstypes.ChannelKeeper,
	portKeeper crosstypes.PortKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	contractModule crosstypes.ContractModule,
	contractHandleDecorator crosstypes.ContractHandleDecorator,
	xccResolver crosstypes.CrossChainChannelResolver,
	commitStore crosstypes.CommitStore,
) Keeper {
	return Keeper{
		cdc:                     cdc,
		storeKey:                storeKey,
		keyPrefix:               keyPrefix,
		channelKeeper:           channelKeeper,
		portKeeper:              portKeeper,
		scopedKeeper:            scopedKeeper,
		commitStore:             commitStore,
		contractModule:          contractModule,
		contractHandleDecorator: contractHandleDecorator,
		resolverProvider:        crosstypes.DefaultResolverProvider(),
		xccResolver:             xccResolver,
	}
}

func (k Keeper) ChannelKeeper() crosstypes.ChannelKeeper {
	return k.channelKeeper
}

func (k Keeper) CrossChainChannelResolver() crosstypes.CrossChainChannelResolver {
	return k.xccResolver
}

func (k Keeper) PrepareCommit(
	ctx sdk.Context,
	txID crosstypes.TxID,
	txIndex crosstypes.TxIndex,
	tx crosstypes.ContractTransaction,
	links []crosstypes.Object,
) error {
	ctx, err := k.setupContext(ctx, tx, links, crosstypes.AtomicMode)
	if err != nil {
		return err
	}
	res, err := k.processTransaction(ctx, tx)
	if err != nil {
		return err
	}
	k.SetContractCallResult(ctx, txID, txIndex, *res)
	return k.commitStore.Precommit(ctx, makeContractTransactionID(txID, txIndex))
}

func (k Keeper) setupContext(
	ctx sdk.Context,
	tx crosstypes.ContractTransaction,
	links []crosstypes.Object,
	commitMode crosstypes.CommitMode,
) (sdk.Context, error) {
	// TODO resolverProvider can be moved into contract package?
	rs, err := k.resolverProvider(k.cdc, links)
	if err != nil {
		return ctx, err
	}

	// Setup a context
	ctx = crosstypes.SetupContractContext(
		ctx,
		tx.Signers,
		crosstypes.ContractRuntimeInfo{
			CommitMode:             commitMode,
			ExternalObjectResolver: rs,
		},
	)
	goCtx, err := k.contractHandleDecorator.Handle(ctx.Context(), tx.CallInfo)
	if err != nil {
		return ctx, err
	}
	return ctx.WithContext(goCtx), nil
}

func (k Keeper) processTransaction(
	ctx sdk.Context,
	tx crosstypes.ContractTransaction,
) (res *crosstypes.ContractCallResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = types.NewErrContractCall(e)
			} else {
				err = types.NewErrContractCall(fmt.Errorf("type=%T value=%#v", e, e))
			}
		}
	}()

	res, err = k.contractModule.OnContractCall(
		sdk.WrapSDKContext(ctx),
		tx.CallInfo,
	)
	if err != nil {
		return nil, err
	}

	if !tx.ReturnValue.IsNil() && !tx.ReturnValue.Equal(crosstypes.NewReturnValue(res.Data)) {
		return nil, fmt.Errorf("unexpected return-value: expected='%X' actual='%X'", *tx.ReturnValue, res.Data)
	}

	return res, nil
}

func (k Keeper) CommitImmediately(
	ctx sdk.Context,
	txID crosstypes.TxID,
	txIndex crosstypes.TxIndex,
	tx crosstypes.ContractTransaction,
	links []crosstypes.Object,
) (*crosstypes.ContractCallResult, error) {
	ctx, err := k.setupContext(ctx, tx, links, crosstypes.BasicMode)
	if err != nil {
		return nil, err
	}
	res, err := k.processTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}
	k.commitStore.CommitImmediately(ctx)
	return res, nil
}

// Commit commits the transaction
func (k Keeper) Commit(
	ctx sdk.Context,
	txID crosstypes.TxID,
	txIndex crosstypes.TxIndex,
) (*crosstypes.ContractCallResult, error) {
	if err := k.commitStore.Commit(ctx, makeContractTransactionID(txID, txIndex)); err != nil {
		return nil, err
	}
	res := k.GetContractCallResult(ctx, txID, txIndex)
	// TODO calls OnCommit handler
	k.RemoveContractCallResult(ctx, txID, txIndex)
	return res, nil
}

// Abort aborts the transaction
func (k Keeper) Abort(
	ctx sdk.Context,
	txID crosstypes.TxID,
	txIndex crosstypes.TxIndex,
) error {
	if err := k.commitStore.Abort(ctx, makeContractTransactionID(txID, txIndex)); err != nil {
		return err
	}
	// TODO calls OnAbort handler
	k.RemoveContractCallResult(ctx, txID, txIndex)
	return nil
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
	bz, err := packets.MarshalJSONPacketData(pd)
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
