package common

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/capability"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/datachainlab/cross/x/ibc/cross/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

type Keeper struct {
	cdc      *codec.Codec // The wire codec for binary encoding/decoding.
	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	channelKeeper types.ChannelKeeper
	portKeeper    types.PortKeeper
	scopedKeeper  capability.ScopedKeeper

	resolverProvider types.ObjectResolverProvider
	channelResolver  types.ChannelResolver
}

func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	scopedKeeper capability.ScopedKeeper,
	resolverProvider types.ObjectResolverProvider,
	channelResolver types.ChannelResolver,
) Keeper {
	return Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		channelKeeper:    channelKeeper,
		portKeeper:       portKeeper,
		scopedKeeper:     scopedKeeper,
		resolverProvider: resolverProvider,
		channelResolver:  channelResolver,
	}
}

func (k Keeper) ChannelKeeper() types.ChannelKeeper {
	return k.channelKeeper
}

func (k Keeper) PortKeeper() types.PortKeeper {
	return k.portKeeper
}

func (k Keeper) ScopedKeeper() capability.ScopedKeeper {
	return k.scopedKeeper
}

func (k Keeper) SendPacket(
	ctx sdk.Context,
	packetSender types.PacketSender,
	payload types.PacketDataPayload,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	timeoutHeight uint64,
	timeoutTimestamp uint64,
) error {
	data := types.NewPacketData(nil, payload.GetBytes())
	// Wrap raw data with a container
	bz, err := types.MarshalPacketData(data)
	if err != nil {
		return err
	}

	// get the next sequence
	seq, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channel.ErrSequenceSendNotFound
	}
	packet := channel.NewPacket(
		bz,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, ibctypes.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrap(channel.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	if err := packetSender.SendPacket(ctx, channelCap, types.NewOutgoingPacket(packet, data, payload)); err != nil {
		return err
	}

	k.SetUnacknowledgedPacket(ctx, sourcePort, sourceChannel, seq)
	return nil
}

func (k Keeper) PrepareTransaction(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	txID types.TxID,
	txIndex types.TxIndex,
	tx types.ContractTransaction,
	links []types.Object,
) error {
	store, res, err := k.processTransaction(ctx, contractHandler, txID, txIndex, tx, links)
	if err != nil {
		return err
	}
	k.SetContractResult(ctx, txID, txIndex, res)
	return store.Precommit(MakeStoreTransactionID(txID, txIndex))
}

func (k Keeper) CommitTransaction(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	txID types.TxID,
	txIndex types.TxIndex,
	txInfo *types.TxInfo,
) (types.ContractHandlerResult, error) {
	state, err := contractHandler.GetState(ctx, txInfo.ContractCallInfo, types.ContractRuntimeInfo{StateConstraintType: types.NoStateConstraint})
	if err != nil {
		return nil, err
	}
	if err := state.Commit(MakeStoreTransactionID(txID, txIndex)); err != nil {
		return nil, err
	}
	r, ok := k.GetContractResult(ctx, txID, txIndex)
	if !ok {
		return nil, fmt.Errorf("Can't find the execution result of contract handler")
	}
	res := contractHandler.OnCommit(ctx, r)
	k.RemoveContractResult(ctx, txID, txIndex)
	return res, nil
}

func (k Keeper) AbortTransaction(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	txID types.TxID,
	txIndex types.TxIndex,
	txInfo *types.TxInfo,
) error {
	state, err := contractHandler.GetState(ctx, txInfo.ContractCallInfo, types.ContractRuntimeInfo{StateConstraintType: types.NoStateConstraint})
	if err != nil {
		return err
	}
	if err := state.Discard(MakeStoreTransactionID(txID, txIndex)); err != nil {
		return err
	}
	return nil
}

func (k Keeper) CommitImmediatelyTransaction(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	txID types.TxID,
	txIndex types.TxIndex,
	tx types.ContractTransaction,
	links []types.Object,
) error {
	store, _, err := k.processTransaction(ctx, contractHandler, txID, txIndex, tx, links)
	if err != nil {
		return err
	}
	return store.CommitImmediately()
}

func (k Keeper) processTransaction(
	ctx sdk.Context,
	contractHandler types.ContractHandler,
	txID types.TxID,
	txIndex types.TxIndex,
	tx types.ContractTransaction,
	links []types.Object,
) (store types.Committer, res types.ContractHandlerResult, err error) {
	rs, err := k.resolverProvider(links)
	if err != nil {
		return nil, nil, err
	}
	store, res, err = contractHandler.Handle(
		types.WithSigners(ctx, tx.Signers),
		tx.CallInfo,
		types.ContractRuntimeInfo{StateConstraintType: tx.StateConstraint.Type, ExternalObjectResolver: rs},
	)
	if err != nil {
		return nil, nil, err
	}

	if rv := tx.ReturnValue; !rv.IsNil() && !rv.Equal(res.GetData()) {
		return nil, nil, fmt.Errorf("unexpected return-value: expected='%X' actual='%X'", *rv, res.GetData())
	}

	if ops := store.OPs(); !ops.Equal(tx.StateConstraint.OPs) {
		return nil, nil, fmt.Errorf("unexpected ops: actual(%v) != expected(%v)", ops.String(), tx.StateConstraint.OPs.String())
	}

	return store, res, nil
}

// PacketExecuted defines a wrapper function for the channel Keeper's function
// in order to expose it to the cross handler.
// Keeper retreives channel capability and passes it into channel keeper for authentication
func (k Keeper) PacketExecuted(ctx sdk.Context, packet channelexported.PacketI, acknowledgement []byte) error {
	chanCap, ok := k.scopedKeeper.GetCapability(ctx, ibctypes.ChannelCapabilityPath(packet.GetDestPort(), packet.GetDestChannel()))
	if !ok {
		return sdkerrors.Wrap(channel.ErrChannelCapabilityNotFound, "channel capability could not be retrieved for packet")
	}
	return k.channelKeeper.PacketExecuted(ctx, chanCap, packet, acknowledgement)
}

func (k Keeper) SetTx(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex, tx types.TxInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	store.Set(types.KeyTx(txID, txIndex), bz)
}

func (k Keeper) EnsureTxStatus(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex, status uint8) (*types.TxInfo, error) {
	tx, found := k.GetTx(ctx, txID, txIndex)
	if !found {
		return nil, fmt.Errorf("txID '%x' not found", txID)
	}
	if tx.Status == status {
		return tx, nil
	} else {
		return nil, fmt.Errorf("expected status is %v, but got %v", status, tx.Status)
	}
}

func (k Keeper) UpdateTxStatus(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex, status uint8) error {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyTx(txID, txIndex))
	if bz == nil {
		return fmt.Errorf("txID '%x' not found", txID)
	}
	var tx types.TxInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)
	tx.Status = status
	k.SetTx(ctx, txID, txIndex, tx)
	return nil
}

