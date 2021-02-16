package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	amc "github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/amc"
	"github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/associatebeneficiary"
	bankaccount "github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/bankaccounts"
	"github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/banks"
	bank "github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/banks"
	branch "github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/branch"
	"github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/checklist"
	fund "github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/funds"
	"github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/notifications"
	"github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/roles"
	security "github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/security"
	tax "github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/tax"
	"github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/transactions/transaction"
	txnhistory "github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/transactions/txnhistory"
	unitholder "github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/unitholder"
	user "github.com/hyperledger/fabric-samples/chaincode/amc-chaincode/entities/users"
)

type SimpleContract struct {
	contractapi.Contract
}

func (sc *SimpleContract) AddBank(ctx contractapi.TransactionContextInterface) string {

	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 3 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 3 arguments for adding a bank."}`
	}
	Code := "BANK_" + params[1]
	bankAsByte, _ := ctx.GetStub().GetState(Code)
	if bankAsByte != nil {
		return `{"status": 400 , "message": "bank code already exists "}`

	}

	if CheckQuery(ctx, "bank", "bank_name", params[0]) == "true" {
		return `{"status": 400 , "message": "Bank name already exists"}`
	}
	bank := banks.Bank{
		ObjectType:   "bank",
		BankName:     params[0],
		BankCode:     Code,
		Status:       "active",
		CreationDate: params[2],
	}
	bankAsBytes, _ := json.Marshal(bank)

	_ = ctx.GetStub().PutState(Code, bankAsBytes)

	return `{"status": 200 , "message": "Bank successfully added"}`
}

// Read returns the value at key in the world state
func (sc *SimpleContract) GetBankInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to query"}`
	}
	Code := "BANK_" + args[0]
	bankAsByte, _ := ctx.GetStub().GetState(Code)
	if bankAsByte == nil {

		return `{"status": 400 , "message": "Bank Data not found for the key provided"}`
	}
	bank := new(bank.Bank)
	_ = json.Unmarshal(bankAsByte, bank)
	s := strings.Split(bank.BankCode, "_")
	bank.BankCode = s[1]
	aAsBytes, _ := json.Marshal(bank)
	return `{"status": 200 , "data" : ` + string(aAsBytes) + `}`
}

func (s *SimpleContract) GetAllBanks(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	count := 0
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetStateByRangeWithPagination("BANK_"+params[1], "BANK_9999999", i, "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		bank := bank.Bank{}
		json.Unmarshal(queryResponse.Value, &bank)
		if bank.ObjectType == "bank" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false
			buffer.WriteString(string(queryResponse.Value))
			count++

		}
	}
	buffer.WriteString("],")
	total := TotalCount(ctx, "BANK_", "BANK_9999999", "")
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "BANK_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (s *SimpleContract) GetQueryData(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 5 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 5 arguments"}`
	}
	page, _ := strconv.Atoi(params[0])
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetQueryResultWithPagination(params[2], i, params[1])

	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}

		if first == false {
			buffer.WriteString(",")
		}
		first = false
		buffer.WriteString(string(queryResponse.Value))

	}
	buffer.WriteString("],")
	total := TotalCount(ctx, params[3], params[4], "")
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, params[3], "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (sc *SimpleContract) UpdateBankStatus(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id and new status to update"}`
	}
	Code := "BANK_" + args[0]
	bankAsBytes, _ := ctx.GetStub().GetState(Code)

	if bankAsBytes == nil {

		return `{"status": 400 , "message": "Bank Data not found for the key provided"}`

	}
	bank := new(bank.Bank)
	_ = json.Unmarshal(bankAsBytes, bank)

	if bank.Status == args[1] {
		return `{"status": 400 , "message": "Status is already same as provided"}`

	} else {
		bank.Status = args[1]
	}

	aAsBytes, _ := json.Marshal(bank)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Bank status updated"}`

}

// amc chaincode methods...

func (sc *SimpleContract) AddAmc(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 16 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 16 arguments for adding a AMC."}`
	}
	Code := "AMC_" + params[0]
	amcAsByte, _ := ctx.GetStub().GetState(Code)
	if amcAsByte != nil {
		return `{"status": 400 , "message": "AMC code already exists "}`

	}
	amc := amc.Amc{
		ObjectType:                "amc",
		Name:                      params[1],
		FocalPersonEmail:          params[2],
		FocalPerson:               params[3],
		AmcSignatories:            params[4],
		AmcTaxAdvisor:             params[5],
		ConcernedOfficer:          params[6],
		SubtituteConcernedOfficer: params[7],
		AmcCode:                   Code,
		Status:                    "active",
		AmcBr:                     params[8],
		AuthAmcSignatories:        params[9],
		AmcAuditor:                params[10],
		TaxExemption:              params[11],
		From:                      params[12],
		To:                        params[13],
		TxnCreatorField:           params[14],
		CreatedAt:                 params[15],
	}
	amcAsBytes, _ := json.Marshal(amc)

	_ = ctx.GetStub().PutState(Code, amcAsBytes)

	return `{"status": 200 , "message": "AMC successfully added"}`
}

func TotalCount(ctx contractapi.TransactionContextInterface, start string, end string, pagee string) int {
	count := 0
	page, _ := strconv.Atoi(pagee)
	i := int32(page)
	resultsIterator, _, _ := ctx.GetStub().GetStateByRangeWithPagination(start, end, i, "")
	defer resultsIterator.Close()
	for resultsIterator.HasNext() {
		q, _ := resultsIterator.Next()
		fmt.Println(q)
		count++
	}
	return count
}

func (sc *SimpleContract) GetAmcInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to query"}`
	}

	Code := "AMC_" + args[0]
	amcAsByte, _ := ctx.GetStub().GetState(Code)
	if amcAsByte == nil {

		return `{"status": 400 , "message": "Amc Data not found for the key provided"}`

	}
	amc := new(amc.Amc)
	_ = json.Unmarshal(amcAsByte, amc)
	s := strings.Split(amc.AmcCode, "_")
	amc.AmcCode = s[1]
	aAsBytes, _ := json.Marshal(amc)
	return `{"status": 200 , "data" : ` + string(aAsBytes) + `}`
}
func GetUserName(ctx contractapi.TransactionContextInterface, email string) string {
	Code := "USER_" + email
	userAsBytes, _ := ctx.GetStub().GetState(Code)
	if userAsBytes == nil {
		return "User Data not found for the key provided"
	}
	user := new(user.User)
	_ = json.Unmarshal(userAsBytes, user)
	return user.Name

}
func GetAmc(ctx contractapi.TransactionContextInterface, ntn string) string {

	Code := "AMC_" + ntn
	brokerAsByte, _ := ctx.GetStub().GetState(Code)
	if brokerAsByte == nil {
		return `{"status": 400 , "message": "AMC Data not found for the key provided"}`
	}
	brokerdata := new(associatebeneficiary.AssociateBeneficiary)
	_ = json.Unmarshal(brokerAsByte, brokerdata)
	s := strings.Split(brokerdata.NTN, "_")
	brokerdata.NTN = s[1]
	aAsBytes, _ := json.Marshal(brokerdata)
	return string(aAsBytes)
}

func (sc *SimpleContract) GetBrokerAmc(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting amc code to query"}`
	}
	amc := strings.Split(args[0], ",")
	var data = "["
	for i := 0; i < len(amc); i++ {
		if i == (len(amc) - 1) {
			data += GetAmc(ctx, amc[i])
		} else {
			data += GetAmc(ctx, amc[i]) + ","
		}
	}
	data += "]"
	return `{"status": 200 , "data" : ` + data + `}`
}

func (sc *SimpleContract) UpdateAmc(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 10 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting 2 arguments to update"}`
	}
	Code := "AMC_" + args[0]
	amcAsBytes, _ := ctx.GetStub().GetState(Code)

	if amcAsBytes == nil {

		return `{"status": 400 , "message": "AMC Data not found for the key provided"}`

	}
	amc := new(amc.Amc)
	_ = json.Unmarshal(amcAsBytes, amc)
	if args[1] != "" {
		amc.Name = args[1]
	}
	if args[2] != "" {
		amc.To = args[2]
	}
	if args[3] != "" {
		amc.From = args[3]
	}
	if args[4] != "" {
		amc.AmcBr = args[4]
	}
	if args[5] != "" {
		amc.FocalPerson = args[5]
	}
	if args[6] != "" {
		amc.FocalPersonEmail = args[6]
	}
	if args[7] != "" {
		amc.AmcAuditor = args[7]
	}
	if args[8] != "" {
		amc.TaxExemption = args[8]
	}
	if args[9] != "" {
		amc.AmcTaxAdvisor = args[9]
	}
	aAsBytes, _ := json.Marshal(amc)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message":"AMC data updated "}`

}

