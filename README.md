# Decentralized Certification 

This chaincode will be used across consortiums to issue digital certificates. The consortium will be formed with a parent institution that has the sole right to issue certificates. The child institutions will send requests to the smart contract which will stored on the ledger. Based on these requests, the parent institution will issue the certificates.

**Note**: Chaincode and SmartContract are interchangble. 

## Folders And Files
There are two mian folder **utils** and **chaincode** and one main file **smartcontract** and rest are the dependencies.

**utils:** It contains the all the utility functions and data type for the smartcontract.The `utils.go` file has the the following  structure which will be used on the chaincode.
```javascript
    type CertificateRequest struct {
        Request_Id           string  `json:request_id` // it will be dynamically created and will be unique
        Student_Name         string  `json:student_name` // it will come from the child institutation
        Student_Id           int     `json:student_id` // it will come from the child institutation
        Degree               string  `json:degree` // it will come from the child institutation
        Major                string  `json:major` // it will come from the child institutation
        Result               float32 `json:result` // it will come from the child institutation
        Requester_Authority  string  `json:requester_authority` // it will be fetched  at time of submitting request to the blockchain from the child organization. It is the identity of the  child organzition.
        Certificate_Hash     string  `json:certificate_hash` // it the ipfs hash of certificate file which will generatted and submited by the issuer authority here which parent organization. it will be unique for very certificate.
        Is_Reqeust_Completed bool    `json:is_request_completed` // It defines the state whhre the request is completed resulting that  the certificate is genrateted if false then the certificate is not generated.
        Issuer_Authority     string  `json:issuer_authority` // it will also be fetched at time of issueing the certificate from the parent organizations when the preant organization will issues certificate.
        Certificate_Id       int     `json:certificate_id` // it will be also dynamically created and will be unique
    }
```

It also has a helper function `CheckRequester()` which takes transatcion context interface to check is the is transaction iniater the parent organization or not which will used in the chaincode to determine transaction initiater identity. Another function is here `IsIssuer()` which used to check is the certificate is the permitted peer.



**chiancode:** This folder contains `chaincode.go` which is actual `SmartContract`. Where all he  bussiness logic is implemented.

### Defination of the `RequestIssueCertificate()`. 
It submits request to the parent institution for issuing certificate of behalf of the student.It Takes `request_id string`, `student_name string`,` student_id int`, `degree string`, `major string`, `result float32` as parameters all these parameter are stored on the chain and those come from the child institutions.
`request_id string` act as tracking id for the whole lifecycle of the certficate issuing.

```javascript
func (contract *SmartContract) RequestIssueCertificate(ctx contractapi.TransactionContextInterface,
	request_id string, student_name string, student_id int, degree string, major string, result float32) (string, error) {

	requester, err := utils.CheckRequester(ctx) // checks that only child institutation can submit request for issung certificate

	if requester != "" && err == nil {
		return "Not Authorized To Request Certificate", nil
	}

	encodedRequetserIdentity, err := ctx.GetClientIdentity().GetID() // it fetched the indenty of child institution which was used when creating the channel. it returned as base64 encoded .

	if err != nil {
		return "", fmt.Errorf("failed read clinet Identity %w", err)
	}
	decodedRequetserIdentity, err := base64.StdEncoding.DecodeString(encodedRequetserIdentity) // decoding the indenity.

	if err != nil {
		return "", fmt.Errorf("failed to decode client Identity %w", err)
	}

    // buildind the request object which will stored the the chain
    // here `Certificate_Hash`,`Issuer_Authority`and `Certificate_Id` are empty beacuse those 
    // field will be filled by the issuer authority. and `Is_Reqeust_Completed` is set to false
    // as the certificate is not created yet.
	request := utils.CertificateRequest{
		Request_Id: request_id, Student_Name: student_name, Student_Id: student_id, Degree: degree, Major: major, Result: result,
		Requester_Authority: string(decodedRequetserIdentity), Certificate_Hash: "", Issuer_Authority: "",
		Is_Reqeust_Completed: false,
		Certificate_Id:       0000,
	}

	requestJson, err := json.Marshal(request) // serializing the request to json object.

	if err != nil {
		return "", fmt.Errorf("failed to json marshal request %w", err)
	}
	err = ctx.GetStub().PutState(request.Request_Id, requestJson) // storing the json object to the ledger
	if err != nil {
		return "", fmt.Errorf("failed to add the request to the ledger %w", err)
	}

	return fmt.Sprintf("Submitted Request Id : ", request.Request_Id), nil // if the request is we will return requested that can used to track the record and for further record.

}
```

