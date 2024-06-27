package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	utils "decentralized_certification_chaincode/utils"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	certKey             = "cert_to_request_id"
	certhashKey         = "cert_hash_to_request_id"
	duplicateRequestKey = "dupliReqKey"
)

type TrackingIdResponse struct {
	TrackingId string `json:tracking_id`
}

type SmartContract struct {
	contractapi.Contract
}

func (contract *SmartContract) RequestIssueCertificate(ctx contractapi.TransactionContextInterface,
	tracking_id string, student_name string, student_id int, degree string, major string, result float32) (*TrackingIdResponse, error) {

	requester, err := utils.CheckRequester(ctx)

	if requester != "" && err == nil {
		request_response := TrackingIdResponse{
			TrackingId: "Not Authorized Submit Request",
		}

		return &request_response, nil
	}

	isExits, err := contract.IsRequestExist(ctx, tracking_id)

	if err != nil {
		return nil, err
	}

	if (isExits == true) && (err == nil) {
		request_response := TrackingIdResponse{
			TrackingId: fmt.Sprintf("A Request  Already Exists With the Id : %s", tracking_id),
		}
		return &request_response, err
	}

	compositKey, err := ctx.GetStub().CreateCompositeKey(duplicateRequestKey, []string{student_name, strconv.Itoa(student_id), degree, major})
	if err != nil {
		return nil, fmt.Errorf("failed create composit key %w", err)
	}

	isrequestExits, err := ctx.GetStub().GetState(compositKey)

	if err != nil {

		return nil, fmt.Errorf("failed read from the ledger")
	}

	if isrequestExits != nil {
		request_response := TrackingIdResponse{
			TrackingId: fmt.Sprintf("A Request  Already Exists With the Name : %s , Student_Id : %s Degree: %s , Major: %s",
				student_name, strconv.Itoa(student_id), degree, major),
		}
		return &request_response, nil
	}

	encodedRequetserIdentity, err := ctx.GetClientIdentity().GetID()

	if err != nil {
		return nil, fmt.Errorf("failed read clinet Identity %w", err)
	}
	decodedRequetserIdentity, err := base64.StdEncoding.DecodeString(encodedRequetserIdentity)

	if err != nil {
		return nil, fmt.Errorf("failed to decode client Identity %w", err)
	}

	request := utils.CertificateRequest{
		Tracking_Id: tracking_id, Student_Name: student_name, Student_Id: student_id, Degree: degree, Major: major, Result: result,
		Requester_Authority: string(decodedRequetserIdentity), Certificate_Hash: "", Issuer_Authority: "",
		Is_Reqeust_Completed: false,
		Certificate_Id:       0000,
	}

	requestJson, err := json.Marshal(request)

	if err != nil {
		return nil, fmt.Errorf("failed to json marshal request %w", err)
	}
	err = ctx.GetStub().PutState(request.Tracking_Id, requestJson)

	if err != nil {
		return nil, fmt.Errorf("failed to add the request to the ledger %w", err)
	}

	err = ctx.GetStub().PutState(compositKey, []byte{0x00})
	if err != nil {
		return nil, fmt.Errorf("failed to add the composit to the ledger %w", err)
	}

	request_response := TrackingIdResponse{
		TrackingId: request.Tracking_Id,
	}

	return &request_response, nil

}

