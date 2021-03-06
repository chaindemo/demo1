/*
Demo Unit Linked Contract chaincode
*/

package main


import (
	"os"
//	"io/ioutil"
	"encoding/json"
	"errors"
	"fmt"
 //       "time" 
//	"strings"
	"log"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/chaindemo/demo1/newulc/shared"
	"net/http" 
 //   	"encoding/binary"
  	"bytes"
)	

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}
//*****************************************
//* Contract scheduler 
var scheduler string
//*****************************************
//* Contract Types
/***
type Fund struct{
 FundId string
 Units  string
}
****/
type Account struct{
  Fnds map[string]string
  LastvaluationDate string
  Valuation string
}
type Life struct{
 Name string
 Gender string
 Dob    string
 Smoker string
}
type Contract struct{
 ContID string
 Acct Account
 Product string
 StartDate string
 SumAssured string
 Term  string
 PaymentFrequency string
 Owner  string
 Beneficiary string
 Lf  Life
 Status string
 Email string
 UWstatus string
}


type Policy struct{
	Cont Contract
}


var policies map[string]string
var lock map[string]string
//*****************************************


var count int
var   xx = shared.Args{1, 2}
var invokeTran string
var url string
var manager string
var commsmanager string
func main() {

	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2"  )
	}
        count=0;
	l := log.New(os.Stderr, "", 0)
	l.Println("*************INIT CHAINCODE Unit Linked ****************")
	commsmanager=args[0]
	url=args[1]
	err := 	stub.PutState("url",[]byte(url) )
	err = 	stub.PutState("commsmanager",[]byte(commsmanager) )
        //fmt.Println( xx.A )
	if err != nil {
		return nil, err
	}
	return nil, err
}

//***********************************************
//* Create a newpolicy
func (t *SimpleChaincode) NewPolicy(stub shim.ChaincodeStubInterface,args []string) ([]byte, error) {
	fmt.Println("************* Unit Linked  New Policy****************")
	if len(args) != 11 {
	 	fmt.Println("Incorrect number of arguments. Expecting 11")
		return nil, errors.New("Incorrect number of arguments. Expecting 11")
	}
        count=0;
	fmt.Println("Policy ID="+stub.GetTxID())
  	var policy Policy 
	if policies == nil {
	  policies=make(map[string]string)
	  //*****************************************
	  // get the mal listing all policies  
	  vb, _ := stub.GetState("policies")
    	  json.Unmarshal(vb , &policies)
         }
	//*****************************************************
	//* set contract number to transaction id for now
	policy.Cont.ContID=stub.GetTxID()

  	policy.Cont.Lf.Gender=args[0]
  	policy.Cont.Lf.Dob=args[1]
  	policy.Cont.Lf.Smoker=args[2]
  	policy.Cont.Product=args[3]
  	policy.Cont.StartDate=args[4]
  	policy.Cont.Term = args[5]
  	policy.Cont.PaymentFrequency=args[6]
  	policy.Cont.Owner=args[7]
  	policy.Cont.Lf.Name=args[8]
  	policy.Cont.Email=args[9]
  	policy.Cont.SumAssured=args[10]
  	policy.Cont.Acct.Valuation="0"
  	policy.Cont.Status="PR"

	// set to ready for now till UW contract is implemented
  	policy.Cont.UWstatus="Ready"
	fmt.Println("Creating New Policy for :"+ policy.Cont.ContID )
	if _ ,ok:=policies[policy.Cont.ContID] ; ok {
		fmt.Println("Contract Already Exist ")
		return nil, errors.New("Contract exists already")
        }

  

	//************************************************
	//* Funds 
	policy.Cont.Acct.Fnds=make(map[string]string)

	policies[policy.Cont.ContID]=policy.Cont.Status
        b, err := json.Marshal(policy)
	err = 	stub.PutState(policy.Cont.ContID, b)
	//*****************************************
	//* Save the stateof  the policies map
        b, err = json.Marshal(policies)
	err = 	stub.PutState("policies", b)

	if err != nil {
		return nil, err
	}
	t.welcome(stub, policy)
	return []byte("Policy Added"), err
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	invokeTran=stub.GetTxID()
	fmt.Println("DE************* Invoke Function "+ function )

	//***********************************************
	// process contract independant functions 
	if function == "init" {
		return  t.Init(stub, "init", args)
	} else if function == "schedule" {
		return t.monthlyProcessing(stub, args )
	} else if function == "NewPolicy" {
		return t. NewPolicy(stub, args)
	}
	//**************************************
	// all remaining functions require a contract number in args[0]
	fmt.Println("invoke for policy " + args[0])
	var policy Policy
	//*****************************************
	// get Contract state 
	valAsbytes, _ := stub.GetState(args[0])
    	json.Unmarshal(valAsbytes , &policy)

        var err error

        //xx = shared.Args{1, 2} 
	// Handle different functions


	if function == "applyPremium" {
		policy,err = t.applyPremium(stub, args, policy)
	} else if function == "surrender" {
		policy,err = t.surrender(stub, args , policy )
	}else{ 
		fmt.Println("invoke did not find func: " + function)
		err=errors.New("Received unknown function invocation: " + function)
        }
	if  policy.Cont.ContID=="" {
		return nil,nil
        }

        //*****************************
        // save policy sate 
        b, err := json.Marshal(policy)
	err = 	stub.PutState(policy.Cont.ContID, b)

	//**********************************
	//* Save Policies map with new policy status
        policies=make(map[string]string )
	valAsbytes, _ = stub.GetState("policies")
    	json.Unmarshal(valAsbytes , &policies)
	
        policies[policy.Cont.ContID]=policy.Cont.Status

        b, err = json.Marshal(policies)
	err = 	stub.PutState("policies", b)

	if err != nil {
		return nil, err
	}

	return nil , err 
}