func (sc *SimpleContract) GetAmcMembers(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to query"}`
	}
	Code := "AMC_" + args[0]
	amcAsBytes, _ := ctx.GetStub().GetState(Code)
	amc := new(amc.Amc)
	_ = json.Unmarshal(amcAsBytes, amc)
	var totaldata = "{"
	var signatorydata = `"signatories":[`
	signatories := strings.Split(amc.AmcSignatories, ",")
	for i := 0; i < len(signatories); i++ {
		if i == (len(signatories) - 1) {
			signatorydata += `{"name" :"` + GetUserName(ctx, signatories[i]) + `","email":"` + signatories[i] + `"}],`
		} else {
			signatorydata += `{"name":" ` + GetUserName(ctx, signatories[i]) + `","email":"` + signatories[i] + `"},`
		}
	}
	totaldata += signatorydata
	totaldata += ` "concerned_officer":{"name":"` + GetUserName(ctx, amc.ConcernedOfficer) + `","email":"` + amc.ConcernedOfficer + `"},`
	totaldata += ` "subtitute_concerned_officer":{"name":"` + GetUserName(ctx, amc.SubtituteConcernedOfficer) + `","email":"` + amc.SubtituteConcernedOfficer + `"},`
	totaldata += ` "txn_creator_field":{"name":"` + GetUserName(ctx, amc.TxnCreatorField) + `","email":"` + amc.TxnCreatorField + `"}}`
	return `{"status": 200 , "data" : ` + totaldata + `}`
}
func (sc *SimpleContract) UpdateAmcStatus(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id and new status to update"}`
	}
	Code := "AMC_" + args[0]
	amcAsBytes, _ := ctx.GetStub().GetState(Code)

	if amcAsBytes == nil {

		return `{"status": 400 , "message": "AMC Data not found for the key provided"}`

	}
	amc := new(amc.Amc)
	_ = json.Unmarshal(amcAsBytes, amc)

	if amc.Status == args[1] {
		return `{"status": 400 , "message": "Status is already same as provided"}`

	} else {
		amc.Status = args[1]
	}

	aAsBytes, _ := json.Marshal(amc)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "AMC status updated"}`

}

func (s *SimpleContract) GetAllAmcs(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	count := 0
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetStateByRangeWithPagination("AMC_"+params[1], "AMC_9999999", i, "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		amc := amc.Amc{}
		json.Unmarshal(queryResponse.Value, &amc)
		if amc.ObjectType == "amc" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false
			buffer.WriteString(string(queryResponse.Value))
			count++

		}
	}
	buffer.WriteString("],")
	total := TotalCount(ctx, "AMC_", "AMC_9999999", "")
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "AMC_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (sc *SimpleContract) AddFund(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 9 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 9 arguments for adding a Fund."}`
	}
	Code := "FUND_" + params[0]
	fundAsByte, _ := ctx.GetStub().GetState(Code)
	if fundAsByte != nil {
		return `{"status": 400 , "message": "Fund Symbol code already exists "}`

	}
	fund := fund.Fund{
		ObjectType:          "fund",
		AmcName:             params[1],
		Nature:              params[2],
		DateOfIncorporation: params[3],
		PsxListing:          params[4],
		FundName:            params[5],
		SymbolCode:          Code,
		DateOfRevocation:    params[6],
		Status:              "active",
		CreatedAt:           params[7],
		Nav:                 params[8],
	}
	fundAsBytes, _ := json.Marshal(fund)

	_ = ctx.GetStub().PutState(Code, fundAsBytes)

	return `{"status": 200 , "message": "Fund successfully added"}`
}

func (s *SimpleContract) GetAllFunds(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	count := 0
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetStateByRangeWithPagination("FUND_"+params[1], "FUND_9999999", i, "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		fund := fund.Fund{}
		json.Unmarshal(queryResponse.Value, &fund)
		if fund.ObjectType == "fund" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false
			buffer.WriteString(string(queryResponse.Value))
			count++

		}
	}
	buffer.WriteString("],")
	total := TotalCount(ctx, "FUND_", "FUND_9999999", "")
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "FUND_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`

}

func (s *SimpleContract) GetFundsByAmcCode(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting amc code to query"}`
	}
	Code := "AMC_" + args[0]
	amcAsBytes, _ := ctx.GetStub().GetState(Code)
	amc := new(amc.Amc)
	_ = json.Unmarshal(amcAsBytes, amc)

	resultsIterator, err := ctx.GetStub().GetStateByRange("FUND_", "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		fund := fund.Fund{}
		json.Unmarshal(queryResponse.Value, &fund)
		if fund.ObjectType == "fund" && fund.Status == "active" && fund.AmcName == amc.Name {
			if first == false {
				buffer.WriteString(",")
			}
			first = false

			buffer.WriteString(string(queryResponse.Value))
		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {

		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "FUND_", "")
		return `{"status": 200 , "data" : ` + refined + `}`

	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`

}

func (sc *SimpleContract) DeleteFund(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to delete"}`
	}
	Code := "FUND_" + args[0]
	fundAsBytes, _ := ctx.GetStub().GetState(Code)

	if fundAsBytes == nil {

		return `{"status": 400 , "message": "Fund Data not found for the key provided"}`

	}
	fund := new(fund.Fund)
	_ = json.Unmarshal(fundAsBytes, fund)

	fund.Status = "deleted"

	aAsBytes, _ := json.Marshal(fund)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message":"fund deleted"}`

}
func (sc *SimpleContract) UpdateFund(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 11 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting 9 arguments to update"}`
	}
	Code := "FUND_" + args[0]
	fundAsBytes, _ := ctx.GetStub().GetState(Code)

	if fundAsBytes == nil {

		return `{"status": 400 , "message": "Fund Data not found for the key provided"}`

	}
	fund := new(fund.Fund)
	_ = json.Unmarshal(fundAsBytes, fund)
	if args[1] != "" {
		fund.AmcName = args[1]
	}
	if args[2] != "" {
		fund.Nature = args[2]
	}
	if args[3] != "" {
		fund.DateOfIncorporation = args[3]
	}
	if args[4] != "" {
		fund.PsxListing = args[4]
	}
	if args[5] != "" {
		fund.FundName = args[5]
	}
	if args[6] != "" {
		fund.DateOfRevocation = args[6]
	}
	if args[7] != "" {
		fund.Status = args[7]
	}
	if args[8] != "" {
		fund.Nav = args[8]
	}
	if args[9] != "" {
		fund.MoneyMarket = args[9]
	}
	if args[10] != "" {
		fund.StockMarket = args[10]
	}
	aAsBytes, _ := json.Marshal(fund)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message":"fund data updated "}`

}

func (sc *SimpleContract) AddBranch(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 7 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 7 arguments for adding a Branch."}`
	}
	Code := "BRANCH_" + params[0]
	branchAsByte, _ := ctx.GetStub().GetState(Code)
	if branchAsByte != nil {
		return `{"status": 400 , "message": "Branch code already exists "}`
	}
	branchdata := new(branch.Branch)
	_ = json.Unmarshal(branchAsByte, branchdata)

	if branchdata.BranchName == params[2] {
		return `{"status": 400 , "message": "Branch name already exists"}`

	}
	branch := branch.Branch{
		ObjectType:    "branch",
		BankName:      params[1],
		BranchName:    params[2],
		BranchCode:    Code,
		BranchAddress: params[3],
		City:          params[4],
		Area:          params[5],
		Status:        "active",
		CreatedAt:     params[6],
	}
	branchAsBytes, _ := json.Marshal(branch)

	_ = ctx.GetStub().PutState(Code, branchAsBytes)

	return `{"status": 200 , "message": "Branch successfully added"}`
}

func (sc *SimpleContract) UpdateBranchInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 6 {
		return `{"status": 400 , "message": "Incorrect number of arguments,Please provide 6 arguments"}`
	}
	Code := "BRANCH_" + args[0]
	branchAsBytes, _ := ctx.GetStub().GetState(Code)

	if branchAsBytes == nil {

		return `{"status": 400 , "message": "Branch data not found for the key provided"}`

	}
	branch := new(branch.Branch)
	_ = json.Unmarshal(branchAsBytes, branch)
	if args[1] != "" {
		branch.BankName = args[1]
	}
	if args[2] != "" {
		branch.BranchName = args[2]
	}
	if args[3] != "" {
		branch.BranchAddress = args[3]
	}
	if args[4] != "" {
		branch.City = args[4]
	}
	if args[5] != "" {
		branch.Area = args[5]
	}
	aAsBytes, _ := json.Marshal(branch)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Branch information updated"}`

}

