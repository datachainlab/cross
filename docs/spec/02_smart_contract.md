# Smart contract

This section explains how smart contract works and how it is developed using Cross.

## The flow of smart contract execution

Let's take a look at the flow of executing smart contract; check out [this image](../images/packet-flow.png) to see the overall flow of Cross-chain transaction.

Each contract state transition is performed in two phases, "Prepare" and "Commit".

### Prepare phase

This phase is initiated by receiving a PreparePacket from the coordinator. 

The participant who receives a packet executes a smart contract according to contract call information written to PacketData. If successful, take the operations out of Store and make sure they match the operations in PacketData. If any of these processes fail, PacketPrepareAcknowledgement is sent to the coordinator with "Failed" set to status without writing anything to Store. If all processes are successful, operations are saved in Store and get a Lock for each key. When writing operations are performed on the State, the exclusive lock is obtained, and when read operations are performed, the shared lock is obtained. Finally, PacketPrepareAcknowledgement is sent to the coordinator with "OK" set to status.

### Commit phase

This phase is initiated by receiving a CommitPacket from the coordinator.

If the status of CommitPacket is "Commit", it first applies(commits) the operation saved in Prepare to state, then releases the acquired Lock, and finally returns PacketCommitAcknowledgement to the coordinator.
If the status of packet is "Abort", it first discards the operation saved in Prepare phase, then releases the acquired Lock, and finally returns PacketCommitAcknowledgement to the coordinator.

That's all the execution flow for the Contract.

Now, let's see how to actually run smart contract!

## How to execute a smart contract on cross-chain

Suppose that smart contract managing the reservation of Train and Hotel is in two different chains.

Note that currently, [the loopback client](https://github.com/cosmos/ics/blob/master/spec/ics-009-loopback-client/README.md) is not yet implemented in Cosmos-SDK, so you need a coordinator chain to call two contracts.(This will be fixed in the near future.)

Then, suppose that each of two chains has an open channel between them and the coordinator chain.

An implementation of each contract is as follows(A full example can be found [here](https://github.com/datachainlab/cross/blob/b66802fde58f9e7fdbd8de4ccf121a590b554b28/example/simapp/contract/reservation.go#L19)):

// Train contract on chainA
```go
func TrainReservationContractHandler(k contract.Keeper) cross.ContractHandler {
	contractHandler := contract.NewContractHandler(k, func(store sdk.KVStore) cross.State {
		return lock.NewStore(store)
	})

	contractHandler.AddRoute("train", GetTrainContract())
	return contractHandler
}

func GetTrainContract() contract.Contract {
	return contract.NewContract([]contract.Method{
		{
			Name: "reserve",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
				reserver := ctx.Signers()[0]
				seatID := contract.Int32(ctx.Args()[0])
				key := MakeSeatKey(seatID)
				if store.Has(key) {
					return nil, fmt.Errorf("seat %v is already reserved", seatID)
				} else {
					store.Set(key, reserver)
				}
				return key, nil
			},
		},
	})
}

func MakeSeatKey(id int32) []byte {
	return []byte(fmt.Sprintf("seat/%v", id))
}
```

// Hotel contract on chainB
```go
func HotelReservationContractHandler(k contract.Keeper) cross.ContractHandler {
	contractHandler := contract.NewContractHandler(k, func(store sdk.KVStore) cross.State {
		return lock.NewStore(store)
	})

	contractHandler.AddRoute("hotel", GetHotelContract())
	return contractHandler
}

func GetHotelContract() contract.Contract {
	return contract.NewContract([]contract.Method{
		{
			Name: "reserve",
			F: func(ctx contract.Context, store cross.Store) ([]byte, error) {
				reserver := ctx.Signers()[0]
				roomID := contract.Int32(ctx.Args()[0])
				key := MakeRoomKey(roomID)
				if store.Has(key) {
					return nil, fmt.Errorf("room %v is already reserved", roomID)
				} else {
					store.Set(key, reserver)
				}
				return key, nil
			},
		},
	})
}

func MakeRoomKey(id int32) []byte {
	return []byte(fmt.Sprintf("room/%v", id))
}
```

Next, we'll use query command to simulate each Contract execution.

```
$ simappcli query contract call --from <your-address> train reserve 0x00000001 --chain-id <train-chain> --save ./train.json
$ cat ./train.json
{
  "chain_id": "<train-chain>",
  "height": "134",
  "signers": [
    "<your-address>"
  ],
  "contract": "GqH1iNgKBXRyYWluEgdyZXNlcnZlGgQAAAAB",
  "ops": [
    {
      "type": "store/lock/Write",
      "value": {
        "K": "c2VhdC8x",
        "V": "QTr/9+/2vTM/n1YsPuatooggalg="
      }
    }
  ]
}
$ simappcli query contract call --from <your-address> hotel reserve 0x00000002 --chain-id <hotel-chain> --save ./hotel.json
$ cat ./hotel.json
{
  "chain_id": "<hotel-chain>",
  "height": "134",
  "signers": [
    "<your-address>"
  ],
  "contract": "GqH1iNgKBWhvdGVsEgdyZXNlcnZlGgQAAAAC",
  "ops": [
    {
      "type": "store/lock/Write",
      "value": {
        "K": "cm9vbS8y",
        "V": "QTr/9+/2vTM/n1YsPuatooggalg="
      }
    }
  ]
}
```

Using the results of these simulations, we create a cross-chain transaction and broadcast it to the coordinator chain.
```
$ simappcli tx cross create --from <your-address> --chain-id <coordinator-chain> \
    --contract ./train.json --channel ${SOURCE_CHANNEL1}:${SOURCE_PORT1} \
    --contract ./hotel.json --channel ${SOURCE_CHANNEL2}:${SOURCE_PORT2} \
    <timeout-height> <nonce> | jq -r '.data'

033E3DF0509AA0B0DD60D7FFC19F2FF5A9ABC132B96503C5C5B11D4F22B47DDB
# Set TxID
$ export TX_ID=033E3DF0509AA0B0DD60D7FFC19F2FF5A9ABC132B96503C5C5B11D4F22B47DDB
```

After this is successful, [relayer](https://github.com/cosmos/ics/tree/master/spec/ics-018-relayer-algorithms) will relay each packet. After all relays are finished, run the following command to get the status to the coordinator.

```
$ simappcli query cross coordinator ${TX_ID} | jq -r '.completed'
true
```

A script that runs these flows can be found [here](https://github.com/datachainlab/cross/blob/master/tests/test-tx.sh). Please take a look at it.

### How to execute a smart contract on single chain

In the previous chapter, we called the Train and Hotel contracts atomically. However, you would normally want to call each contract on top of single chain as well as ethereum smart contract. In this case, you can use [this command](https://github.com/datachainlab/cross/blob/aa5d7da3fb51e7c034523d12c9d0cdf49df12028/x/ibc/contract/client/cli/tx.go#L35) to invoke it as follows:
```
$ simappcli tx contract call --chain-id <train-chain> --from <your-address> train reserve 0x00000001
$ simappcli tx contract call --chain-id <hotel-chain> --from <your-address> hotel reserve 0x00000002
```
