# Decentralized Certification 

This chaincode will be used across consortiums to issue digital certificates. The consortium will be formed with a parent institution that has the sole right to issue certificates. The child institutions will send requests to the smart contract which will stored on the ledger. Based on these requests, the parent institution will issue the certificates.

**Note**: Chaincode and SmartContract are interchangble. 

## Folders And Files
There are two mian folder **utils** and **chaincode** and one main file **smartcontract** and rest are the dependencies.

**utils:** It contains the all the utility functions and data type for the smartcontract.The `utils.go` file has the the following  structure which will be used on the chaincode.
```javascript
    type CertificateRequest struct {
        Tracking_Id           string  `json:tracking_id` // it will be dynamically created and will be unique
        Student_Name         string  `json:student_name` // it will come from the child institutation
        Student_Id           int     `json:student_id` // it will come from the child institutation
        Student_Email        string  `json:student_email` // it will come from the child institutation
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
It submits request to the parent institution for issuing certificate of behalf of the student.It Takes `tracking_id string`, `student_name string`,` student_id int`, `degree string`, `major string`, `result float32` as parameters all these parameter are stored on the chain and those come from the child institutions.
`tracking_id string` act as tracking id for the whole lifecycle of the certficate issuing.

```javascript
func (contract *SmartContract) RequestIssueCertificate(ctx contractapi.TransactionContextInterface,
	tracking_id string, student_name string, student_id int, student_email string, degree string, major string, result float32) (string, error) {

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
		Request_Id: request_id, Student_Name: student_name, Student_Id: student_id, Student_Email :student_email, Degree: degree, Major: major, Result: result,
		Requester_Authority: string(decodedRequetserIdentity), Certificate_Hash: "", Issuer_Authority: "",
		Is_Reqeust_Completed: false,
		Certificate_Id:       0000,
	}

	requestJson, err := json.Marshal(request) // serializing the request to json object.

	if err != nil {
		return "", fmt.Errorf("failed to json marshal request %w", err)
	}
	err = ctx.GetStub().PutState(request.Tracking_Id, requestJson) // storing the json object to the ledger
	if err != nil {
		return "", fmt.Errorf("failed to add the request to the ledger %w", err)
	}

	return fmt.Sprintf("Submitted Request Id : ", request.Tracking_Id), nil // if the request is we will return requested that can used to track the record and for further record.

}
```

### Defination of the `IssueCertificate()` 
The `IssueCertificate()` function issues a certificate by the parent institution. It takes three parameters:` tracking_id (string)`, `certificate_hash (string)`, and` certificate_id (int)`. The `tracking_id` is the ID created by the child institution. The parent institution generates the hash of the certificate file using `IPFS` (handled on the front end of the parent organization) and assigns a certificate ID sequentially.

```javascript
func (contract *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface,
	tracking_id string, certitficate_hash string, certificate_id int) (int, error) {

	Issuer, err := utils.IsIssuer(ctx) //The function calls `utils.IsIssuer(ctx)` to check if the transaction creator is permitted to issue certificates.
	if err != nil {
		return 0000, err
	}
	if !Issuer {
		return 0000, fmt.Errorf("Not Authorized To Issue Certificate")
	}

	// It calls `contract.IsRequestExist(ctx, tracking_id)` to verify if the request with the given tracking_id exists.
	exits, err := contract.IsRequestExist(ctx, tracking_id)
	if err != nil {
		return 0000, fmt.Errorf("%w", err)
	}
	if !exits {
		return 0000, fmt.Errorf("Request does not exists with id  : %w", tracking_id)
	}

	// The function retrieves the request using `contract.ReadRequest(ctx, tracking_id)`
	request, err := contract.ReadRequest(ctx, tracking_id)
	// Double Certificate Creation Check
	if request.Is_Reqeust_Completed {
		return -1, nil
	}

	if err != nil {
		return 0000, fmt.Errorf("%w", err)
	}

	// It fetches the identity of the client from the channel configuration using `ctx.GetClientIdentity().GetID().`
	encodeIssuerIdenity, err := ctx.GetClientIdentity().GetID()

	if err != nil {

		return 0000, fmt.Errorf("failed read clinet Identity %w", err)
	}

	//The function decodes the issuer identity using` base64.StdEncoding.DecodeString(encodeIssuerIdenity)`.
	decodedIssuerIdentity, err := base64.StdEncoding.DecodeString(encodeIssuerIdenity)
	if err != nil {

		return 0000, fmt.Errorf("failed read clinet Identity %w", err)
	}

	// Updates the request object with the following
	request.Issuer_Authority = string(decodedIssuerIdentity)
	request.Certificate_Id = certificate_id
	request.Certificate_Hash = certitficate_hash
	request.Is_Reqeust_Completed = true

	// The updated request is serialized into JSON using `json.Marshal(request)`
	jsonRequest, err := json.Marshal(request)

	if err != nil {
		return 0000, fmt.Errorf("failed to marshal request %w", err)
	}

	// The function creates a composite key for the certificate using `ctx.GetStub().CreateCompositeKey(certKey, []string{strconv.Itoa(request.Certificate_Id), request.Tracking_Id})`.
	// This Composit Key Will be used to retrive a specific certifcate naesd on its id from the chain.
	compositKey, err := ctx.GetStub().CreateCompositeKey(certKey, []string{strconv.Itoa(request.Certificate_Id), request.Tracking_Id})

	if err != nil {
		return 0000, fmt.Errorf("failed to create composite key: %w", err)
	}

	// It stores the composite key in the ledger using `ctx.GetStub().PutState(compositKey, []byte{0x00})`.
	err = ctx.GetStub().PutState(compositKey, []byte{0x00})

	if err != nil {
		return 0000, fmt.Errorf("failed to add  compositeKey to ledger %w", err)
	}

	// The function creates a composite key for the certificate hash using `ctx.GetStub().		CreateCompositeKey(certhashKey, []string{request.Certificate_Hash, request.Tracking_Id})`.
	// It will be used to retrive certicafte or verify a certificate by its hash from on chain.

	compositKeyForCertHash, err := ctx.GetStub().CreateCompositeKey(certhashKey, []string{request.Certificate_Hash, request.Tracking_Id})

	if err != nil {
		return 0000, fmt.Errorf("failed to create composite key for hash: %w", err)
	}

	err = ctx.GetStub().PutState(compositKeyForCertHash, []byte{0x00})

	if err != nil {
		return 0000, fmt.Errorf("failed to add  compositeKeyforcerthash to ledger %w", err)
	}

	// The function stores the completed request in the ledger using `ctx.GetStub().PutState(tracking_id, jsonRequest)`.
	err = ctx.GetStub().PutState(tracking_id, jsonRequest)

	if err != nil {
		return 0000, fmt.Errorf("failed to add  request to ledger %w", err)
	}
	// On successful completion, the function returns the `Certificate_Id` which can be used for later verification from the chain.
	return request.Certificate_Id, nil
}
```

### Defination of the `ReadRequest()`
It reads and return request from the ledger if the request exist in the ledger. it takes one`tracking_id` parameter to checks and return the request.

```javascript