func (s *SimpleContract) GetAllBranches(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetStateByRangeWithPagination("BRANCH_"+params[1], "BRANCH_9999999", i, "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		branche := branch.Branch{}
		json.Unmarshal(queryResponse.Value, &branche)
		if len(params) == 2 {
			if branche.ObjectType == "branch" {
				if first == false {
					buffer.WriteString(",")
				}
				first = false

				buffer.WriteString(string(queryResponse.Value))
			}
		} else {
			if branche.ObjectType == "branch" && branche.BankName == params[2] {
				if first == false {
					buffer.WriteString(",")
				}
				first = false

				buffer.WriteString(string(queryResponse.Value))
			}
		}
	}
	buffer.WriteString("],")
	total := TotalCount(ctx, "BRANCH_", "BRANCH_9999999", "")
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "BRANCH_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (sc *SimpleContract) UpdateBranchStatus(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id and new status to update"}`
	}
	Code := "BRANCH_" + args[0]
	branchAsBytes, _ := ctx.GetStub().GetState(Code)

	if branchAsBytes == nil {

		return `{"status": 400 , "message": "Branch Data not found for the key provided"}`

	}
	branch := new(branch.Branch)
	_ = json.Unmarshal(branchAsBytes, branch)

	if branch.Status == args[1] {
		return `{"status": 400 , "message": "Status is already same as provided"}`

	} else {
		branch.Status = args[1]
	}

	aAsBytes, _ := json.Marshal(branch)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Branch status updated"}`

}

func (sc *SimpleContract) AddBankAccount(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 14 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 14 arguments for adding a bank account."}`
	}
	Code := "ACCOUNT_" + params[0]
	accountAsByte, _ := ctx.GetStub().GetState(Code)
	if accountAsByte != nil {
		return `{"status": 400 , "message": "Account No already exists "}`

	}
	fmt.Println("params", params)
	account := bankaccount.BankAccount{
		ObjectType:         "account",
		AmcName:            params[1],
		BankName:           params[2],
		BranchName:         params[3],
		ProductPurpose:     params[4],
		CreatedAt:          params[5],
		AccountNo:          Code,
		IBAN:               params[6],
		FundName:           params[7],
		Status:             "active",
		AccountTitle:       params[8],
		NatureOfAccount:    params[9],
		Currency:           params[10],
		BalanceAmount:      params[11],
		OperationHeadEmail: params[12],
		SMA:                params[13],
	}
	accountAsBytes, _ := json.Marshal(account)

	_ = ctx.GetStub().PutState(Code, accountAsBytes)

	return `{"status": 200 , "message": "Account successfully added"}`
}

// Read returns the value at key in the world state
func (sc *SimpleContract) GetBankAccountInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to query"}`
	}
	Code := "ACCOUNT_" + args[0]
	bankAsByte, _ := ctx.GetStub().GetState(Code)
	if bankAsByte == nil {

		return `{"status": 400 , "message": "Bank Account Data not found for the key provided"}`
	}
	bankaccount := new(bankaccount.BankAccount)
	_ = json.Unmarshal(bankAsByte, bankaccount)
	s := strings.Split(bankaccount.AccountNo, "_")
	bankaccount.AccountNo = s[1]
	aAsBytes, _ := json.Marshal(bankaccount)
	return `{"status": 200 , "data" : ` + string(aAsBytes) + `}`
}

func (sc *SimpleContract) AddBroker(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 10 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 10 arguments for adding a broker."}`
	}
	Code := "BROKER_" + params[0]
	brokerAsByte, _ := ctx.GetStub().GetState(Code)
	if brokerAsByte != nil {
		return `{"status": 400 , "message": "Broker already exists "}`

	}
	broker := associatebeneficiary.AssociateBeneficiary{
		ObjectType:   "broker",
		CompanyName:  params[1],
		BankName:     params[2],
		FocalEmail:   params[3],
		CompanyCode:  Code,
		IBAN:         params[4],
		STN:          params[5],
		Branch:       params[6],
		CreatedAt:    params[7],
		AccountTitle: params[8],
		NTN:          params[9],
	}
	brokerAsBytes, _ := json.Marshal(broker)

	_ = ctx.GetStub().PutState(Code, brokerAsBytes)

	return `{"status": 200 , "message": "Broker successfully added"}`
}
func (sc *SimpleContract) GetBrokerInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting ntn to query"}`
	}

	Code := "BROKER_" + args[0]
	brokerAsByte, _ := ctx.GetStub().GetState(Code)
	if brokerAsByte == nil {
		return `{"status": 400 , "message": "Broker Data not found for the key provided"}`
	}
	brokerdata := new(associatebeneficiary.AssociateBeneficiary)
	_ = json.Unmarshal(brokerAsByte, brokerdata)
	s := strings.Split(brokerdata.CompanyCode, "_")
	brokerdata.CompanyCode = s[1]
	aAsBytes, _ := json.Marshal(brokerdata)
	return `{"status": 200 , "data" : ` + string(aAsBytes) + `}`
}

func (sc *SimpleContract) UpdateBrokerInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	attrval, _, _ := ctx.GetClientIdentity().GetAttributeValue("email")
	if len(args) != 9 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting 9 params"}`
	}
	Code := "BROKER_" + args[0]
	brokerAsBytes, _ := ctx.GetStub().GetState(Code)

	if brokerAsBytes == nil {

		return `{"status": 400 , "message": "Broker Data not found for the key provided"}`

	}
	broker := new(associatebeneficiary.AssociateBeneficiary)
	_ = json.Unmarshal(brokerAsBytes, broker)

	if args[1] != "" {
		broker.CompanyName = args[1]
	}
	if args[2] != "" {
		broker.BankName = args[2]
	}
	if args[3] != "" {
		broker.FocalEmail = args[3]
	}
	if args[4] != "" {
		broker.IBAN = args[4]
	}
	if args[5] != "" {
		broker.STN = args[5]
	}
	if args[6] != "" {
		broker.Branch = args[6]
	}
	if args[7] != "" {
		broker.Amc = args[7]
		broker.Assignee = attrval
	}
	if args[8] != "" {
		broker.AccountTitle = args[8]
	}

	aAsBytes, _ := json.Marshal(broker)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Company info updated"}`

}

func (s *SimpleContract) GetAllBrokers(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	count := 0
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetStateByRangeWithPagination("BROKER_"+params[1], "BROKER_9999999", i, "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		broker := associatebeneficiary.AssociateBeneficiary{}
		json.Unmarshal(queryResponse.Value, &broker)
		if len(params) != 3 {
			if broker.ObjectType == "broker" {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString(string(queryResponse.Value))
				count++

			}
		} else {
			if broker.ObjectType == "broker" && broker.CompanyName == params[2] {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString(string(queryResponse.Value))
				count++

			}
		}
	}
	buffer.WriteString("],")
	total := TotalCount(ctx, "BROKER_", "BROKER_9999999", "")
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "BROKER_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (s *SimpleContract) GetAllBankAccounts(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	count := 0
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetStateByRangeWithPagination("ACCOUNT_"+params[1], "", i, "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		account := bankaccount.BankAccount{}
		json.Unmarshal(queryResponse.Value, &account)
		if len(params) != 3 {
			if account.ObjectType == "account" {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString(string(queryResponse.Value))
				count++

			}
		} else {
			if account.ObjectType == "account" && account.SMA == "true" {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString(string(queryResponse.Value))
				count++

			}
		}
	}
	buffer.WriteString("],")
	total := TotalCount(ctx, "ACCOUNT_", "", "")
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "ACCOUNT_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (s *SimpleContract) GetAccountsByAmcCode(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting amc code to query"}`
	}
	Code := "AMC_" + args[0]
	amcAsBytes, _ := ctx.GetStub().GetState(Code)
	amc := new(amc.Amc)
	_ = json.Unmarshal(amcAsBytes, amc)

	resultsIterator, err := ctx.GetStub().GetStateByRange("ACCOUNT_", "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		account := bankaccount.BankAccount{}
		json.Unmarshal(queryResponse.Value, &account)
		if account.ObjectType == "account" && account.AmcName == amc.Name {
			if first == false {
				buffer.WriteString(",")
			}
			first = false

			buffer.WriteString(string(queryResponse.Value))
		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "ACCOUNT_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (s *SimpleContract) GetAccountsByFund(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	Code := "AMC_" + args[0]
	amcAsBytes, _ := ctx.GetStub().GetState(Code)
	count := 0
	amc := new(amc.Amc)
	_ = json.Unmarshal(amcAsBytes, amc)

	resultsIterator, err := ctx.GetStub().GetStateByRange("ACCOUNT_", "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		account := bankaccount.BankAccount{}
		json.Unmarshal(queryResponse.Value, &account)
		if len(args) != 3 {
			if account.ObjectType == "account" {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString(string(queryResponse.Value))
				count++

			}
		} else {
			if account.ObjectType == "account" && account.FundName == args[2] {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString(string(queryResponse.Value))
				count++

			}
		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "ACCOUNT_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (sc *SimpleContract) UpdateBankAccountStatus(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id and new status to update"}`
	}
	Code := "ACCOUNT_" + args[0]
	accountAsBytes, _ := ctx.GetStub().GetState(Code)

	if accountAsBytes == nil {

		return `{"status": 400 , "message": "Account Data not found for the key provided"}`

	}
	account := new(bankaccount.BankAccount)
	_ = json.Unmarshal(accountAsBytes, account)

	if account.Status == args[1] {
		return `{"status": 400 , "message": "Status is already same as provided"}`

	} else {
		account.Status = args[1]
	}

	aAsBytes, _ := json.Marshal(account)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Bank Account status updated"}`

}

