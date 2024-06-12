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


**smartcontract.go:** It contains the `main()` function from wherer the chaincode is initiated and started. In golang `main()` function is the entrypoint for starting the program.










