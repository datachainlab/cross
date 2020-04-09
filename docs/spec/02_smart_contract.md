# Smart contract

This section explains how smart contract works and how it is developed using Cross.

## The flow of smart contract execution

Let's take a look at the flow of executing smart contract; check out [this image](../images/packet-flow.png) to see the overall flow of Cross-chain transaction.

Each contract state transition is performed in two phases, "Prepare" and "Commit".

### Prepare phase

This phase is initiated by receiving a PreparePacket from the coordinator. 

The participant who receives a packet executes a smart contract according to contract call information written to PacketData. If successful, take the operations out of Store and make sure they match the operations in PacketData. If any of these processes fail, PreparedResultPacket is sent to the coordinator with "Failed" set to prepare status without writing anything to Store. If all processes are successful, operations are saved in Store and get a Lock for each key. When writing operations are performed on the State, the exclusive lock is obtained, and when read operations are performed, the shared lock is obtained. Finally, PreparedResultPacket is sent to the coordinator with "OK" set to prepare status.

### Commit phase

This phase is initiated by receiving a CommitPacket from the coordinator.

If the status of CommitPacket is "Commit", it first applies(commits) the operation saved in Prepare to state, then releases the acquired Lock, and finally returns the Ack packet to the coordinator.
If the status of packet is "Abort", it first discards the operation saved in Prepare phase, then releases the acquired Lock, and finally returns the Ack packet to the coordinator.


That's all the execution flow for the Contract.

Let's create a simple smart contract in the next section and check out its execution flow and internal state!

## How to execute a smart contract on cross-chain

TODO: modify this section

Currently, a complete example can be found [here](https://github.com/datachainlab/cross/blob/master/tests/test-tx.sh).