func (sc *SimpleContract) UpdateBankAccountInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 10 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting 10 params"}`
	}
	Code := "ACCOUNT_" + args[0]
	accountAsBytes, _ := ctx.GetStub().GetState(Code)

	if accountAsBytes == nil {

		return `{"status": 400 , "message": "Account Data not found for the key provided"}`

	}
	account := new(bankaccount.BankAccount)
	_ = json.Unmarshal(accountAsBytes, account)

	if args[1] != "" {
		account.BankName = args[1]
	}
	if args[2] != "" {
		account.BranchName = args[2]
	}
	if args[3] != "" {
		account.ProductPurpose = args[3]
	}
	if args[4] != "" {
		account.FundName = args[4]
	}
	if args[5] != "" {
		account.Status = args[5]
	}
	if args[6] != "" {
		account.AccountTitle = args[6]
	}
	if args[7] != "" {
		account.NatureOfAccount = args[7]
	}
	if args[8] != "" {
		account.OperationHeadEmail = args[8]
	}
	if args[9] != "" {
		account.SMA = args[9]
	}

	aAsBytes, _ := json.Marshal(account)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Bank Account info updated"}`

}

func (sc *SimpleContract) UpdateBankAccountAmount(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 3 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting account no amount and flag"}`
	}
	Code := "ACCOUNT_" + args[0]
	accountAsBytes, _ := ctx.GetStub().GetState(Code)

	if accountAsBytes == nil {

		return `{"status": 400 , "message": "Account Data not found for the key provided"}`
	}
	account := new(bankaccount.BankAccount)
	_ = json.Unmarshal(accountAsBytes, account)
	balance, _ := strconv.Atoi(account.BalanceAmount)
	amount, _ := strconv.Atoi(args[1])

	if args[2] == "PAYMENT" {
		if amount > balance {
			return `{"status": 400 , "message": "Insufficient balance for the amount provided"}`
		}
		balance = balance - amount
	} else {
		balance = balance + amount
	}
	account.BalanceAmount = strconv.Itoa(balance)
	aAsBytes, _ := json.Marshal(account)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Bank account amount updated"}`

}

func (sc *SimpleContract) AddUnitHolder(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 22 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 22 arguments for adding a Unit Holder."}`
	}
	Code := "UNITHOLDER_" + params[0]

	unitholderAsByte, _ := ctx.GetStub().GetState(Code)
	if unitholderAsByte != nil {
		return `{"status": 400 , "message": "Folio No already exists "}`

	}

	unitholder := unitholder.UnitHolder{
		ObjectType:       "unitholder",
		AccountName:      params[1],
		AccountNo:        params[2],
		BankName:         params[3],
		FundSelection:    params[4],
		BalanceUnit:      params[5],
		Cnic:             params[6],
		Status:           "active",
		Mobile:           params[7],
		City:             params[8],
		RegistrationDate: params[9],
		AccountTitle:     params[10],
		Nature:           params[11],
		BranchName:       params[12],
		AmcName:          params[13],
		ClientCode:       params[14],
		Type:             params[15],
		FolioNo:          Code,
		Ntn:              params[16],
		Address:          params[17],
		Country:          params[18],
		Upload:           params[19],
		CreatedAt:        params[20],
		Name:             params[21],
	}
	unitholderAsBytes, _ := json.Marshal(unitholder)

	_ = ctx.GetStub().PutState(Code, unitholderAsBytes)

	return `{"status": 200 , "message": "Unitholder successfully added"}`
}

func (s *SimpleContract) GetAllUnitHolders(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	count := 0
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetStateByRangeWithPagination("UNITHOLDER_"+params[1], "UNITHOLDER_9999999", i, "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		unitholder := unitholder.UnitHolder{}
		json.Unmarshal(queryResponse.Value, &unitholder)
		if unitholder.ObjectType == "unitholder" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false
			buffer.WriteString(string(queryResponse.Value))
			count++

		}
	}
	buffer.WriteString("],")
	total := TotalCount(ctx, "UNITHOLDER_", "UNITHOLDER_9999999", "")
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "UNITHOLDER_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}
func (sc *SimpleContract) UpdateUnitHolderStatus(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id and new status to update"}`
	}
	Code := "UNITHOLDER_" + args[0]
	unitAsBytes, _ := ctx.GetStub().GetState(Code)

	if unitAsBytes == nil {

		return `{"status": 400 , "message": "Unit Holder Data not found for the key provided"}`

	}
	unit := new(unitholder.UnitHolder)
	_ = json.Unmarshal(unitAsBytes, unit)

	if unit.Status == args[1] {
		return `{"status": 400 , "message": "Status is already same as provided"}`

	} else {
		unit.Status = args[1]
	}

	aAsBytes, _ := json.Marshal(unit)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Unit Holder status updated"}`

}

func (sc *SimpleContract) UpdateUnitHolderInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 19 {
		return `{"status": 400 , "message": "Incorrect number of arguments, please provide 19 parameters to update"}`
	}
	Code := "UNITHOLDER_" + args[0]
	unitAsBytes, _ := ctx.GetStub().GetState(Code)

	if unitAsBytes == nil {

		return `{"status": 400 , "message": "Unit Holder Data not found for the key provided"}`

	}
	unit := new(unitholder.UnitHolder)
	_ = json.Unmarshal(unitAsBytes, unit)

	if args[1] != "" {
		unit.AccountName = args[1]
	}
	if args[2] != "" {
		unit.AccountNo = args[2]
	}
	if args[3] != "" {
		unit.BankName = args[3]
	}
	if args[4] != "" {
		unit.FundSelection = args[4]
	}
	if args[5] != "" {
		unit.BalanceUnit = args[5]
	}
	if args[6] != "" {
		unit.Status = args[6]
	}
	if args[7] != "" {
		unit.Cnic = args[7]
	}
	if args[8] != "" {
		unit.Mobile = args[8]
	}
	if args[9] != "" {
		unit.City = args[9]
	}
	if args[10] != "" {
		unit.AccountTitle = args[10]
	}
	if args[11] != "" {
		unit.Nature = args[11]
	}
	if args[12] != "" {
		unit.BranchName = args[12]
	}
	if args[13] != "" {
		unit.AmcName = args[13]
	}
	if args[14] != "" {
		unit.Type = args[14]
	}
	if args[15] != "" {
		unit.Ntn = args[15]
	}
	if args[16] != "" {
		unit.Address = args[16]
	}
	if args[17] != "" {
		unit.Country = args[17]
	}
	if args[18] != "" {
		unit.Upload = args[18]
	}
	if args[19] != "" {
		unit.Name = args[19]
	}

	aAsBytes, _ := json.Marshal(unit)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Unit Holder info updated"}`

}

