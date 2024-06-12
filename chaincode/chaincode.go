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

	encodedRequetserIdentity, err := ctx.GetClientIdentity().GetID()

	if err != nil {
		return "", fmt.Errorf("failed read clinet Identity %w", err)
	}
	decodedRequetserIdentity, err := base64.StdEncoding.DecodeString(encodedRequetserIdentity)

	if err != nil {
		return "", fmt.Errorf("failed to decode client Identity %w", err)
	}

	request := utils.CertificateRequest{
		Request_Id: request_id, Student_Name: student_name, Student_Id: student_id, Degree: degree, Major: major, Result: result,
		Requester_Authority: string(decodedRequetserIdentity), Certificate_Hash: "", Issuer_Authority: "",
		Is_Reqeust_Completed: false,
		Certificate_Id:       0000,
	}

	requestJson, err := json.Marshal(request)

	if err != nil {
		return "", fmt.Errorf("failed to json marshal request %w", err)
	}
	err = ctx.GetStub().PutState(request.Request_Id, requestJson)
	if err != nil {
		return "", fmt.Errorf("failed to add the request to the ledger %w", err)
	}

	return fmt.Sprintf("Submitted Request Id : ", request.Request_Id), nil

}

func (contract *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface,
	request_id string, certitficate_hash string, certificate_id int) (int, error) {

	Issuer, err := utils.IsIssuer(ctx)

	if err != nil {
		return 0000, err
	}

	if !Issuer {
		return 0000, fmt.Errorf("Not Authorized To Issue Certificate")
	}

	exits, err := contract.IsRequestExist(ctx, request_id)

	if err != nil {
		return 0000, fmt.Errorf("%w", err)
	}

	if !exits {
		return 0000, fmt.Errorf("Request does not exists with id  : %w", request_id)
	}

	request, err := contract.ReadRequest(ctx, request_id)

	if err != nil {
		return 0000, fmt.Errorf("%w", err)
	}

	encodeIssuerIdenity, err := ctx.GetClientIdentity().GetID()

	if err != nil {

		return 0000, fmt.Errorf("failed read clinet Identity %w", err)
	}

	decodedIssuerIdentity, err := base64.StdEncoding.DecodeString(encodeIssuerIdenity)
	if err != nil {

		return 0000, fmt.Errorf("failed read clinet Identity %w", err)
	}
	request.Issuer_Authority = string(decodedIssuerIdentity)
	request.Certificate_Id = certificate_id
	request.Certificate_Hash = certitficate_hash
	request.Is_Reqeust_Completed = true

	jsonRequest, err := json.Marshal(request)

	if err != nil {
		return 0000, fmt.Errorf("failed to marshal request %w", err)
	}

	err = ctx.GetStub().PutState(request_id, jsonRequest)

	if err != nil {
		return 0000, fmt.Errorf("failed to add  request to ledger %w", err)
	}

	return request.Certificate_Id, nil
}

func (contract *SmartContract) ReadRequest(ctx contractapi.TransactionContextInterface, request_id string) (*utils.CertificateRequest, error) {

	jsonRequest, err := ctx.GetStub().GetState(request_id)

	if err != nil {
		return nil, fmt.Errorf("failed to read request from legder %w", err)
	}

	var request utils.CertificateRequest

	err = json.Unmarshal(jsonRequest, &request)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request %w", err)
	}
	return &request, nil
}

func (contract *SmartContract) GetAllTheRequests(ctx contractapi.TransactionContextInterface) ([]*utils.CertificateRequest, error) {

	RequestQueryIterartor, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, fmt.Errorf("failed to get state %w", err)
	}

	defer RequestQueryIterartor.Close()

	var requests []*utils.CertificateRequest

	for RequestQueryIterartor.HasNext() {
		queryResoonse, err := RequestQueryIterartor.Next()

		if err != nil {
			return nil, err
		}

		var request utils.CertificateRequest

		err = json.Unmarshal(queryResoonse.Value, &request)

		if err != nil {
			return nil, err
		}

		requests = append(requests, &request)
	}

	return requests, nil
}

func (contract *SmartContract) HistoryOfRequest(ctx contractapi.TransactionContextInterface, request_id string) ([]*utils.CertificateRequest, error) {

	exits, err := contract.IsRequestExist(ctx, request_id)

	if err != nil {
		return nil, err
	}
	if !exits {
		return nil, fmt.Errorf("no request found for the provide request id %w", request_id)
	}

	RequestHistoryIterator, err := ctx.GetStub().GetHistoryForKey(request_id)

	var requestHistories []*utils.CertificateRequest
	for RequestHistoryIterator.HasNext() {
		queryResponse, err := RequestHistoryIterator.Next()
		if err != nil {
			return nil, err
		}

		var request utils.CertificateRequest
		err = json.Unmarshal(queryResponse.Value, &request)

		if err != nil {
			return nil, err
		}

		requestHistories = append(requestHistories, &request)
	}

	return requestHistories, nil

}

func (contract *SmartContract) ReadCertificate() {

}
func (contract *SmartContract) VerifyCertificte() {

}

func (contract *SmartContract) IsRequestExist(ctx contractapi.TransactionContextInterface, reqquest_id string) (bool, error) {
	jsonRequest, err := ctx.GetStub().GetState(reqquest_id)

	if err != nil {
		return false, fmt.Errorf("failed to read asset %w", err)
	}

	return jsonRequest != nil, nil
}
