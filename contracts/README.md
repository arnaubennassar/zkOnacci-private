# Contracts

TBD

## Requirements

- [Geth tooling](https://github.com/ethereum/go-ethereum#executables)
- [Solidity compiler (solc)](https://docs.soliditylang.org/en/v0.8.6/installing-solidity.html)
- [Go](https://golang.org/doc/install)

## Build

1. Generate the verifier by following [this instructions](../circuits/README.md)
2. Compile the contracts: `abigen -sol zkonacci.sol -pkg contracts -out zkonacci.go`

You may need to manually change the solidity version of `verifier.sol`

## Test

`go test ./...`

## Deploy

TBD
