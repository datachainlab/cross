# Cross Framework

<div align="center">
  <img width="400" alt="Cross Framework" src="./docs/images/cross_logo.png" />
</div>

<div align="center">
  <a href="https://pkg.go.dev/github.com/datachainlab/cross?tab=doc">
    <img alt="godoc" src="https://godoc.org/github.com/datachainlab/cross?status.svg" />
  </a>
  <img alt="gh-action" src="https://github.com/datachainlab/cross/workflows/Test/badge.svg" />
</div>

Cross is a framework that enables distributed transactions over multiple blockchains. It is implemented as [Cosmos module](https://github.com/cosmos/cosmos-sdk).

Cross provides several key features:

- **Cross-chain transaction support** - supports [Atomic Commit Protocol](https://datachainlab.github.io/cross-docs/architecture/atomic-commit-protocol) beteen multiple blockchains.
- **General application support** - provides a framework to enable the support of "general" application as smart contract ("general" application means Account model like Ethereum, not UTXO model). Using this framework, smart contract developers are not forced to implement Atomic Commit and locking protocol at each contract development.
- **Built on [IBC](https://github.com/cosmos/ibc)** - supports above features on networks where membership changes dynamically

## Documents

Please see [here](https://github.com/datachainlab/public-docs/blob/master/cross/concepts_ja.md).

## Related Projects

- https://github.com/datachainlab/cross-solidity: A solidity implementation for Ethereum
- https://github.com/datachainlab/cross-cdt: A store that supports efficient concurrent operations
- https://github.com/hyperledger-labs/yui-docs: YUI enables interoperability between multiple heterogeneous ledgers

## Motivation

It is difficult to atomically execute general smart contract on multiple networks. One such example is [Train-And-Hotel problem](https://github.com/ethereum/wiki/wiki/Sharding-FAQ#what-is-the-train-and-hotel-problem). If we can convert Train and Hotel reservation rights into NFT that can be moved to any chain using Two-way peg method such as [ics-020](https://github.com/cosmos/ibc/tree/master/spec/app/ics-020-fungible-token-transfer), it may be possible to solve this problem simply by doing atomicswap on a single chain. However, if each Token's metadata (e.g. a whitelist of owner) depends on other states of its origin chain and common state is referenced by other contract states, it is difficult to move between chains.

To solve such problem, we need to be able to execute All or Nothing reservation contracts that exist in two different chains, rather than pegging to single blockchain. This is similar to Atomic commit protocol for distributed systems. To achieve this, each contract's state machine needs to lock the state required for "commit" at "prepare" phase. But it is not safe to enforce these requirements on each contract developers. Therefore, we decided to implement Atomic commit protocol in compliance with ics-004 and create a datastore that transparently meets the required locking protocol.

## Maintainers

- [Jun Kimura](https://github.com/bluele)
- [Ryo Sato](https://github.com/3100)

## Contact

Please contact us [here](https://form.run/@datachain-contact-en).
