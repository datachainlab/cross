# FAQ

## Development

### Do you have a complete example that works?

Yes, we do. Please see [here](./docs/spec/02_smart_contract.md#how-to-execute-a-smart-contract-on-cross-chain) and [this script](./tests/test-tx.sh).

## Protocol

### Are there any blocking case during an execution of atomic commit?

No, but our protocol requires some assumptions. They are here:
1. Many assumptions required by [IBC](https://github.com/cosmos/ics/tree/master/spec)
1. Any [relayers](https://github.com/cosmos/ics/tree/master/spec/ics-018-relayer-algorithms) work as expected.