func (t *SimpleChaincode) surrender(stub shim.ChaincodeStubInterface, args []string , policy Policy ) ( Policy, error) {
        var err error
	return policy , err
}


func (t *SimpleChaincode) applyPremium(stub shim.ChaincodeStubInterface, args []string, policy Policy) (Policy, error) {

	// payment is arg[1]
	premium, _ := strconv.ParseFloat( args[1] , 10);

 	log.Print("DE***** Contract value="+policy.Cont.Acct.Valuation + "Payment="+args[1])


	i, _ := strconv.ParseFloat( policy.Cont.Acct.Valuation , 10);
	i = i + float64(premium)
        policy.Cont.Acct.Valuation= strconv.FormatFloat(i,  'f' , 2,  64)
 	log.Print("DE***** Contract value now="+policy.Cont.Acct.Valuation)

	//*************************************
	//* Now set the policy in force 

	if 	policy.Cont.Status=="IF" {
		//*****************************************************
		// email
		subject:="Thank you for your Payment"
		body:=`Dear Mr `+ policy.Cont.Lf.Name + `#N Thank you for your payment of $` +strconv.FormatFloat(premium,  'f' , 2,  64)+ ` for your policy `+policy.Cont.ContID+` #N Many thanks`
 		t.mailto(stub, subject , body, policy )
	} else{

	  if  policy.Cont.UWstatus=="Ready"{
  		policy.Cont.Status="InForce"
		subject:="Your Policy is now in Force"
		body:=`Dear Mr `+policy.Cont.Lf.Name+ ` #N Thank you for your payment of $`+strconv.FormatFloat(premium,  'f' , 2,  64)+ ` for your new policy #N we are pleased to inform you that your policy is now in force #N Many thanks`
		t.mailto(stub, subject , body , policy )
	  }else{
		//*****************************************************
		// email
		subject:="Thank you for your Payment"
		body:=`Dear Mr `+ policy.Cont.Lf.Name + `#N Thank you for your payment of $` +strconv.FormatFloat(premium,  'f' , 2,  64)+ ` for your policy #N Many thanks`
 		t.mailto(stub, subject , body, policy )
	  }
       }
	policy.Cont.Status="IF"
 	//t.activate(stub, args )
	return  policy, nil
}