func (sc *SimpleContract) GetUnitHolderInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting foliono to query"}`
	}
	Code := "UNITHOLDER_" + args[0]
	unitholderAsByte, _ := ctx.GetStub().GetState(Code)
	if unitholderAsByte == nil {

		return `{"status": 400 , "message": "Unit holder Data not found for the key provided"}`
	}

	unitholder := new(unitholder.UnitHolder)
	_ = json.Unmarshal(unitholderAsByte, unitholder)
	s := strings.Split(unitholder.FolioNo, "_")
	unitholder.FolioNo = s[1]
	aAsBytes, _ := json.Marshal(unitholder)
	return `{"status": 200 , "data" : ` + string(aAsBytes) + `}`
}

func (sc *SimpleContract) UpdateUnitHolderUnits(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 3 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting account no amount and flag"}`
	}
	Code := "UNITHOLDER_" + args[0]
	unitholderAsBytes, _ := ctx.GetStub().GetState(Code)

	if unitholderAsBytes == nil {

		return `{"status": 400 , "message": "Unit Holder Data not found for the key provided"}`
	}
	unitholder := new(unitholder.UnitHolder)
	_ = json.Unmarshal(unitholderAsBytes, unitholder)
	balance, _ := strconv.Atoi(unitholder.BalanceUnit)
	amount, _ := strconv.Atoi(args[1])

	if args[2] != "saleofunit" {
		if amount > balance {
			return `{"status": 400 , "message": "Insufficient units for the transaction"}`
		}
		balance = balance - amount
	} else {
		balance = balance + amount
	}
	unitholder.BalanceUnit = strconv.Itoa(balance)
	aAsBytes, _ := json.Marshal(unitholder)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Units updated"}`

}

func (sc *SimpleContract) AddTax(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 10 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 10 arguments for adding a tax."}`
	}
	Code := "TAX_" + params[0]
	taxAsByte, _ := ctx.GetStub().GetState(Code)
	if taxAsByte != nil {
		return `{"status": 400 , "message": "Tax Code already exists "}`

	}
	if CheckQuery(ctx, "tax", "name", params[2]) == "true" {
		return `{"status": 400 , "message": "Tax name already exists"}`
	}
	tax := tax.Tax{
		ObjectType:      "tax",
		TransactionType: params[1],
		Name:            params[2],
		Days:            params[3],
		AmountFrom:      params[4],
		FixedCharges:    params[5],
		Code:            Code,
		CreatedAt:       params[6],
		Description:     params[7],
		AmountTo:        params[8],
		Rate:            params[9],
	}
	taxAsBytes, _ := json.Marshal(tax)

	_ = ctx.GetStub().PutState(Code, taxAsBytes)

	return `{"status": 200 , "message": "tax successfully added"}`
}

func (sc *SimpleContract) UpdateTaxInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 8 {
		return `{"status": 400 , "message": "Incorrect number of arguments,Please provide 8 arguments"}`
	}
	Code := "TAX_" + args[0]
	taxAsBytes, _ := ctx.GetStub().GetState(Code)

	if taxAsBytes == nil {

		return `{"status": 400 , "message": "Tax data not found for the key provided"}`

	}
	tax := new(tax.Tax)
	_ = json.Unmarshal(taxAsBytes, tax)
	if args[1] != "" {
		tax.Name = args[1]
	}
	if args[2] != "" {
		tax.Days = args[2]
	}
	if args[3] != "" {
		tax.AmountFrom = args[3]
	}
	if args[4] != "" {
		tax.AmountTo = args[4]
	}
	if args[5] != "" {
		tax.FixedCharges = args[5]
	}
	if args[6] != "" {
		tax.Description = args[6]
	}
	if args[7] != "" {
		tax.Rate = args[7]
	}

	aAsBytes, _ := json.Marshal(tax)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Tax information updated"}`

}
func (sc *SimpleContract) GetTaxInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to query"}`
	}

	Code := "TAX_" + args[0]
	tax, _ := ctx.GetStub().GetState(Code)
	if tax == nil {

		return `{"status": 400 , "message": "Tax data not found for the key provided"}`

	}

	return `{"status": 200 , "data" : ` + string(tax) + `}`
}

func (s *SimpleContract) GetAllTaxes(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	count := 0
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetStateByRangeWithPagination("TAX_"+params[1], "TAX_9999999", i, "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		tax := tax.Tax{}
		json.Unmarshal(queryResponse.Value, &tax)
		if tax.ObjectType == "tax" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false
			buffer.WriteString(string(queryResponse.Value))
			count++

		}
	}
	buffer.WriteString("],")
	total := TotalCount(ctx, "TAX_", "TAX_9999999", "")
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "TAX_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (s *SimpleContract) GetTaxAmount(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	totalamount := 0.0
	amount := ""
	resultsIterator, err := ctx.GetStub().GetStateByRange("TAX_", "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	defer resultsIterator.Close()
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		tax := tax.Tax{}
		json.Unmarshal(queryResponse.Value, &tax)
		if tax.ObjectType == "tax" && tax.TransactionType == params[0] && tax.Name == params[2] {
			Gross, _ := strconv.Atoi(params[1])
			From, _ := strconv.Atoi(tax.AmountFrom)
			To, _ := strconv.Atoi(tax.AmountTo)
			if Gross >= From && Gross <= To {
				rate, _ := strconv.Atoi(tax.Rate)
				fixed, _ := strconv.Atoi(tax.FixedCharges)
				totalamount = (float64(rate) / 100) * float64(Gross)
				totalamount += float64(fixed)
				amount = fmt.Sprintf("%g", totalamount)

			}
		}
	}
	if amount == "" {
		return `{"status": 200 , "data" : "empty" }`
	} else {
		return `{"status": 200 , "data" : ` + amount + ` }`
	}

}

func (sc *SimpleContract) AddSecurity(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 4 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 4 arguments for adding a security."}`
	}
	Code := "SECURITY_" + params[1]
	securityAsByte, _ := ctx.GetStub().GetState(Code)
	if securityAsByte != nil {
		return `{"status": 400 , "message": "Security code already exists "}`

	}
	security := security.Security{
		ObjectType: "security",
		Name:       params[0],
		Code:       Code,
		Companies:  params[2],
		CreatedAt:  params[3],
	}
	securityAsBytes, _ := json.Marshal(security)

	_ = ctx.GetStub().PutState(Code, securityAsBytes)

	return `{"status": 200 , "message": "security successfully added"}`
}

func (s *SimpleContract) GetAllSecurities(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	count := 0
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetStateByRangeWithPagination("SECURITY_"+params[1], "SECURITY_9999999", i, "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		security := security.Security{}
		json.Unmarshal(queryResponse.Value, &security)
		if security.ObjectType == "security" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false
			buffer.WriteString(string(queryResponse.Value))
			count++

		}
	}
	buffer.WriteString("],")
	total := TotalCount(ctx, "SECURITY_", "SECURITY_9999999", "")
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "SECURITY_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

// Read returns the value at key in the world state
func (sc *SimpleContract) GetSecurityInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to query"}`
	}
	Code := "SECURITY_" + args[0]
	security, _ := ctx.GetStub().GetState(Code)

	if security == nil {

		return `{"status": 400 , "message": "Security Data not found for the key provided"}`
	}

	return `{"status": 200 , "data" : ` + string(security) + `}`
}

func (sc *SimpleContract) AddUser(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 9 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 9 arguments for adding a user."}`
	}
	Code := "USER_" + params[0]

	userAsByte, _ := ctx.GetStub().GetState(Code)
	if userAsByte != nil {

		return `{"status": 400 , "message": "user already exists "}`

	}
	user := user.User{
		ObjectType:   "user",
		Name:         params[1],
		Email:        Code,
		Password:     params[2],
		TwoFaEnabled: params[3],
		Role:         "ROLE_" + params[4],
		Status:       "active",
		TwoFaCode:    params[5],
		UserType:     params[6],
		CreatedAt:    params[7],
		AmcCode:      params[8],
	}
	userAsBytes, _ := json.Marshal(user)

	_ = ctx.GetStub().PutState(Code, userAsBytes)

	return `{"status": 200 , "message": "user successfully added"}`
}

