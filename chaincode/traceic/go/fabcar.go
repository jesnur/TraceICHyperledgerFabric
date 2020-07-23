/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/*
 * The sample smart contract for documentation topic:
 * Writing Your First Blockchain Application
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the ic structure, with 4 properties.  Structure tags are used by encoding/json library
type IC struct {
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
	CRP        string `json:"crp"`
	Owner      string `json:"owner"`
}

/*
 * The Init method is called when the Smart Contract "fabic" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabic"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryIC" {
		return s.queryIC(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "createIC" {
		return s.createIC(APIstub, args)
	} else if function == "queryAllICs" {
		return s.queryAllICs(APIstub)
	} else if function == "queryAllICsAttr" {
		return s.queryAllICsAttr(APIstub)
	} else if function == "changeICOwner" {
		return s.changeICOwner(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryIC(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	icAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(icAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	ics := []IC{
		IC{Identifier: "a", Type: "laptop", CRP: "1234", Owner: "Org1"},
		IC{Identifier: "b", Type: "desktop", CRP: "2345", Owner: "Org1"},

		IC{Identifier: "c", Type: "server", CRP: "3456", Owner: "Org2"},

		IC{Identifier: "a", Type: "phone", CRP: "4567", Owner: "Org2"},
		IC{Identifier: "e", Type: "iot", CRP: "5678", Owner: "Org2"},
		IC{Identifier: "f", Type: "network", CRP: "6789", Owner: "Org1"},

		IC{Identifier: "g", Type: "laptop", CRP: "7890", Owner: "Org1"},
	}

	i := 0
	for i < len(ics) {
		fmt.Println("i is ", i)
		icAsBytes, _ := json.Marshal(ics[i])
		APIstub.PutState("IC"+strconv.Itoa(i), icAsBytes)
		fmt.Println("Added", ics[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createIC(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	//check the MSPID of the sender
	mspid, err := cid.GetMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}

	//apply the condition that the designer/manufacturer can register a new IC
	if mspid != "Org1MSP" {
		return shim.Error("Assert error")
	}

	//get the id of the sender
	id, err := cid.GetID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	//when the designer/manufacturer creates an IC, it puts itself as the owner
	//var ic = IC{Identifier: args[1], Type: args[2], CRP: args[3], Owner: args[4]}
	var ic = IC{Identifier: args[1], Type: args[2], CRP: args[3], Owner: id}

	icAsBytes, _ := json.Marshal(ic)
	APIstub.PutState(args[0], icAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) queryAllICs(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "IC0"
	endKey := "IC999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllICs:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryAllICsAttr(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Check GetAttributeValue
	icidentifier, identifierfound, err := cid.GetAttributeValue(APIstub, "icidentifier")
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("IC Identifier found : ", identifierfound)
	fmt.Println("IC Identifier is : ", icidentifier)

	// Check AssertAttibuteValue
	identifierErr := cid.AssertAttributeValue(APIstub, "icidentifier", "a")
	if identifierErr != nil {
		return shim.Error("Assert error")
	}

	startKey := "IC0"
	endKey := "IC999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var objIC IC
		json.Unmarshal(queryResponse.Value, &objIC)
		fmt.Println("IC identifier in loop : ", objIC.Identifier)

		if strings.Contains(objIC.Identifier, icidentifier) == true {
			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Record\":")
			// Record is a JSON object, so we write as-is
			buffer.WriteString(string(queryResponse.Value))
			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllICs:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) changeICOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	id, err := cid.GetID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("Message sender id is : ", id)

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	icAsBytes, _ := APIstub.GetState(args[0])
	ic := IC{}

	json.Unmarshal(icAsBytes, &ic)
	//check the genuineness of the ownership of the IC
	//only the owner of the IC can transfer the ownership
	if ic.Owner != id {
		return shim.Error("Ownership transfer not permitted by the sender...")
	}
	//now change the owner to new owner
	ic.Owner = args[1]

	icAsBytes, _ = json.Marshal(ic)
	APIstub.PutState(args[0], icAsBytes)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