func (contract *SmartContract) ReadRequest(ctx contractapi.TransactionContextInterface, tracking_id string) (*utils.CertificateRequest, error) {

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

```





### Defination of the `GetAllTheRequests()`
The function `GetAllTheRequests` is a method of the SmartContract. It retrieves all certificate requests stored in the blockchain ledger.

```javascript
	func (contract *SmartContract) GetAllTheRequests(ctx contractapi.TransactionContextInterface) ([]*utils.CertificateRequest, error) {

	//It  initializes an iterator to fetch all the states in the ledger. An empty range ("", "") means it will return all key-value pairs.
	RequestQueryIterartor, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, fmt.Errorf("failed to get state %w", err)
	}
	// Ensures that the iterator is closed when the function exits.
	defer RequestQueryIterartor.Close()

	// Array to hold the requests
	var requests []*utils.CertificateRequest

	// HasNext(): Checks if there are more results.
	// Next(): Retrieves the next key-value pair from the iterator.
	// json.Unmarshal: Converts the JSON-encoded value to a CertificateRequest object.
	// append: Adds the CertificateRequest object to the requests slice.

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
	// Returns the array of `CertificateRequest` objects 
	return requests, nil
}
```

### Defination of the `HistoryOfRequest()`
It returns an array of the all the changes of a particular request. it takes `tracking_id` as a parameter.
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

### Defination of the `ReadCertificateByCertificateId()`
The` ReadCertificateByCertificateId()` function retrieves a certificate request from the ledger using a provided `certificate_id (int)` The unique ID  certificate ID. It takes the certificate ID as input and returns the corresponding CertificateRequest object if found.

```javascript

