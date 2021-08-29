package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os/exec"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-circom-prover-verifier/parsers"
	"github.com/iden3/go-circom-prover-verifier/types"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db/memory"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

type testingEnv struct {
	auth       *bind.TransactOpts
	blockchain *backends.SimulatedBackend
	scAddr     common.Address
	zkOnacci   *ZKOnacci
	client     *backends.SimulatedBackend
	provingKey *types.Pk
}

func newTestingEnv() (testingEnv, error) {
	balance := big.NewInt(0)
	balance.SetString("10000000000000000000000000", 10) // 10 ETH in wei
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return testingEnv{}, err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	if err != nil {
		return testingEnv{}, err
	}

	auth.GasLimit = 99999999999
	address := auth.From
	genesisAlloc := map[common.Address]core.GenesisAccount{
		address: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(999999999999999999)
	client := backends.NewSimulatedBackend(genesisAlloc, blockGasLimit)

	// Deploy contracts
	verifierAddr, _, _, err := DeployVerifier(
		auth,
		client,
	)
	if err != nil {
		return testingEnv{}, err
	}
	scAddr, _, zkOnacci, err := DeployZKOnacci(
		auth,
		client,
		verifierAddr,
	)
	if err != nil {
		return testingEnv{}, err
	}
	client.Commit()
	return testingEnv{
		auth:       auth,
		blockchain: client,
		scAddr:     scAddr,
		zkOnacci:   zkOnacci,
		client:     client,
	}, nil
}

const nLevels = 6

type zkInput struct {
	Sender           common.Address     `json:"senderInput"`
	Root             *merkletree.Hash   `json:"stateRoot"`
	N                int                `json:"n"`
	Fn               int                `json:"Fn"`
	SiblingsFn       []*merkletree.Hash `json:"siblingsFn"`
	OldKeyFn         *merkletree.Hash   `json:"oldKeyFn"`
	OldValueFn       *merkletree.Hash   `json:"oldValueFn"`
	IsOld0Fn         bool               `json:"isOld0Fn"`
	FnMinOne         int                `json:"FnMinOne"`
	SiblingsFnMinOne []*merkletree.Hash `json:"siblingsFnMinOne"`
	FnMinTwo         int                `json:"FnMinTwo"`
	SiblingsFnMinTwo []*merkletree.Hash `json:"siblingsFnMinTwo"`
}

func TestMintNFT(t *testing.T) {
	// Set up testing environment
	testEnv, err := newTestingEnv()
	require.NoError(t, err)
	callOpts := &bind.CallOpts{}
	// Get tokenURIs by tier
	nTiers, err := testEnv.zkOnacci.NTiers(callOpts)
	require.NoError(t, err)
	baseURI, err := testEnv.zkOnacci.BaseURI(callOpts)
	require.NoError(t, err)
	tokenTiers := []uint16{}
	tokenURIs := []string{}
	for i := 0; i < int(nTiers); i++ {
		// Tier
		iTier, err := testEnv.zkOnacci.TokenTiers(callOpts, big.NewInt(int64(i)))
		require.NoError(t, err)
		tokenTiers = append(tokenTiers, iTier)
		// URI
		iURI, err := testEnv.zkOnacci.TokenURIs(callOpts, big.NewInt(int64(i)))
		require.NoError(t, err)
		tokenURIs = append(tokenURIs, baseURI+iURI)
	}
	// Calculate initial state
	merkleTree, err := merkletree.NewMerkleTree(memory.NewMemoryStorage(), nLevels)
	require.NoError(t, err)
	require.NoError(t, merkleTree.Add(big.NewInt(0), big.NewInt(0)))
	require.NoError(t, merkleTree.Add(big.NewInt(1), big.NewInt(1)))
	// Mint all tokens +1 (to test that the supply is limited as expected)
	var n uint16 = 2
	maxTier := tokenTiers[len(tokenTiers)-1]
	FnMinOne := 1
	FnMinTwo := 0
	for n < maxTier+2 {
		fmt.Printf("Minting NFT #%d, nMinusOne = %d, nMinusTwo = %d, nFib = %d\n", n, FnMinOne, FnMinTwo, FnMinOne+FnMinTwo)
		// Generate proof
		// Existence proofs for Fn-1 and Fn-2 BEFORE processing Fn
		oldRoot := merkleTree.Root()
		mtpNMinOne, err := merkleTree.GenerateCircomVerifierProof(big.NewInt(int64(n-1)), nil)
		require.NoError(t, err)
		mtpNMinTwo, err := merkleTree.GenerateCircomVerifierProof(big.NewInt(int64(n-2)), nil)
		require.NoError(t, err)
		// Add Fn and get processing proof
		mtpN, err := merkleTree.AddAndGetCircomProof(big.NewInt(int64(n)), big.NewInt(int64(FnMinOne+FnMinTwo)))
		require.NoError(t, err)
		proofA, proofB, proofC, _, err := generateProof(zkInput{
			Sender:           testEnv.auth.From,
			Root:             oldRoot,
			N:                int(n),
			Fn:               FnMinOne + FnMinTwo,
			SiblingsFn:       mtpN.Siblings,
			OldKeyFn:         mtpN.OldKey,
			OldValueFn:       mtpN.OldValue,
			IsOld0Fn:         mtpN.IsOld0,
			FnMinOne:         FnMinOne,
			SiblingsFnMinOne: mtpNMinOne.Siblings,
			FnMinTwo:         FnMinTwo,
			SiblingsFnMinTwo: mtpNMinTwo.Siblings,
		})
		require.NoError(t, err)
		// Capture the flag (mint token): send tx
		nonce, err := testEnv.client.NonceAt(context.Background(), testEnv.auth.From, nil)
		require.NoError(t, err)
		testEnv.auth.Nonce = big.NewInt(int64(nonce))
		tx, err := testEnv.zkOnacci.CaptureTheFlag(
			testEnv.auth,
			proofA,
			proofB,
			proofC,
			merkleTree.Root().BigInt(),
		)
		require.NoError(t, err)
		testEnv.client.Commit()
		txReceipt, err := testEnv.client.TransactionReceipt(context.Background(), tx.Hash())
		require.NoError(t, err)
		if n-2 < maxTier+1 { // New token should have been minted
			// No error on tx
			require.Equal(t, uint64(1), txReceipt.Status)
			// Assert owner
			owner, err := testEnv.zkOnacci.OwnerOf(callOpts, big.NewInt(int64(n-2)))
			require.NoError(t, err)
			assert.Equal(t, testEnv.auth.From, owner)
			// Assert tokenURI
			uri, err := testEnv.zkOnacci.TokenURI(callOpts, big.NewInt(int64(n-2)))
			require.NoError(t, err)
			assert.Equal(t, expectedURI(n-2, tokenTiers, tokenURIs), uri)
			// Values for next iteration
			n++
			tmpMinusOne := FnMinOne
			FnMinOne += FnMinTwo
			FnMinTwo = tmpMinusOne
		} else { // All tokens already minted
			assert.Equal(t, uint64(0), txReceipt.Status)
			// TODO: should receive "ZKOnacci::captureTheFlag: ALL_TOKENS_MINTED"
			break
		}
	}
}

func generateProof(input zkInput) (
	proofA [2]*big.Int,
	proofB [2][2]*big.Int,
	proofC [2]*big.Int,
	output [1]*big.Int,
	err error,
) {
	inputJson, err := json.Marshal(input)
	if err != nil {
		return
	}
	if err = ioutil.WriteFile(`../circuits/input.json`, inputJson, 0777); err != nil {
		return
	}
	// Calculate witness
	var cmdOut []byte
	if cmdOut, err = exec.Command(
		`snarkjs`, `wtns`, `calculate`,
		`../circuits/zkOnacci.wasm`, `../circuits/input.json`, `../circuits/witness.wtns`,
	).Output(); err != nil {
		fmt.Println(string(cmdOut))
		return
	}
	// Generate proof
	if cmdOut, err = exec.Command(`snarkjs`, `groth16`, `prove`,
		`../circuits/zkOnacci_final.zkey`, `../circuits/witness.wtns`,
		`../circuits/proof.json`, `../circuits/public.json`,
	).Output(); err != nil {
		fmt.Println(string(cmdOut))
		return
	}
	proofJSON, err := ioutil.ReadFile("../circuits/proof.json")
	if err != nil {
		return
	}
	proof, err := parsers.ParseProof(proofJSON)
	proofSC := parsers.ProofToSmartContractFormat(proof)
	a0, _ := big.NewInt(0).SetString(proofSC.A[0], 10)
	a1, _ := big.NewInt(0).SetString(proofSC.A[1], 10)
	b00, _ := big.NewInt(0).SetString(proofSC.B[0][0], 10)
	b01, _ := big.NewInt(0).SetString(proofSC.B[0][1], 10)
	b10, _ := big.NewInt(0).SetString(proofSC.B[1][0], 10)
	b11, _ := big.NewInt(0).SetString(proofSC.B[1][1], 10)
	c0, _ := big.NewInt(0).SetString(proofSC.C[0], 10)
	c1, _ := big.NewInt(0).SetString(proofSC.C[1], 10)
	return [2]*big.Int{a0, a1},
		[2][2]*big.Int{{b00, b01}, {b10, b11}},
		[2]*big.Int{c0, c1},
		[1]*big.Int{},
		nil
}

func expectedURI(id uint16, tokenTiers []uint16, tokenURIs []string) string {
	var tier int
	for tier < len(tokenTiers)-1 && id > tokenTiers[tier] {
		tier++
	}
	return tokenURIs[tier]
}