func (t *SimpleChaincode) monthlyProcessing(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
        fmt.Println("Starting Scheduled Processing")
	var err error
	  policies=make(map[string]string)
	  //*****************************************
	  // get the mal listing all policies  
	  vb, _ := stub.GetState("policies")
    	  json.Unmarshal(vb , &policies)

   //Iterate over the contracts & process all that are in force
        for key , value := range policies {
        fmt.Println("Sheduler Looking at Policy" + key +" Status="+value)
		if value == "IF" {
		  var policy Policy

		  //*****************************************
		  // get Contract state 
		  valAsbytes, _ := stub.GetState(key)
    		  json.Unmarshal(valAsbytes , &policy)

		  policy , err = t.ProcessPolicy(stub, args , policy)

		  policies[policy.Cont.ContID]=policy.Cont.Status
        	  b, _ := json.Marshal(policy)
		  err =	stub.PutState(key , b)
	        }
	}
        fmt.Println("completed Scheduled Processing")
	return nil , err
}
func (t *SimpleChaincode) ProcessPolicy(stub shim.ChaincodeStubInterface, args []string, policy Policy) (Policy, error) {
		  var err error
		  policy , err = t.ProcessCharges(stub, args , policy)
		  t.statement(stub , args , policy ) 
	return policy, err	
}
func (t *SimpleChaincode) ProcessCharges(stub shim.ChaincodeStubInterface, args []string, policy Policy) (Policy, error) {
	var contract Contract=policy.Cont
        fmt.Println("Scheduled Processing for contract" + policy.Cont.ContID)
	coi:=33
        fmc:=10
	adc:=12
 	fmt.Print("DE***** Contract value="+contract.Acct.Valuation)

	i, err := strconv.ParseFloat( contract.Acct.Valuation , 10);
	i = i - float64(coi+fmc+adc)
        contract.Acct.Valuation= strconv.FormatFloat(i,  'f' , 2,  64)
 	log.Print("DE***** Contract value="+contract.Acct.Valuation)
     	return policy, err  
}


func (t *SimpleChaincode) statement(stub shim.ChaincodeStubInterface, args []string , policy Policy) ([]byte, error) {
	var contract Contract=policy.Cont
	subject:="Your Monthly Statement for "+contract.ContID 
	body:= `Your Statement of account: #N Account Holder:`+contract.Owner+` #N Value:`+contract.Acct.Valuation+` #N Yours Sincerely, Danny `
	t.mailto(stub, subject , body, policy )
 return nil,nil
}

func(t *SimpleChaincode) welcome(stub shim.ChaincodeStubInterface , policy Policy) ([]byte, error) {
	var contract Contract=policy.Cont
	subject:="Thank you for your application"
	body:=`Dear Mr `+ contract.Lf.Name + `#N Policy Number=`+policy.Cont.ContID+`#N Thank you for your application, which has now been accepted #N We will activate your new Policy as soon as payment is received`

 t.mailto(stub, subject , body, policy )
 return nil,nil
}


func (t *SimpleChaincode) mailto(stub shim.ChaincodeStubInterface, subject string, body string , policy Policy ) ([]byte, error) {
	var contract Contract=policy.Cont
 	valAsbytes, err := stub.GetState("scheduler")
	scheduler=string(valAsbytes)

	valAsbytes, err = stub.GetState("url")
	url=string(valAsbytes)


	 var jsonStr = []byte( `{
   	  "jsonrpc": "2.0",
    	 "method": "query",
    	 "params": {
      	   "type": 1,
     	    "chaincodeID": {
      	       "name":"`+commsmanager+`"
         },
         "ctorMsg": {
             "function": "mailto",
             "args": [
                 "`+stub.GetTxID()+`",
		 "`+subject+`",
		 "`+body+`",
		 "`+contract.Email+`"
             ]
         },
         "secureContext": "admin"
     },
     "id": 3
 }` )


    fmt.Println("Send Email:", string(jsonStr) )
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("Content-Type", "application/json")
    //req.Header.Set("Postman-Token", "")
    req.Header.Set("Cache-Control", "no-cache")
    req.Header.Set("accept", "application/json")
    client := &http.Client{}
    resp, err2 := client.Do(req)
    err=err2
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    fmt.Println("Email To Status:", resp.Status)
   
   
     return  []byte("Mail sent"), err
}



func (t *SimpleChaincode) valuation(stub shim.ChaincodeStubInterface, args []string , policy Policy) ([]byte, error) {
	var contract Contract=policy.Cont
	valAsbytes, err := stub.GetState("Contract")
    	json.Unmarshal(valAsbytes , &contract)

	return  []byte("Valuation="+contract.Acct.Valuation), err
}



// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)


	fmt.Println("invoke for policy " + args[0])
	var policy Policy
	//*****************************************
	// get Contract state 
	valAsbytes, _ := stub.GetState(args[0])
    	json.Unmarshal(valAsbytes , &policy)
	// Handle different functions

        if function == "valuation" {
		return t.valuation(stub, args, policy)
	} else if function == "dump" {
		return t.dump(stub, args, policy)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}



// read - query function to read key/value pair
func (t *SimpleChaincode) dump(stub shim.ChaincodeStubInterface, args []string , policy Policy ) ([]byte, error) {
    	valAsbytes, err :=json.Marshal( policy)
	fmt.Println( "POLICY DUMP="+ string(valAsbytes))
	return valAsbytes, err
}