### Defination of the `IssueCertificate()` 
Which is used to issues certificate by the parent institution.
it takes `request_id string`,` certitficate_hash string`, `certificate_id int` as parameters. The `request_id string` is the id of the request that has been created by the child institution. Parent institution creates the hash of certificate file using `IPFS` and gives an certifcate id based on the sequencial order.Implementation `IPFS` is not impplemented it will done on the front end part of parent org. we have one `IPFS` service to genrate hash for the file.

```javascript
 func (contract *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface,
	request_id string, certitficate_hash string, certificate_id int) (int, error) {

	Issuer, err := utils.IsIssuer(ctx) // it checks  the transaction creator is permitted or not. 

	if err != nil {
		return 0000, err
	}

	if !Issuer {
		return 0000, fmt.Errorf("Not Authorized To Issue Certificate")
	}

	exits, err := contract.IsRequestExist(ctx, request_id) // Checks the request id exist or not.

	if err != nil {
		return 0000, fmt.Errorf("%w", err)
	}

	if !exits {
		return 0000, fmt.Errorf("Request does not exists with id  : %w", request_id)
	}

    // retrive the request and parse into go struct.
	request, err := contract.ReadRequest(ctx, request_id)

	if err != nil {
		return 0000, fmt.Errorf("%w", err)
	}

	encodeIssuerIdenity, err := ctx.GetClientIdentity().GetID() // it fetches the identity of the client from the channel configuartion. it same as implemented in  the `RequestIssueCertificate()`

	if err != nil {

		return 0000, fmt.Errorf("failed read clinet Identity %w", err)
	}

	decodedIssuerIdentity, err := base64.StdEncoding.DecodeString(encodeIssuerIdenity) // decoding the idenity.
	if err != nil {

		return 0000, fmt.Errorf("failed read clinet Identity %w", err)
	}
    
    // Updationg the request object which was incomplete when
    // request is submitted.
	request.Issuer_Authority = string(decodedIssuerIdentity)
	request.Certificate_Id = certificate_id
	request.Certificate_Hash = certitficate_hash
	request.Is_Reqeust_Completed = true

	jsonRequest, err := json.Marshal(request)

	if err != nil {
		return 0000, fmt.Errorf("failed to marshal request %w", err)
	}

    // Storing the completed request to the ledger
	err = ctx.GetStub().PutState(request_id, jsonRequest)

	if err != nil {
		return 0000, fmt.Errorf("failed to add  request to ledger %w", err)
	}

    // on successfull completation we will return the certificate id which used to verify 
    // from the chain later.
	return request.Certificate_Id, nil
}
```

### Defination of the `ReadRequest()`
It reads and return request from the ledger if the request exist in the ledger. it takes one`request_id` parameter to checks and return the request.

```javascript


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

```

### Defination of the `GetAllTheRequests()`
It returns an array of the all the request objects.

### Defination of the `HistoryOfRequest()`
It returns an array of the all the changes of a particular request. it takes `request_id` as a parameter.
From we get the all the history of a reqeuest what happened to it, who has done what to it.
even sometime try to temper a data from their side we can verufy from it.

```javascript
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
```



**smartcontract.go:** It contains the `main()` function from wherer the chaincode is initiated and started. In golang `main()` function is the entrypoint for starting the program.