func (k Keeper) GetTx(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) (*types.TxInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyTx(txID, txIndex))
	if bz == nil {
		return nil, false
	}
	var tx types.TxInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)
	return &tx, true
}

func (k Keeper) SetCoordinator(ctx sdk.Context, txID types.TxID, ci types.CoordinatorInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(ci)
	store.Set(types.KeyCoordinator(txID), bz)
}

func (k Keeper) GetCoordinator(ctx sdk.Context, txID types.TxID) (*types.CoordinatorInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyCoordinator(txID))
	if bz == nil {
		return nil, false
	}
	var ci types.CoordinatorInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &ci)
	return &ci, true
}

func (k Keeper) UpdateCoordinatorStatus(ctx sdk.Context, txID types.TxID, status uint8) error {
	ci, found := k.GetCoordinator(ctx, txID)
	if !found {
		return fmt.Errorf("txID '%x' not found", txID)
	}
	ci.Status = status
	k.SetCoordinator(ctx, txID, *ci)
	return nil
}

func (k Keeper) EnsureCoordinatorStatus(ctx sdk.Context, txID types.TxID, status uint8) (*types.CoordinatorInfo, error) {
	ci, found := k.GetCoordinator(ctx, txID)
	if !found {
		return nil, fmt.Errorf("txID '%x' not found", txID)
	}
	if ci.Status == status {
		return ci, nil
	} else {
		return nil, fmt.Errorf("expected status is %v, but got %v", status, ci.Status)
	}
}

func (k Keeper) SetUnacknowledgedPacket(ctx sdk.Context, sourcePort, sourceChannel string, seq uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyUnacknowledgedPacket(sourcePort, sourceChannel, seq), []byte{0})
}

func (k Keeper) RemoveUnacknowledgedPacket(ctx sdk.Context, sourcePort, sourceChannel string, seq uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyUnacknowledgedPacket(sourcePort, sourceChannel, seq))
}

func (k Keeper) IterateUnacknowledgedPackets(ctx sdk.Context, cb func(key []byte) bool) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.KeyPrefixBytes(types.KeyUnacknowledgedPacketPrefix)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		key = []byte(strings.TrimPrefix(string(key), string(prefix)))
		if cb(key) {
			break
		}
	}
}

func (k Keeper) SetContractResult(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex, result types.ContractHandlerResult) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(result)
	store.Set(types.KeyContractResult(txID, txIndex), bz)
}

func (k Keeper) GetContractResult(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) (types.ContractHandlerResult, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyContractResult(txID, txIndex))
	if bz == nil {
		return nil, false
	}
	var result types.ContractHandlerResult
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &result)
	return result, true
}

func (k Keeper) RemoveContractResult(ctx sdk.Context, txID types.TxID, txIndex types.TxIndex) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyContractResult(txID, txIndex))
}

func (k Keeper) ChannelResolver() types.ChannelResolver {
	return k.channelResolver
}

func MakeTxID(ctx sdk.Context, msg types.MsgInitiate) types.TxID {
	var txID [32]byte

	a := tmhash.Sum(msg.GetSignBytes())
	b := tmhash.Sum(types.MakeHashFromABCIHeader(ctx.BlockHeader()).Hash())

	h := tmhash.New()
	h.Write(a)
	h.Write(b)
	copy(txID[:], h.Sum(nil))
	return txID
}

func MakeStoreTransactionID(txID types.TxID, txIndex uint8) []byte {
	size := len(txID)
	bz := make([]byte, size+1)
	copy(bz[:size], txID[:])
	bz[size] = txIndex
	return bz
}
