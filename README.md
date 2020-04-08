# Cross

![](https://github.com/datachainlab/cross/workflows/Test/badge.svg)

Cross is a framework for Cross-chain transaction. It is implemented as [Cosmos module](https://github.com/cosmos/cosmos-sdk).

Cross provides several key features:

- **Cross-chain transaction support** - Supports the transaction that can support an atomic execution on different blockchains. We call such a transaction "Cross-chain transaction".
- **General application support** - Provides a framework to enable the support of "general" applications as smart contracts. ("general" application refers to something like Ethereum's smart contract, not something like the UTXO model.) With Cross framework, smart contract developers are not forced to implement Atomic commit and locking protocol at each contract develop.
- **Compliant with [ics-004](https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics)** - Supports above features on networks where membership changes dynamically

## Motivation

It is difficult to atomically execute general smart contract on multiple networks. One such example is Train-Hotel problem. If we can convert Train and Hotel reservation rights into NFT that can be moved to any chain using Two-way peg method such as [ics-020](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer), it may be possible to solve this problem simply by doing atomicswap on a single chain. However, if each Token's metadata (e.g. a whitelist of owner) depends on other states of its origin chain and common state is referenced by other contract states, it is difficult to move between chains.

To solve such problem, we need to be able to execute ALL or Nothing reservation contracts that exist in two different chains, rather than pegging to single blockchain. This is similar to Atomic commit protocol for distributed systems. To achieve this, each contract state machine must be able to support "Prepared" state,  so it must be able to lock the state required for "Commit". But it is not safe to enforce these requirements on each contract developers. So we decided to create a framework that supports the Atomic commit protocol and a datastore that transparently meets the required locking protocol.

## Documents

For specs and documents, see [here](./docs/spec)

## Future works

Currently, Smart contract layer supports only Golang, but there is plan to support EVM in future. This will bring not only scaling, but also interoperability to existing smart contract that is developed as Ethereum contracts.

## Q&A

- Q. Are there any blocking case during an execution of Atomic commit?
- A. No. But our protocol requires some assumptions. They are here:
    1. Many assumptions required by [IBC](https://github.com/cosmos/ics/tree/master/spec)
    1. Any [relayers](https://github.com/cosmos/ics/tree/master/spec/ics-018-relayer-algorithms) work as expected.

## Maintainers

- [Jun Kimura](https://github.com/bluele)