func (s *SimpleContract) GetAllUsers(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	count := 0
	i := int32(page)
	resultsIterator, data, err := ctx.GetStub().GetStateByRangeWithPagination("USER_"+params[1], "", i, "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		user := user.User{}
		json.Unmarshal(queryResponse.Value, &user)
		if len(params) != 3 {
			if user.ObjectType == "user" {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString(string(queryResponse.Value))
				count++

			}
		} else {
			if user.ObjectType == "user" && user.AmcCode != "" {
				if first == false {
					buffer.WriteString(",")
				}
				first = false
				buffer.WriteString(string(queryResponse.Value))
				count++

			}

		}
	}
	buffer.WriteString("],")
	total := TotalCount(ctx, "USER_", "", "")
	if len(params) == 3 {
		total = count
	}
	str := `"page_info":{
	"next_page_number":"` + bookmark + `",
	"total_count":"` + strconv.Itoa(total) + `"
    }`
	buffer.WriteString(str)
	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "USER_", "")
		refined = strings.ReplaceAll(refined, "ROLE_", "")
		return `{"status": 200 , "data" : ` + refined + `}`
	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (s *SimpleContract) GetUsersByRole(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	resultsIterator, err := ctx.GetStub().GetStateByRange("USER_", "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	Role := strings.Split(args[0], ",")
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		user := user.User{}
		json.Unmarshal(queryResponse.Value, &user)
		if len(Role) > 1 {
			if (user.ObjectType == "user" && user.Role == Role[0] && user.AmcCode == args[1]) || (user.ObjectType == "user" && user.Role == Role[1] && user.AmcCode == args[1]) {
				if first == false {
					buffer.WriteString(",")
				}
				first = false

				buffer.WriteString(string(queryResponse.Value))
			}
		} else {
			if user.ObjectType == "user" && user.Role == Role[0] && user.AmcCode == args[1] {
				if first == false {
					buffer.WriteString(",")
				}
				first = false

				buffer.WriteString(string(queryResponse.Value))
			}
		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "USER_", "")
		refined = strings.ReplaceAll(refined, "ROLE_", "")
		return `{"status": 200 , "data" : ` + refined + `}`

	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (sc *SimpleContract) GetUserInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting user email to query"}`
	}

	Code := "USER_" + args[0]
	userAsBytes, _ := ctx.GetStub().GetState(Code)
	if userAsBytes == nil {

		return `{"status": 400 , "message": "User Data not found for the key provided"}`
	}
	user := new(user.User)
	_ = json.Unmarshal(userAsBytes, user)
	user.Email = args[0]
	roleAsBytes, _ := ctx.GetStub().GetState(user.Role)
	if roleAsBytes == nil {

		return `{"status": 400 , "message": "User Role data not found for the key provided"}`
	}
	role := new(roles.Role)
	_ = json.Unmarshal(roleAsBytes, role)

	feature, _ := json.Marshal(role.Features)
	fstring := string(feature)
	user.Features = fstring
	user.RoleStatus = role.Status
	aAsBytes, _ := json.Marshal(user)
	data := string(aAsBytes)
	refined := strings.ReplaceAll(data, "ROLE_", "")
	return `{"status": 200 , "data" : ` + refined + `}`
}

func (sc *SimpleContract) UpdateUserStatus(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id and new status to update"}`
	}
	Code := "USER_" + args[0]
	userAsBytes, _ := ctx.GetStub().GetState(Code)

	if userAsBytes == nil {

		return `{"status": 400 , "message": "User Data not found for the key provided"}`

	}
	user := new(user.User)
	_ = json.Unmarshal(userAsBytes, user)

	if user.Status == args[1] {
		return `{"status": 400 , "message": "Status is already same as provided"}`

	} else {
		user.Status = args[1]
	}

	aAsBytes, _ := json.Marshal(user)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "User status updated"}`

}

func (sc *SimpleContract) UpdateUserPassword(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id and new password to update"}`
	}
	Code := "USER_" + args[0]
	userAsBytes, _ := ctx.GetStub().GetState(Code)

	if userAsBytes == nil {

		return `{"status": 400 , "message": "User Data not found for the key provided"}`

	}
	user := new(user.User)
	_ = json.Unmarshal(userAsBytes, user)

	if user.Password == args[1] {
		return `{"status": 400 , "message": "Password is already same as provided"}`

	} else {
		user.Password = args[1]
	}

	aAsBytes, _ := json.Marshal(user)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Password updated successfully"}`

}

func (sc *SimpleContract) UpdateUser(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 6 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting 6 paramets to update"}`
	}
	Code := "USER_" + args[0]
	userAsBytes, _ := ctx.GetStub().GetState(Code)

	if userAsBytes == nil {

		return `{"status": 400 , "message": "User Data not found for the key provided"}`

	}
	user := new(user.User)
	_ = json.Unmarshal(userAsBytes, user)

	user.Name = args[1]
	user.Password = args[2]
	user.TwoFaEnabled = args[3]
	user.Role = "ROLE_" + args[4]
	user.TwoFaCode = args[5]

	aAsBytes, _ := json.Marshal(user)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "User updated successfully"}`

}

func (sc *SimpleContract) UpdateUserRole(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id and new role to update"}`
	}
	Code := "USER_" + args[0]
	userAsBytes, _ := ctx.GetStub().GetState(Code)

	if userAsBytes == nil {

		return `{"status": 400 , "message": "User data not found for the key provided"}`

	}
	user := new(user.User)
	_ = json.Unmarshal(userAsBytes, user)

	if user.Role == "ROLE_"+args[1] {
		return `{"status": 400 , "message": "Role is already assigned"}`

	} else {
		user.Role = "ROLE_" + args[1]
	}

	aAsBytes, _ := json.Marshal(user)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "User Roles updated"}`

}

func (sc *SimpleContract) AddRole(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 5 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 5 arguments for adding a role."}`
	}
	Code := "ROLE_" + params[0]

	roleAsByte, _ := ctx.GetStub().GetState(Code)
	if roleAsByte != nil {

		return `{"status": 400 , "message": "Role already exists "}`

	}
	role := roles.Role{
		ObjectType:  "role",
		RoleName:    Code,
		Description: params[1],
		Features:    params[2],
		Status:      params[3],
		CreatedAt:   params[4],
	}
	roleAsBytes, _ := json.Marshal(role)

	_ = ctx.GetStub().PutState(Code, roleAsBytes)

	return `{"status": 200 , "message": "Role successfully added"}`
}

func (sc *SimpleContract) UpdateRole(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 4 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting 4 paramets to update Role"}`
	}
	Code := "ROLE_" + args[0]
	roleAsBytes, _ := ctx.GetStub().GetState(Code)

	if roleAsBytes == nil {

		return `{"status": 400 , "message": "Role data not found for the key provided"}`

	}
	role := new(roles.Role)
	_ = json.Unmarshal(roleAsBytes, role)

	role.Description = args[1]
	role.Features = args[2]
	role.Status = args[3]

	aAsBytes, _ := json.Marshal(role)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Role updated successfully"}`

}

func (s *SimpleContract) GetAllRoles(ctx contractapi.TransactionContextInterface) string {
	resultsIterator, err := ctx.GetStub().GetStateByRange("ROLE_", "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		role := roles.Role{}
		json.Unmarshal(queryResponse.Value, &role)
		if role.ObjectType == "role" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false

			buffer.WriteString(string(queryResponse.Value))
		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "ROLE_", "")
		return `{"status": 200 , "data" : ` + refined + `}`

	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}
func (sc *SimpleContract) GetRoleInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to query"}`
	}

	Code := "ROLE_" + args[0]
	amc, _ := ctx.GetStub().GetState(Code)
	if amc == nil {

		return `{"status": 400 , "message": "Role Data not found for the key provided"}`

	}

	return `{"status": 200 , "data" : ` + string(amc) + `}`
}