func (contract *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface,
	tracking_id string, certitficate_hash string, certificate_id int) (int, error) {

	Issuer, err := utils.IsIssuer(ctx)

	if err != nil {
		return 0000, err
	}

	if !Issuer {
		return 0000, fmt.Errorf("Not Authorized To Issue Certificate")
	}

	exits, err := contract.IsRequestExist(ctx, tracking_id)

	if err != nil {
		return 0000, fmt.Errorf("%w", err)
	}

	if !exits {
		return 0000, fmt.Errorf("Request does not exists with id  : %w", tracking_id)
	}

	request, err := contract.ReadRequest(ctx, tracking_id)
	// Double Certificate Creation
	if request.Is_Reqeust_Completed {
		return 0000, nil
	}

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

	compositKey, err := ctx.GetStub().CreateCompositeKey(certKey, []string{strconv.Itoa(request.Certificate_Id), request.Tracking_Id})

	if err != nil {
		return 0000, fmt.Errorf("failed to create composite key: %w", err)
	}

	err = ctx.GetStub().PutState(compositKey, []byte{0x00})

	if err != nil {
		return 0000, fmt.Errorf("failed to add  compositeKey to ledger %w", err)
	}

	// CreateComposite Key For The Certificate Hash

	compositKeyForCertHash, err := ctx.GetStub().CreateCompositeKey(certhashKey, []string{request.Certificate_Hash, request.Tracking_Id})

	if err != nil {
		return 0000, fmt.Errorf("failed to create composite key for hash: %w", err)
	}

	err = ctx.GetStub().PutState(compositKeyForCertHash, []byte{0x00})

	if err != nil {
		return 0000, fmt.Errorf("failed to add  compositeKeyforcerthash to ledger %w", err)
	}

	err = ctx.GetStub().PutState(tracking_id, jsonRequest)

	if err != nil {
		return 0000, fmt.Errorf("failed to add  request to ledger %w", err)
	}

	return request.Certificate_Id, nil
}

func (contract *SmartContract) ReadRequest(ctx contractapi.TransactionContextInterface, tracking_id string) (*utils.CertificateRequest, error) {

	exits, err := contract.IsRequestExist(ctx, tracking_id)

	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if !exits {
		return nil, fmt.Errorf("Request does not exists with id  : %w", tracking_id)
	}

	jsonRequest, err := ctx.GetStub().GetState(tracking_id)

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

func (contract *SmartContract) HistoryOfRequest(ctx contractapi.TransactionContextInterface, tracking_id string) ([]*utils.CertificateRequest, error) {

	exits, err := contract.IsRequestExist(ctx, tracking_id)

	if err != nil {
		return nil, err
	}
	if !exits {
		return nil, fmt.Errorf("no request found for the provide request id %w", tracking_id)
	}

	RequestHistoryIterator, err := ctx.GetStub().GetHistoryForKey(tracking_id)

	if err != nil {
		return nil, err
	}

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

func (contract *SmartContract) ReadCertificateByCertificateId(ctx contractapi.TransactionContextInterface, certificate_id int) (*utils.CertificateRequest, error) {

	resultIterartor, err := ctx.GetStub().GetStateByPartialCompositeKey(certKey, []string{strconv.Itoa(certificate_id)})

	if err != nil {
		return nil, err
	}

	defer resultIterartor.Close()

	if !resultIterartor.HasNext() {
		return nil, fmt.Errorf("not certificate found for the id %d", certificate_id)
	}

	queryResponse, err := resultIterartor.Next()

	if err != nil {
		return nil, err
	}

	_, compositeKey, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)

	if err != nil {
		return nil, err
	}

	tracking_id := compositeKey[1]

	request, err := contract.ReadRequest(ctx, tracking_id)

	if err != nil {
		return nil, err
	}

	return request, nil

}
func (contract *SmartContract) VerifyCertificateByCertificateHash(ctx contractapi.TransactionContextInterface, cert_hash string) (*utils.CertificateRequest, error) {
	resultIterartor, err := ctx.GetStub().GetStateByPartialCompositeKey(certhashKey, []string{cert_hash})

	if err != nil {
		return nil, err
	}

	defer resultIterartor.Close()

	if !resultIterartor.HasNext() {
		return nil, fmt.Errorf("No Certificate found for the hash : %w", cert_hash)
	}

	queryResponse, err := resultIterartor.Next()

	if err != nil {
		return nil, err
	}

	_, compositKeyForHash, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)

	tracking_id := compositKeyForHash[1]

	request, err := contract.ReadRequest(ctx, tracking_id)

	if err != nil {
		return nil, err
	}

	return request, nil

}

func (contract *SmartContract) IsRequestExist(ctx contractapi.TransactionContextInterface, tracking_id string) (bool, error) {

	jsonRequest, err := ctx.GetStub().GetState(tracking_id)

	if err != nil {
		return false, fmt.Errorf("failed to read asset %w", err)
	}

	return jsonRequest != nil, nil
}
