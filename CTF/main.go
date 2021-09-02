package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"

	"github.com/arnaubennassar/zkOnacci/contracts"
	"github.com/arnaubennassar/zkOnacci/contracts/zkinputs"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db/memory"
)

const nLevels = 6

func main() {
	// Set up client
	web3URL := os.Getenv("WEB3_URL")
	if web3URL == "" {
		panic("Must provide the env var WEB3_URL")
	}
	client, err := ethclient.Dial(web3URL)
	if err != nil {
		panic(err)
	}
	privKeyStr := os.Getenv("PRIVATE_KEY")
	if privKeyStr == "" {
		panic("Must provide the env var PRIVATE_KEY")
	}
	privateKey, err := crypto.HexToECDSA(privKeyStr)
	if err != nil {
		panic(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		panic(err)
	}
	scAddrHex := os.Getenv("SC_ADDR")
	scAddr := common.HexToAddress(scAddrHex)
	zkOnacci, err := contracts.NewZKOnacci(scAddr, client)
	if err != nil {
		panic(err)
	}
	// Read the SC
	callOpts := &bind.CallOpts{}
	nMintedTokens, err := zkOnacci.TokenCounter(callOpts)
	if err != nil {
		panic(err)
	}
	fmt.Println(nMintedTokens, " tokens already minted")
	// Add existing numbers of the sequence to the tree
	n := int(nMintedTokens.Int64() + 2)
	FnMinOne := 1
	FnMinTwo := 0
	merkleTree, err := merkletree.NewMerkleTree(memory.NewMemoryStorage(), nLevels)
	if err != nil {
		panic(err)
	}
	if err := merkleTree.Add(big.NewInt(0), big.NewInt(0)); err != nil {
		panic(err)
	}
	if err := merkleTree.Add(big.NewInt(1), big.NewInt(1)); err != nil {
		panic(err)
	}
	for i := int64(2); i < int64(n); i++ {
		Fn := FnMinOne + FnMinTwo
		if err := merkleTree.Add(big.NewInt(i), big.NewInt(int64(Fn))); err != nil {
			panic(err)
		}
		// Values for next iteration
		FnMinTwo = FnMinOne
		FnMinOne = Fn
	}
	// Calculate proof
	oldRoot := merkleTree.Root()
	mtpNMinOne, err := merkleTree.GenerateCircomVerifierProof(big.NewInt(int64(n-1)), nil)
	if err != nil {
		panic(err)
	}
	mtpNMinTwo, err := merkleTree.GenerateCircomVerifierProof(big.NewInt(int64(n-2)), nil)
	if err != nil {
		panic(err)
	}
	// Add Fn and get processing proof
	mtpN, err := merkleTree.AddAndGetCircomProof(big.NewInt(int64(n)), big.NewInt(int64(FnMinOne+FnMinTwo)))
	if err != nil {
		panic(err)
	}
	proofA, proofB, proofC, err := zkinputs.GenerateProof(zkinputs.ZKInput{
		Sender:           fromAddress,
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
	}, "../circuits")
	if err != nil {
		panic(err)
	}
	// Send tx
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}
	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(1500000) // in units
	auth.GasPrice = gasPrice
	tx, err := zkOnacci.CaptureTheFlag(auth, proofA, proofB, proofC, merkleTree.Root().BigInt())
	if err != nil {
		panic(err)
	}
	fmt.Println("Tx sent to the blockchain. Tx Hash:", tx.Hash())
}