func CheckQuery(ctx contractapi.TransactionContextInterface, doc_type string, field_name string, value string) string {
	query2 := "{\"selector\":{\"$and\":[{\"doc_type\":{\"$eq\":\"" + doc_type + "\"}},{\"" + field_name + "\":{\"$eq\":\"" + value + "\"}}]}}"
	resultsIterator, _ := ctx.GetStub().GetQueryResult(query2)
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `error`
		}
		if first == false {
			buffer.WriteString(",")
		}
		first = false

		buffer.WriteString(string(queryResponse.Value))
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		return `true`
	}
	return `false`
}
func (s *SimpleContract) GetTransactionsRich(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	page, _ := strconv.Atoi(params[0])
	i := int32(page)
	query := ""
	status := strings.Split(params[2], ",")

	if len(status) < 2 {
		query = "{\"selector\":{\"txn_status\":\"" + params[2] + "\",\"doc_type\":\"transaction\"}, \"use_index\":[\"_design/transactiondoc\", \"indextransaction\"]}"
	} else {
		querry := "{\"selector\":{\"$or\":[{\"txn_status\":\"" + status[0] + "\"},"
		remain := "],\"doc_type\":\"transaction\"},\"use_index\":[\"_design/transactiondoc\",\"indextransaction\"]}"
		for i := 1; i < len(status); i++ {
			if i == (len(status) - 1) {
				querry += "{\"txn_status\":\"" + status[i] + "\"}"
			} else {
				querry += "{\"txn_status\":\"" + status[i] + "\"},"
			}
		}
		querry += remain
		query = querry
	}
	if len(params) == 4 && params[3] != "" || len(params) == 5 && params[3] != "" {
		id := "TXN_" + params[3]
		query = "{\"selector\":{\"txn_id\":\"" + id + "\",\"doc_type\":\"transaction\"}, \"use_index\":[\"_design/transactiondoc\", \"indextransaction\"]}"
	}
	if (params[0] == "" && params[1] == "" && params[2] == "" && params[3] == "") || (params[0] != "" && params[2] == "") {
		query = "{\"selector\":{\"doc_type\":\"transaction\"}, \"use_index\":[\"_design/transactiondoc\", \"indextransaction\"]}"
	}
	if len(params) == 5 && params[4] != "" {
		query = "{\"selector\":{\"created_by\":\"" + params[4] + "\",\"doc_type\":\"transaction\"}, \"use_index\":[\"_design/transactiondoc\", \"indextransaction\"]}"

	}
	resultsIterator, data, _ := ctx.GetStub().GetQueryResultWithPagination(query, i, params[1])
	bookmark := data.GetBookmark()
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		if first == false {
			buffer.WriteString(",")
		}
		first = false

		buffer.WriteString(string(queryResponse.Value))
	}
	total := GetTransactionsCount(ctx)
	buffer.WriteString("],")
	str := `"page_info":{
		"next_page_number":"` + bookmark + `",
		"total_count":"` + total + `"
		}`
	buffer.WriteString(str)

	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "TXN_", "")
		return `{"status": 200 , "data" : ` + refined + `}`

	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}
func GetTransactionsCount(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	query := ""
	page, _ := strconv.Atoi(params[0])
	i := int32(page)
	status := strings.Split(params[2], ",")
	if len(status) < 2 {
		query = "{\"selector\":{\"txn_status\":\"" + params[2] + "\",\"doc_type\":\"transaction\"}, \"use_index\":[\"_design/transactiondoc\", \"indextransaction\"]}"
	} else {
		querry := "{\"selector\":{\"$or\":[{\"txn_status\":\"" + status[0] + "\"},"
		remain := "],\"doc_type\":\"transaction\"},\"use_index\":[\"_design/transactiondoc\",\"indextransaction\"]}"
		for i := 1; i < len(status); i++ {
			if i == (len(status) - 1) {
				querry += "{\"txn_status\":\"" + status[i] + "\"}"
			} else {
				querry += "{\"txn_status\":\"" + status[i] + "\"},"
			}
		}
		querry += remain
		query = querry
	}
	if len(params) == 4 && params[3] != "" {
		id := "TXN_" + params[3]
		query = "{\"selector\":{\"txn_id\":\"" + id + "\",\"doc_type\":\"transaction\"}, \"use_index\":[\"_design/transactiondoc\", \"indextransaction\"]}"
	}
	if (params[0] == "" && params[1] == "" && params[2] == "" && params[3] == "") || (params[0] != "" && params[2] == "") {
		query = "{\"selector\":{\"doc_type\":\"transaction\"}, \"use_index\":[\"_design/transactiondoc\", \"indextransaction\"]}"

	}
	if len(params) == 5 && params[4] != "" {
		query = "{\"selector\":{\"created_by\":\"" + params[4] + "\",\"doc_type\":\"transaction\"}, \"use_index\":[\"_design/transactiondoc\", \"indextransaction\"]}"

	}
	i = 0
	resultsIterator, _, _ := ctx.GetStub().GetQueryResultWithPagination(query, i, "")
	count := 0
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		if first == false {
			buffer.WriteString(",")
		}
		first = false

		buffer.WriteString(string(queryResponse.Value))
		count++
	}
	buffer.WriteString("],")
	return strconv.Itoa(count)
}

func (s *SimpleContract) GetTxnCount(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	resultsIterator, err := ctx.GetStub().GetStateByRange("TXN_", "TXN_9999999999999")
	count := 0
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		txn := transaction.Transaction{}
		json.Unmarshal(queryResponse.Value, &txn)
		if len(args) == 1 {
			if txn.ObjectType == "transaction" && txn.CreatedBy == args[0] {

				if first == false {
					buffer.WriteString(",")
				}
				first = false

				buffer.WriteString(string(queryResponse.Value))
				count++
				fmt.Println("count in 1", count)
			}
		} else {
			if txn.ObjectType == "transaction" && txn.CreatedBy == args[0] && txn.TxnStatus == args[1] {

				if first == false {
					buffer.WriteString(",")
				}
				first = false

				buffer.WriteString(string(queryResponse.Value))
				count++
				fmt.Println("count in 2", count)

			}
		}
	}
	buffer.WriteString("]")
	return `{"status": 200 , "data" : ` + strconv.Itoa(count) + `}`

}

func (sc *SimpleContract) GetTxnInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to query"}`
	}

	Code := "TXN_" + args[0]
	txn, _ := ctx.GetStub().GetState(Code)
	if txn == nil {
		return `{"status": 200 , "data" : ` + "[]" + `}`

	}
	return `{"status": 200 , "data" : ` + string(txn) + `}`
}

func (sc *SimpleContract) UpdateTransactionStatus(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id,updated data to update"}`
	}
	Code := "TXN_" + args[0]
	_ = ctx.GetStub().PutState(Code, []byte(args[1]))

	return `{"status": 200 , "message": "Transaction updated successfully"}`

}

func (sc *SimpleContract) AddJsonData(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id,data to add"}`
	}
	_ = ctx.GetStub().PutState(args[0], []byte(args[1]))
	return `{"status": 200 , "message": "Data added successfully"}`
}

func (sc *SimpleContract) GetPsxInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to query"}`
	}

	Code := "PSX_" + args[0]
	psx, _ := ctx.GetStub().GetState(Code)
	if psx == nil {
		return `{"status": 200 , "data" : ` + "[]" + `}`

	}
	return `{"status": 200 , "data" : ` + string(psx) + `}`
}

func (s *SimpleContract) GetAllTxnByKey(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	start := "TXN_" + params[0]
	end := "TXN_" + params[1]
	resultsIterator, err := ctx.GetStub().GetStateByRange(start, end)
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		transaction := transaction.Transaction{}
		json.Unmarshal(queryResponse.Value, &transaction)
		if transaction.ObjectType == "transaction" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false

			buffer.WriteString(string(queryResponse.Value))
		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "TXN_", "")
		return `{"status": 200 , "data" : ` + refined + `}`

	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (sc *SimpleContract) AddCheckList(ctx contractapi.TransactionContextInterface) string {

	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 2 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 2 arguments for adding a checklist."}`
	}
	Code := "CHECKLIST_" + params[0]
	checklistAsByte, _ := ctx.GetStub().GetState(Code)
	if checklistAsByte != nil {
		return `{"status": 400 , "message": "serial no already exists "}`

	}
	checklist := checklist.CheckList{
		ObjectType: "checklist",
		SerialNo:   Code,
		Message:    params[1],
		Value:      "",
		Comment:    "",
	}
	checklistAsBytes, _ := json.Marshal(checklist)

	_ = ctx.GetStub().PutState(Code, checklistAsBytes)

	return `{"status": 200 , "message": "Checklist successfully added"}`
}

// Read returns the value at key in the world state
func (sc *SimpleContract) GetCheckListInfo(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 1 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id to query"}`
	}
	Code := "CHECKLIST_" + args[0]
	checklistAsByte, _ := ctx.GetStub().GetState(Code)
	if checklistAsByte == nil {

		return `{"status": 400 , "message": "Checklist Data not found for the key provided"}`
	}
	checklist := new(checklist.CheckList)
	_ = json.Unmarshal(checklistAsByte, checklist)
	s := strings.Split(checklist.SerialNo, "_")
	checklist.SerialNo = s[1]
	aAsBytes, _ := json.Marshal(checklist)
	return `{"status": 200 , "data" : ` + string(aAsBytes) + `}`
}