func (contract *SmartContract) ReadCertificateByCertificateId(ctx contractapi.TransactionContextInterface, certificate_id int) (*utils.CertificateRequest, error) {

	// The function creates a query to retrieve the state by a partial composite key using `ctx.GetStub().GetStateByPartialCompositeKey(certKey, []string{strconv.Itoa(certificate_id)})`
	resultIterartor, err := ctx.GetStub().GetStateByPartialCompositeKey(certKey, []string{strconv.Itoa(certificate_id)})

	if err != nil {
		return nil, err
	}
	
	// It uses defer resultIterator.Close() to ensure the iterator is closed after the function execution.
	defer resultIterartor.Close()
	
	// The function checks if the iterator has any results using resultIterator.HasNext().
	if !resultIterartor.HasNext() {
		return nil, fmt.Errorf("not certificate found for the id %d", certificate_id)
	}

	// It retrieves the next result from the iterator using resultIterator.Next().
	queryResponse, err := resultIterartor.Next()

	if err != nil {
		return nil, err
	}

	// The function splits the composite key using ctx.GetStub().SplitCompositeKey(queryResponse.Key).
	_, compositeKey, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)

	if err != nil {
		return nil, err
	}

	// It extracts the tracking_id from the composite key parts. The tracking_id is the second element in the compositeKey array.
	tracking_id := compositeKey[1]
	
	// The function calls contract.ReadRequest(ctx, tracking_id) to retrieve the certificate request using the extracted tracking_id.
	request, err := contract.ReadRequest(ctx, tracking_id)

	if err != nil {
		return nil, err
	}

	// On successful completion, the function returns the CertificateRequest object.
	return request, nil

}

```
### Defination of `VerifyCertificateByCertificateHash()`
The function `VerifyCertificateByCertificateHash()` verifies a certificate is it legit or by cheking it preseance on the chain.  It takes `cert_hash (string)` as parameter. It verifies a certificate by its hash and retrieves the corresponding certificate request from the ledger if it present on the chain. 


```javascript
func (contract *SmartContract) VerifyCertificateByCertificateHash(ctx contractapi.TransactionContextInterface, cert_hash string) (*utils.CertificateRequest, error) {
	
	// It initializes an iterator to fetch all states that match the partial composite key formed using `certhashKey` and `cert_hash`.
	resultIterartor, err := ctx.GetStub().GetStateByPartialCompositeKey(certhashKey, []string{cert_hash})

	if err != nil {
		return nil, err
	}
	// Ensures that the iterator is closed when the function exits to release resources.
	defer resultIterartor.Close()

	// If the iterator does not have any results, the function returns nil and an error indicating that no certificate was found for the given hash.
	if !resultIterartor.HasNext() {
		return nil, fmt.Errorf("No Certificate found for the hash : %w", cert_hash)
	}

	// Retrieve the next result
	queryResponse, err := resultIterartor.Next()

	if err != nil {
		return nil, err
	}

	// It splits the composite key to extract its components. The second component is assumed to be the `tracking_id`.
	_, compositKeyForHash, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)

	// retrieve the tracking ID.
	tracking_id := compositKeyForHash[1]

	//It calls the ReadRequest method (presumably defined elsewhere) to read the certificate request associated with the tracking_id.
	request, err := contract.ReadRequest(ctx, tracking_id)

	if err != nil {
		return nil, err
	}

	// Return the certificate request:
	return request, nil

}
```

**smartcontract.go:** It contains the `main()` function from wherer the chaincode is initiated and started. In golang `main()` function is the entrypoint for starting the program.









