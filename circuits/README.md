# Circuits

TBD

## Requirements

- Circom: `npm install -g circom`
- snarkJS: `npm install -g snarkjs`

## Generate verifier (TESTING)

1. Compile: `circom zkOnacci.circom --r1cs --wasm --sym`
2. Generate powers of that **run only the first time**: `snarkjs powersoftau new bn128 12 pot12_0000.ptau -v && snarkjs powersoftau contribute pot12_0000.ptau pot12_0001.ptau --name="First contribution" -v && snarkjs powersoftau prepare phase2 pot12_0001.ptau pot12_final.ptau -v`
3. Generate private key: `snarkjs zkey new zkOnacci.r1cs pot12_final.ptau zkOnacci_0000.zkey && snarkjs zkey contribute zkOnacci_0000.zkey zkOnacci_final.zkey --name="1st Contributor Name" -v && snarkjs zkey export verificationkey zkOnacci_final.zkey verification_key.json`
4. Generate solidity verifier `snarkjs zkey export solidityverifier zkOnacci_final.zkey verifier.sol`
5. Move the verifier to the contracts directory: `mv verifier.sol ../contracts`

## Generate verifier (PROD)

TBD