func (s *SimpleContract) GetAllCheckList(ctx contractapi.TransactionContextInterface) string {
	resultsIterator, err := ctx.GetStub().GetStateByRange("CHECKLIST_", "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		bank := bank.Bank{}
		json.Unmarshal(queryResponse.Value, &bank)
		if bank.ObjectType == "checklist" {
			if first == false {
				buffer.WriteString(",")
			}
			first = false

			buffer.WriteString(string(queryResponse.Value))

		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "CHECKLIST_", "")
		return `{"status": 200 , "data" : ` + refined + `}`

	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (sc *SimpleContract) UpdateCheckList(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting serial no and new message to update"}`
	}
	Code := "CHECKLIST_" + args[0]
	checklistAsBytes, _ := ctx.GetStub().GetState(Code)

	if checklistAsBytes == nil {

		return `{"status": 400 , "message": "Checklist Data not found for the key provided"}`

	}
	checklist := new(checklist.CheckList)
	_ = json.Unmarshal(checklistAsBytes, checklist)

	if checklist.Message == args[1] {
		return `{"status": 400 , "message": "Message is already same as provided"}`

	} else {
		checklist.Message = args[1]
	}

	aAsBytes, _ := json.Marshal(checklist)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Checklist updated"}`

}

func (sc *SimpleContract) AddNotification(ctx contractapi.TransactionContextInterface) string {

	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 5 {
		return `{"status": 400 , "message" :"Incorrect no of arguments .Please Provide 5 arguments."}`
	}
	Code := "NOTIFICATION_" + params[0]

	notification := notifications.Notification{
		ObjectType:     "notification",
		UserID:         params[1],
		NotificationID: Code,
		Message:        params[2],
		TxnID:          params[3],
		CreatedAt:      params[4],
		Status:         "0",
	}
	notificationAsBytes, _ := json.Marshal(notification)

	_ = ctx.GetStub().PutState(Code, notificationAsBytes)

	return `{"status": 200 , "message": "notification successfully added"}`
}
func (s *SimpleContract) GetAllNotifications(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()

	resultsIterator, err := ctx.GetStub().GetStateByRange("NOTIFICATION_", "")
	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		notification := notifications.Notification{}
		json.Unmarshal(queryResponse.Value, &notification)
		if notification.ObjectType == "notification" && notification.UserID == params[0] && notification.Status == params[1] {
			if first == false {
				buffer.WriteString(",")
			}
			first = false

			buffer.WriteString(string(queryResponse.Value))

		}
	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		data := string(buffer.Bytes())
		refined := strings.ReplaceAll(data, "NOTIFICATION_", "")
		return `{"status": 200 , "data" : ` + refined + `}`

	}
	return `{"status": 200 , "data" : ` + string(buffer.Bytes()) + `}`
}

func (sc *SimpleContract) UpdateNotification(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	if len(args) != 2 {
		return `{"status": 400 , "message": "Incorrect number of arguments, Expecting id and new status to update"}`
	}
	Code := "NOTIFICATION_" + args[0]
	notAsBytes, _ := ctx.GetStub().GetState(Code)

	if notAsBytes == nil {

		return `{"status": 400 , "message": "Bank Data not found for the key provided"}`

	}
	not := new(notifications.Notification)
	_ = json.Unmarshal(notAsBytes, not)

	if args[1] != "" {
		not.Status = args[1]
	}

	aAsBytes, _ := json.Marshal(not)

	_ = ctx.GetStub().PutState(Code, aAsBytes)

	return `{"status": 200 , "message": "Notification updated"}`

}

func (s *SimpleContract) GetDataCount(ctx contractapi.TransactionContextInterface) string {
	_, args := ctx.GetStub().GetFunctionAndParameters()
	count := 0
	fmt.Println("queryy", args[0])
	resultsIterator, err := ctx.GetStub().GetStateByRange(args[0], args[1])

	if err != nil {
		return `{"status" : 500 , "message": "` + string(err.Error()) + `"}`
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	var first = true
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return `{"status": 500 , "message": "` + string(err.Error()) + `"}`
		}
		if first == false {
			buffer.WriteString(",")
		}
		first = false
		buffer.WriteString(string(queryResponse.Value))
		count++

	}
	buffer.WriteString("]")

	if len(buffer.Bytes()) > 2 {
		return `{"status": 200 , "data" : ` + strconv.Itoa(count) + `}`

	}
	return `{"status": 200 , "data" : ` + strconv.Itoa(count) + `}`
}

func (sc *SimpleContract) AddTransaction(ctx contractapi.TransactionContextInterface) string {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	if len(params) != 64 {
		return `{"status": 400 , "message": "Incorrect number of arguments,provide 64 arguments for adding a transaction"}`
	}
	attrval, ok, err := ctx.GetClientIdentity().GetAttributeValue("email")
	usertype, okk, _ := ctx.GetClientIdentity().GetAttributeValue("user_type")
	var Status = ""
	if params[61] != "" {
		Status = "COMPLETED"
	} else {
		if usertype == "AMC" {
			Status = "CREATED"
		} else {
			Status = "COMPLIANCE"
		}
		if !okk {
			fmt.Println("usertype no attribute")
		}
	}
	var hist []txnhistory.TxnHistory
	history := txnhistory.TxnHistory{
		Note:      "",
		CreatedAt: params[51],
		CreatedBy: attrval,
		UpdatedAt: "",
		UpdatedBy: "",
		Action:    "CREATED",
	}
	hist = append(hist, history)
	if err != nil {
		fmt.Println("error while retrieving")
	}
	if !ok {
		fmt.Println("email no attribute")
	}
	Code := "TXN_" + params[54]
	txn := transaction.Transaction{
		ObjectType:            "transaction",
		TxnId:                 Code,
		TxnRType:              params[0],
		TxnStatus:             Status,
		AmcName:               params[1],
		FundName:              params[2],
		InstructionDate:       params[3],
		ExecutionDate:         params[4],
		ExecutedDate:          params[5],
		FundAccount:           params[6],
		AccountTitle:          params[7],
		AccountNumber:         params[8],
		Bank:                  params[9],
		CounterAccountType:    params[10],
		CounterAccountTitle:   params[11],
		CounterAccountNumber:  params[12],
		CounterBank:           params[13],
		CounterBranch:         params[14],
		ModeOfPayment:         params[15],
		PaymentType:           params[16],
		InstrumentType:        params[17],
		InstrumentNo:          params[18],
		InstrumentDate:        params[19],
		RealizedDate:          params[20],
		GrossAmount:           params[21],
		TotalCharges:          parseValue(params[22]),
		TxnCharges:            params[23],
		NetAmount:             parseValue(params[24]),
		CrAmount:              parseValue(params[25]),
		DrAmount:              parseValue(params[26]),
		Balance:               parseValue(params[27]),
		FolioNo:               params[28],
		UnitHolderName:        params[29],
		Units:                 parseValue(params[30]),
		Nav:                   parseValue(params[31]),
		TxnHistory:            hist,
		CreatedBy:             attrval,
		MadeBy:                usertype,
		SaleDate:              params[32],
		CurrentHolding:        parseValue(params[33]),
		TotalHolding:          parseValue(params[34]),
		Symbol:                params[35],
		DividendPercentage:    parseValue(params[36]),
		CreditDate:            params[37],
		MaturityType:          params[38],
		SecurityType:          params[39],
		IssueDate:             params[40],
		MaturityDate:          params[41],
		CouponRate:            params[42],
		Price:                 parseValue(params[43]),
		FaceValue:             parseValue(params[44]),
		Detail:                params[45],
		Type:                  params[46],
		RedemptionDate:        params[47],
		RemainHolding:         parseValue(params[48]),
		SettlementBy:          params[49],
		TaxType:               params[50],
		CreatedAt:             params[51],
		AssignedTo:            params[52],
		AssignTime:            params[53],
		Broker:                params[55],
		Branch:                params[56],
		ReturnDate:            params[57],
		Securities:            params[58],
		ConversionDate:        params[59],
		AssociatedTransaction: params[60],
		CounterType:           params[62],
		SettlementDate:        params[63],
	}
	txnAsBytes, _ := json.Marshal(txn)
	_ = ctx.GetStub().PutState(Code, txnAsBytes)
	return `{"status": 200 , "message": "Transaction successfully created"}`
}

func parseValue(val string) float32 {
	value, err := strconv.ParseFloat(val, 32)
	if err != nil {
	}
	float := float32(value)
	return float
}
