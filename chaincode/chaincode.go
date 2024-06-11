package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	utils "decentralized_certification_chaincode/utils"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

func (contract *SmartContract) RequestIssueCertificate(ctx contractapi.TransactionContextInterface,
	request_id string, student_name string, student_id int, degree string, major string, result float32) (string, error) {

	requester, err := utils.CheckRequester(ctx)

	if requester != "" && err == nil {
		return "Not Authorized To Request Certificate", nil
	}

	encoded_requetser_identity, err := ctx.GetClientIdentity().GetID()

	if err != nil {
		return "", fmt.Errorf("failed read clinet Identity %v", err)
	}
	decoded_requetser_identity, err := base64.StdEncoding.DecodeString(encoded_requetser_identity)

	if err != nil {
		return "", fmt.Errorf("failed to decode client Identity %v", err)
	}

	request := utils.CertificateRequest{
		Request_Id: request_id, Student_Name: student_name, Student_Id: student_id, Degree: degree, Major: major, Result: result,
		Requester_Authority: string(decoded_requetser_identity), Certificate_Hash: "", Issuer_Authority: "", Is_Reqeust_Completed: false,
	}

	requestJson, err := json.Marshal(request)

	if err != nil {
		return "", fmt.Errorf("failed to json marshal request", err)
	}
	err = ctx.GetStub().PutState(request.Request_Id, requestJson)
	if err != nil {
		return "", fmt.Errorf("failed to add the request to the ledger %v", err)
	}

	return fmt.Sprintf("Submitted Request Id : ", request.Request_Id), nil

}

func (contract *SmartContract) IssueCertificate() {

}

func (contract *SmartContract) ReadCertificate() {

}

func (contract *SmartContract) GetAllTheRequests() {

}

func (contract *SmartContract) HistoryOfRequest() {

}

func (contract *SmartContract) VerifyCertificte() {

}
