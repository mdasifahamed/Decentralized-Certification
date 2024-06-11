package main

import (
	"decentralized_certification_chaincode/chaincode"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {

	chaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})

	if err != nil {
		log.Panicf("failed to initiate chaincode %v", err)
	}

	err = chaincode.Start()

	if err != nil {
		log.Panicf("failed to start chaincode")
	}
}
