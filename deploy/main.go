package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/arnaubennassar/zkOnacci/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
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

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(1500000) // in units
	auth.GasPrice = gasPrice

	// Deploy verifier
	verifierAddr, tx, _, err := contracts.DeployVerifier(
		auth,
		client,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("verifier deployment tx sent:")
	fmt.Println(verifierAddr.Hex())
	fmt.Println(tx.Hash().Hex())
	for {
		time.Sleep(time.Second * 15)
		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			panic(err)
		}
		if receipt.Status == 1 {
			fmt.Println("verifier deployed successfully")
			break
		} else {
			// Wait before checking again if tx has already been forged
			fmt.Println("tx not mined yet")
		}
	}

	// Deploy zkOnacci
	gasPrice, err = client.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}
	auth.GasPrice = gasPrice
	auth.GasLimit = uint64(2000000) // in units
	auth.Nonce = big.NewInt(int64(nonce + 1))
	scAddr, tx, _, err := contracts.DeployZKOnacci(
		auth,
		client,
		verifierAddr,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("zkOnacci deployment tx sent:")
	fmt.Println(scAddr.Hex()) // 0x09aC8A7DD8D00C049af7C6117ECa9E3aeD8a43Ac
	fmt.Println(tx.Hash().Hex())
	for {
		time.Sleep(time.Second * 15)
		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			panic(err)
		}
		if receipt.Status == 1 {
			fmt.Println("zkOnacci deployed successfully")
			break
		} else {
			// Wait before checking again if tx has already been forged
			fmt.Println("tx not mined yet")
		}
	}
}
