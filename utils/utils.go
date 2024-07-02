package utils

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type CertificateRequest struct {
	Tracking_Id          string  `json:tracking_id`
	Student_Name         string  `json:student_name`
	Student_Id           int     `json:student_id`
	Student_Email        string  `json:student_email`
	Degree               string  `json:degree`
	Major                string  `json:major`
	Result               float32 `json:result`
	Requester_Authority  string  `json:requester_authority`
	Certificate_Hash     string  `json:certificate_hash`
	Is_Reqeust_Completed bool    `json:is_request_completed`
	Issuer_Authority     string  `json:issuer_authority`
	Certificate_Id       int     `json:certificate_id`
}

func CheckRequester(ctx contractapi.TransactionContextInterface) (string, error) {
	requester_msp_id, err := ctx.GetClientIdentity().GetMSPID()

	if err != nil {
		return "", fmt.Errorf("failed read clinet Identity %v", err)
	}
	if requester_msp_id == "Org1MSP" {
		return requester_msp_id, nil
	}

	return "", nil

}

func IsIssuer(ctx contractapi.TransactionContextInterface) (bool, error) {

	issuerMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return false, err
	}

	return issuerMSP == "Org1MSP", nil

}
