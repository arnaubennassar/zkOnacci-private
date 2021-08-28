# zkOnacci

CTF game where players will need to create a zkProof that demonstrates the knowledge of the next number of the [Fibonacci sequence](https://en.wikipedia.org/wiki/Fibonacci_number).

Flags will be deployed on Ethereum in a form of NFTs (note that in the current spec this is represented just as a list of addresses who captured the flag because this is a WIP), and will be captured every time a player submits a valid proof.

Keep in mind that the goal is not to preserve privacy in terms of who knows which number of the series, but to obfuscate the problem in a way that is hard for the CTF players to know what they have to input. It would be very obvious what to do if the SC calculated the next value of the sequence and pass it as output of the circuit.

## Requirements

- [Node 14](https://nodejs.org/en/), it's recommende to use [nvm](https://github.com/nvm-sh/nvm) to easily choose the right version (14.17.5)
- Circom: `npm install -g circom`
- snarkJS: `npm install -g snarkjs`
- [Geth tooling](https://github.com/ethereum/go-ethereum#executables)
- [Solidity compiler (solc)](https://docs.soliditylang.org/en/v0.8.6/installing-solidity.html), it's recommended to use [solc-select](https://github.com/crytic/solc-select) to easily choose the right version (0.8.6)
- [Go 1.16](https://golang.org/doc/install)

## Setup

Install dependencies: `npm i`. Note that this will run the common phase of the trusted setup, for testing purposes.

## Build

- Compile everything: `npm build`
- Compile circuits only: `npm build-circuits`
- Compile contracts only: `npm build-contracts`

Note that it's required to rebuild the contracts if the circuits are changed in order to be able to run tests. Therefore it's recommended to use always `npn run build` unless changes only affect contracts, in this case it's safe and faster to use `npm run build-contracts`

## Test

Run tests: `npm test` or `cd contracts && go test -v`

## Architecture

In order to obfuscate the solution (a valid proof that demonstrates the knowledge of the next number of the fibonacci sequence), the problem will be represented as a MT of fixed size. This MT will be built by adding the nth value of the fibonacci sequence to the nth leafs:

| Leaf index | Value                 |
| ---------- | --------------------- |
| 0          | 0                     |
| 1          | 1                     |
| 2          | 1                     |
| 3          | 2                     |
| 4          | 3                     |
| ...        | ...                   |
| N          | Leaf[N-1] + Leaf[N-2] |

### Circuit

#### Private inputs

- n: Position in the sequence
- Fn: value of the n element of the fibonacci sequence
- FnMinOne: value of the n-1 element of the fibonacci sequence
- FnMinTwo: value of the n-2 element of the fibonacci sequence
- epMinOne: MTP that the value of FnMinOne exists on the n-1 position of the tree
- epMinTwo: MTP that the value of FnMinTwo exists on the n-2 position of the tree
- epN: MTP that the value of Fn exists on the n position of the tree

#### Output

- currentRoot: Hash of the root of the MT BEFORE adding Fn into the n leaf
- nextRoot: Hash of the root of the MT AFTER adding Fn into the n leaf

#### Constrains

- Fn = FnMinOne + FnMinTwo
- The two previous numbers existed in the tree before the call (epMinOne, epMinTwo are valid against currentRoot)
- The new number exists in the tree after the call (epN is valid against nextRoot)

### Smart contract

The SC will be deployed with an initial value of the `currentRoot` that represent the tree when it has the two first numbers (otherwise some constrains would always fail in the first iteration)

Main interface will be `captureTheFlag(nextRoot, privateInputs)`. This function will:

- Will check that `nextRoot != currentRoot`
- Verify the zkProof: `verify(privateInputs, currentRoot, nextRoot)` (note that current root is stored in SC)
- If proof succeeds:
  - `currentRoot = nextRoot`
  - Add the address of the caller to a winner list (note that this will help players know which is the current number of the sequence)
    - TODO: update with NFT minting process

---

Ideas

- Create different levels of difficulty: on each level of difficulty a different NFT will be used. Once all the NFTs of the current level are minted, a new hint will be published
  - 1st hint (published on first release): etherscan link to the SC
  - 2nd hint (published after first X flags are captured): example of private inputs (maybe last used)
  - 3rd hint (published after first Y flags are captured): explanation of the private inputs
  - TODO: define more hints
  - Last hint: publish all the code, docs and a tutorial (working solution)
- Store off chain data (circuit code, proofing artifacts, ...) on IPFS, write the IPFS addr on a comment in the SC
