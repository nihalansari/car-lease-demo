package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"regexp"
)

var logger = shim.NewLogger("CLDChaincode")

//==============================================================================================================================
//	 Status types - Asset lifecycle is broken down into 5 statuses, this is part of the business logic to determine what can 
//					be done to the vehicle at points in it's lifecycle
//==============================================================================================================================

const   STATE_TEMPLATE  			=  0
const   STATE_MANUFACTURE  			=  1
const   STATE_PRIVATE_OWNERSHIP 	=  2
const   STATE_LEASED_OUT 			=  3
const   STATE_BEING_SCRAPPED  		=  4

const   STATE_ONROUTE 				= 11
const   STATE_DAMAGED  				= 12
const   STATE_DELIVERED 			= 100



//==============================================================================================================================
//	 Structure Definitions 
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}


//==============================================================================================================================
//	CargoPack - Defines the structure for a logistic package object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type CargoPack struct {
	
	//Note: commented descriptions are mapping of logistics properties to existing car lease properties 
	// so that we have to do minimal change to the existing code
	
	//package type
	Type            string `json:"type"`
	//particulars
	Particulars     string `json:"particulars"`
	//source
	SourceCity      string `json:"sourceCity"`
	//destination
	DestCity      	string `json:"destCity"`
	//weight
	Weight          int    `json:"weight"`					
	//owner
	Owner           string `json:"owner"`
	//Y/N delivered
	Delivered       int   `json:"delivered"`
	//status(item condition)
	Status          int    `json:"status"`
	//last location
	LastLocation   	string `json:"lastLocation"`
	//Package ID
	V5cID           string `json:"v5cID"`
	//date of dispatch
	DispatchDate 	string `json:"dispatchDate"`
	//date of delivery
	DeliveredDate 	string `json:"deliveredDate"`
	//Dimensions
	Dimensions 		string `json:"dimensions"`
	
}

//==============================================================================================================================
//	V5C Holder - Defines the structure that holds all the v5cIDs for vehicles that have been created.
//				Used as an index when querying all vehicles.
//==============================================================================================================================

type V5C_Holder struct {
	V5Cs 	[]string `json:"v5cs"`
}
	
	
//==============================================================================================================================
//	User_and_eCert - Struct for storing the JSON of a user and their ecert
//==============================================================================================================================

type User_and_eCert struct {
	Identity string `json:"identity"`
	eCert string `json:"ecert"`
}



//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode																	
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	
	//Args
	//				0
	//			peer_address

	var v5cIDs V5C_Holder

	bytes, err := json.Marshal(v5cIDs)

    if err != nil { return nil, errors.New("Error creating V5C_Holder record") }

	err = stub.PutState("v5cIDs", bytes)

	for i:=0; i < len(args); i=i+2 {
		t.add_ecert(stub, args[i], args[i+1])
	}

	return nil, nil
}

//*100
//==============================================================================================================================
//	 General Functions
//==============================================================================================================================
//	 get_ecert - Takes the name passed and calls out to the REST API for HyperLedger to retrieve the ecert
//				 for that user. Returns the ecert as retrived including html encoding.
//==============================================================================================================================
func (t *SimpleChaincode) get_ecert(stub shim.ChaincodeStubInterface, name string) ([]byte, error) {

	ecert, err := stub.GetState(name)

	if err != nil { return nil, errors.New("Couldn't retrieve ecert for user " + name) }

	return ecert, nil
}

//==============================================================================================================================
//	 add_ecert - Adds a new ecert and user pair to the table of ecerts
//==============================================================================================================================

func (t *SimpleChaincode) add_ecert(stub shim.ChaincodeStubInterface, name string, ecert string) ([]byte, error) {


	err := stub.PutState(name, []byte(ecert))

	if err == nil {
		return nil, errors.New("Error storing eCert for user " + name + " identity: " + ecert)
	}

	return nil, nil

}

//==============================================================================================================================
//	 get_caller - Retrieves the username of the user who invoked the chaincode.
//				  Returns the username as a string.
//==============================================================================================================================

func (t *SimpleChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

    username, err := stub.ReadCertAttribute("username");
	if err != nil { return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error()) }
	return string(username), nil
}

//==============================================================================================================================
//	 check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
// 				  		certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) check_affiliation(stub shim.ChaincodeStubInterface) (string, error) {
    affiliation, err := stub.ReadCertAttribute("role");
	if err != nil { return "", errors.New("Couldn't get attribute 'role'. Error: " + err.Error()) }
	return string(affiliation), nil

}

//==============================================================================================================================
//	 get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *SimpleChaincode) get_caller_data(stub shim.ChaincodeStubInterface) (string, string, error){

	user, err := t.get_username(stub)

    // if err != nil { return "", "", err }

	// ecert, err := t.get_ecert(stub, user);

    // if err != nil { return "", "", err }

	affiliation, err := t.check_affiliation(stub);

    if err != nil { return "", "", err }

	return user, affiliation, nil
}

//==============================================================================================================================
//	 retrieve_v5c - Gets the state of the data at v5cID in the ledger then converts it from the stored 
//					JSON into the Vehicle struct for use in the contract. Returns the Vehcile struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_v5c(stub shim.ChaincodeStubInterface, v5cID string) (CargoPack, error) {
	
	var v CargoPack

	bytes, err := stub.GetState(v5cID);					
				
															if err != nil {	fmt.Printf("RETRIEVE_V5C: Failed to invoke vehicle_code: %s", err); return v, errors.New("RETRIEVE_V5C: Error retrieving vehicle with v5cID = " + v5cID) }
	
	err = json.Unmarshal(bytes, &v);						

															if err != nil {	fmt.Printf("RETRIEVE_V5C: Corrupt vehicle record "+string(bytes)+": %s", err); return v, errors.New("RETRIEVE_V5C: Corrupt vehicle record"+string(bytes))	}
	
	return v, nil
}

//==============================================================================================================================
// save_changes - Writes to the ledger the Vehicle struct passed in a JSON format. Uses the shim file's 
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, v CargoPack) (bool, error) {
	 
	bytes, err := json.Marshal(v)
	
																if err != nil { fmt.Printf("SAVE_CHANGES: Error converting vehicle record: %s", err); return false, errors.New("Error converting vehicle record") }

	err = stub.PutState(v.V5cID, bytes)
	
																if err != nil { fmt.Printf("SAVE_CHANGES: Error storing vehicle record: %s", err); return false, errors.New("Error storing vehicle record") }
	
	return true, nil
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	
	var i1 int

	caller, caller_affiliation, err := t.get_caller_data(stub)
	
	if err != nil { fmt.Printf("Error retrieving caller information")}
	
	fmt.Printf("function: ", function)
    fmt.Printf("caller: ", caller)
    fmt.Printf("affiliation: ", caller_affiliation)

	
	
	
	if function == "create_package" { return t.create_package(stub, caller, args[0])
	} else if function == "ping" {
        return t.ping(stub)
    } else { 																				// If the function is not a create then there must be a car so we need to retrieve the car.
		
		argPos := 1
		
		if function == "deliver_package" {																// If its a scrap vehicle then only two arguments are passed (no update value) all others have three arguments and the v5cID is expected in the last argument
			argPos = 0
		}
		
		v, err := t.retrieve_v5c(stub, args[argPos])
		
																							if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); return nil, errors.New("Error retrieving v5c") }
																		
		if strings.Contains(function, "update") == false           && 
		   function 							!= "deliver_package"    { 									// If the function is not an update or a scrappage it must be a transfer so we need to get the ecert of the recipient.
			
				
				if 		   function == "authority_to_manufacturer" { return t.authority_to_manufacturer(stub, v, caller, args[0])
				} else if  function == "manufacturer_to_private"   { return t.manufacturer_to_private(stub, v, caller,  args[0])
				} else if  function == "private_to_private" 	   { return t.private_to_private(stub, v, caller,  args[0])
				} else if  function == "private_to_lease_company"  { return t.private_to_lease_company(stub, v, caller,  args[0])
				} else if  function == "lease_company_to_private"  { return t.lease_company_to_private(stub, v, caller,  args[0])
				} else if  function == "private_to_scrap_merchant" { return t.private_to_scrap_merchant(stub, v, caller,  args[0])
				}
			
		} else if function == "update_type"  	    	{ return t.update_type(stub, v, caller,  args[0])
		} else if function == "update_particulars"      { return t.update_particulars(stub, v, caller,  args[0])
		} else if function == "update_sourcecity" 		{ return t.update_sourcecity(stub, v, caller,  args[0])
		} else if function == "update_destcity" 		{ return t.update_destcity(stub, v, caller,  args[0])
		} else if function == "update_weight" 			{ 	i1, err = strconv.Atoi(args[0])
															return t.update_weight(stub, v, caller,  i1) 
		} else if function == "update_owner" 			{ return t.update_owner(stub, v, caller,  args[0])
		} else if function == "update_delivered"  	    { 	i1, err = strconv.Atoi(args[0])
															return t.update_delivered(stub, v, caller,  i1)
		} else if function == "update_status"        	{   i1, err = strconv.Atoi(args[0])
															return t.update_status(stub, v, caller,  i1)
		} else if function == "update_lastlocation" 	{ return t.update_lastlocation(stub, v, caller,  args[0])
		} else if function == "update_dispatchdate" 	{ return t.update_dispatchdate(stub, v, caller,  args[0])
		} else if function == "update_delivereddate" 	{ return t.update_delivereddate(stub, v, caller,  args[0])
		} else if function == "update_dimensions" 		{ return t.update_dimensions(stub, v, caller,  args[0])
		
		} else if function == "deliver_package" 		{ return t.deliver_package(stub, v, caller) }
		
																						return nil, errors.New("Function of that name doesn't exist.")
			
	}
}
//=================================================================================================================================	
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================	
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
													
	fmt.Printf("Nihal Copy of chaincode running!")
	caller, caller_affiliation, err := t.get_caller_data(stub)
	
	if err != nil { fmt.Printf("Error retrieving caller information")}

	fmt.Printf("function: ", function)
    fmt.Printf("caller: ", caller)
    fmt.Printf("affiliation: ", caller_affiliation)
													
////******************************************************************* commented on 07 dec 2016														
//	if function == "get_package_details" { 
//	
//			if len(args) != 1 { fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed") }
//	
//	
//			v, err := t.retrieve_v5c(stub, args[0])
//																							if err != nil { fmt.Printf("QUERY: Error retrieving v5c: %s", err); return nil, errors.New("QUERY: Error retrieving v5c "+err.Error()) }
//	
//			return t.get_package_details(stub, v, caller)
//			
//	} else if function == "get_packages" {
//			return t.get_packages(stub, caller)
//	} 
//	return nil, errors.New("Received unknown function invocation")
//*******************************************************************

	if function == "get_package_details" {
		if len(args) != 1 { fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed") }
		v, err := t.retrieve_v5c(stub, args[0])
		if err != nil { fmt.Printf("QUERY: Error retrieving v5c: %s", err); return nil, errors.New("QUERY: Error retrieving v5c "+err.Error()) }
		return t.get_package_details(stub, v, caller)
	} else if function == "check_unique_v5c" {
		return t.check_unique_v5c(stub, args[0], caller, caller_affiliation)
	} else if function == "get_packages" {
		return t.get_packages(stub, caller)
	} else if function == "get_ecert" {
		return t.get_ecert(stub, args[0])
	} else if function == "ping" {
		return t.ping(stub)
	}

	return nil, errors.New("Received unknown function invocation " + function)

}
//=================================================================================================================================
//	 Ping Function
//=================================================================================================================================
//	 Pings the peer to keep the connection alive
//=================================================================================================================================
func (t *SimpleChaincode) ping(stub shim.ChaincodeStubInterface) ([]byte, error) {
	return []byte("Hello, world!"), nil
}


//=================================================================================================================================
//	 Create Function
//=================================================================================================================================									
//	 Create Vehicle - Creates the initial JSON for the vehcile and then saves it to the ledger.									
//=================================================================================================================================
func (t *SimpleChaincode) create_package(stub shim.ChaincodeStubInterface, caller string, v5cID string) ([]byte, error) {								
fmt.Printf("Nihal Copy of chaincode running!")
	var v CargoPack																																										
	
	v5c_ID          := "\"v5cID\":\""+v5cID+"\", "							// Variables to define the JSON
	type2           := "\"Type\":\"UNDEFINED\", "
	particulars2    := "\"Particulars\":\"UNDEFINED\", "
	sourcecity2     := "\"SourceCity\":\"UNDEFINED\", "
	destcity2       := "\"DestCity\":\"UNDEFINED\", "
	weight2         := "\"Weight\":0, "
	owner2          := "\"Owner\":\""+caller+"\", "
	delivered2    	:= "\"Delivered\":0, "
	status2     	:= "\"Status\":1, "
	lastlocation2   := "\"LastLocation\":\"UNDEFINED\", "
	dispatchdate2   := "\"DispatchDate\":\"mm-dd-yyyy\", "
	delivereddate2  := "\"DeliveredDate\":\"mm-dd-yyyy\", "
	dimensions2 	:= "\"Dimensions\":\"hh-ww-ll\" "
	
	
	// Concatenate the variables to create the total JSON object
	vehicle_json := "{"+v5c_ID+type2+particulars2+sourcecity2+destcity2+weight2+owner2+delivered2+status2+lastlocation2+dispatchdate2+delivereddate2+dimensions2+"}" 	
	
	
	fmt.Printf("vehicle_json=%s",vehicle_json)
	
	matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(v5cID))  				// matched = true if the v5cID passed fits format of two letters followed by seven digits
	
												if err != nil { fmt.Printf("CREATE_VEHICLE: Invalid v5cID: %s", err); return nil, errors.New("Invalid v5cID") }
	
	if 				v5c_ID  == "" 	 || 
					matched == false    {
																		fmt.Printf("CREATE_VEHICLE: Invalid v5cID provided=" + v5cID)
																		return nil, errors.New("Invalid v5cID provided=" + v5cID)
	}

	err = json.Unmarshal([]byte(vehicle_json), &v)							// Convert the JSON defined above into a vehicle object for go
	
																		if err != nil { return nil, errors.New("Invalid JSON object") }

	record, err := stub.GetState(v.V5cID) 								// If not an error then a record exists so cant create a new car with this V5cID as it must be unique
	
																		if record != nil { return nil, errors.New("Vehicle already exists") }

//	if 	caller_affiliation != AUTHORITY {							// Only the regulator can create a new v5c
//
//		return nil, errors.New(fmt.Sprintf("Permission Denied. create_vehicle. %v === %v", caller_affiliation, AUTHORITY))

//	}

	
	_, err  = t.save_changes(stub, v)									
			
																		if err != nil { fmt.Printf("CREATE_VEHICLE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	bytes, err := stub.GetState("v5cIDs")

																		if err != nil { return nil, errors.New("Unable to get v5cIDs") }
																		
	var v5cIDs V5C_Holder
	
	err = json.Unmarshal(bytes, &v5cIDs)
	
																		if err != nil {	return nil, errors.New("Corrupt V5C_Holder record") }
															
	v5cIDs.V5Cs = append(v5cIDs.V5Cs, v5cID)
	
	
	bytes, err = json.Marshal(v5cIDs)
	
															if err != nil { fmt.Print("Error creating V5C_Holder record") }

	err = stub.PutState("v5cIDs", bytes)

															if err != nil { return nil, errors.New("Unable to put the state") }
	
	return nil, nil

}

//=================================================================================================================================
//	 Transfer Functions
//=================================================================================================================================
//	 authority_to_manufacturer
//=================================================================================================================================
func (t *SimpleChaincode) authority_to_manufacturer(stub shim.ChaincodeStubInterface, v CargoPack, caller string, recipient_name string) ([]byte, error) {
	
	if     	v.Status				== STATE_TEMPLATE	&&
			v.Owner					== caller			&&
			v.Delivered				== 0			{		// If the roles and users are ok 
	
					v.Owner  = recipient_name		// then make the owner the new owner
					v.Status = STATE_MANUFACTURE			// and mark it in the state of manufacture
	
	} else {									// Otherwise if there is an error
	
															fmt.Printf("AUTHORITY_TO_MANUFACTURER: Permission Denied");
															return nil, errors.New("Permission Denied")
	
	}
	
	_, err := t.save_changes(stub, v)						// Write new state

															if err != nil {	fmt.Printf("AUTHORITY_TO_MANUFACTURER: Error saving changes: %s", err); return nil, errors.New("Error saving changes")	}
														
	return nil, nil									// We are Done
	
}

//=================================================================================================================================
//	 manufacturer_to_private
//=================================================================================================================================
func (t *SimpleChaincode) manufacturer_to_private(stub shim.ChaincodeStubInterface, v CargoPack, caller string, recipient_name string) ([]byte, error) {
	
	if 		v.Type 	 		== "UNDEFINED" || 					
			v.Particulars   == "UNDEFINED" || 
			v.Dimensions 	== "UNDEFINED" || 
			v.SourceCity    == "UNDEFINED" || 
			v.Weight        == 0				{	//If any part of the car is undefined it has not bene fully manufacturered so cannot be sent
															fmt.Printf("MANUFACTURER_TO_PRIVATE: Car not fully defined")
															return nil, errors.New("Car not fully defined")
	}
	
	if 		v.Status				== STATE_MANUFACTURE	&& 
			v.Owner					== caller				&& 
			v.Delivered     == 0							{
			
					v.Owner = recipient_name
					v.Status = STATE_PRIVATE_OWNERSHIP
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("MANUFACTURER_TO_PRIVATE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 private_to_private
//=================================================================================================================================
func (t *SimpleChaincode) private_to_private(stub shim.ChaincodeStubInterface, v CargoPack, caller string, recipient_name string) ([]byte, error) {
	
	//if 		v.Status				== STATE_PRIVATE_OWNERSHIP	&&
	//		v.Owner					== caller					&&
	//		v.Delivered				== 0					{
	//		
					v.Owner = recipient_name
	//				
	//} else {
	//	
	//														return nil, errors.New("Permission denied")
	//
	//}
	//
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("PRIVATE_TO_PRIVATE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 private_to_lease_company
//=================================================================================================================================
func (t *SimpleChaincode) private_to_lease_company(stub shim.ChaincodeStubInterface, v CargoPack, caller string, recipient_name string) ([]byte, error) {
	
	if 		v.Status				== STATE_PRIVATE_OWNERSHIP	&& 
			v.Owner					== caller					&& 
			v.Delivered     			== 0					{
		
					v.Owner = recipient_name
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
															if err != nil { fmt.Printf("PRIVATE_TO_LEASE_COMPANY: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 lease_company_to_private
//=================================================================================================================================
func (t *SimpleChaincode) lease_company_to_private(stub shim.ChaincodeStubInterface, v CargoPack, caller string, recipient_name string) ([]byte, error) {
	
	if		v.Status				== STATE_PRIVATE_OWNERSHIP	&&
			v.Owner  				== caller					&& 
			v.Delivered				== 0					{
		
				v.Owner = recipient_name
	
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
															if err != nil { fmt.Printf("LEASE_COMPANY_TO_PRIVATE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}

//=================================================================================================================================
//	 private_to_scrap_merchant
//=================================================================================================================================
func (t *SimpleChaincode) private_to_scrap_merchant(stub shim.ChaincodeStubInterface, v CargoPack, caller string, recipient_name string) ([]byte, error) {
	
	if		v.Status				== STATE_PRIVATE_OWNERSHIP	&&
			v.Owner					== caller					&& 
			v.Delivered				== 0					{
			
					v.Owner = recipient_name
					v.Status = STATE_BEING_SCRAPPED
	
	} else {
		
															return nil, errors.New("Permission denied")
	
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("PRIVATE_TO_SCRAP_MERCHANT: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
	
}



//=================================================================================================================================
//	 Read Functions
//=================================================================================================================================
//	 get_package_details
//=================================================================================================================================
func (t *SimpleChaincode) get_package_details(stub shim.ChaincodeStubInterface, v CargoPack, caller string) ([]byte, error) {
	fmt.Printf("Nihal Copy of chaincode running!")
	bytes, err := json.Marshal(v)
	
	if err == nil 
	{
		return bytes, nil 
		
	}
	return nil, errors.New("GET_VEHICLE_DETAILS: Invalid vehicle object") 
	
	
	//if 		v.Owner	== caller		{
	//			
	//				return bytes, nil		
	//} else {
	//															return nil, errors.New("Permission Denied")	
	//}

}

//=================================================================================================================================
//	 get_package_details
//=================================================================================================================================

func (t *SimpleChaincode) get_packages(stub shim.ChaincodeStubInterface, caller string) ([]byte, error) {

fmt.Printf("Nihal Copy of chaincode running!")
	bytes, err := stub.GetState("v5cIDs")
		
																			if err != nil { return nil, errors.New("Unable to get v5cIDs") }
																	
	var v5cIDs V5C_Holder
	
	err = json.Unmarshal(bytes, &v5cIDs)						
	
																			if err != nil {	return nil, errors.New("Corrupt V5C_Holder") }
	
	result := "["
	
	var temp []byte
	var v CargoPack
	
	for _, v5c := range v5cIDs.V5Cs {
		
		v, err = t.retrieve_v5c(stub, v5c)
		
		if err != nil {return nil, errors.New("Failed to retrieve V5C")}
		
		temp, err = t.get_package_details(stub, v, caller)
		
		if err == nil {
			result += string(temp) + ","	
		}
	}
	
	if len(result) == 1 {
		result = "[]"
	} else {
	
		result = result[:len(result)-1] + "]"
	}
	
	return []byte(result), nil
}


//**** UPDATE functions
//**** 

//=================================================================================================================================
//	 update_type
//=================================================================================================================================
func (t *SimpleChaincode) update_type(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value string) ([]byte, error) {
		fmt.Printf("update_type called")
		
		if 	v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.Type = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil

}

//=================================================================================================================================
//	 update_particulars
//=================================================================================================================================

func (t *SimpleChaincode) update_particulars(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value string) ([]byte, error) {
	
	fmt.Printf("update_particulars called")
	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.Particulars = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}


//=================================================================================================================================
//	 update_sourcecity
//=================================================================================================================================
func (t *SimpleChaincode) update_sourcecity(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value string) ([]byte, error) {
	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.SourceCity = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}

//=================================================================================================================================
//	 update_particulars
//=================================================================================================================================

func (t *SimpleChaincode) update_destcity(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value string) ([]byte, error) {
	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.DestCity = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}

//=================================================================================================================================
//	 update_weight
//=================================================================================================================================
func (t *SimpleChaincode) update_weight(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value int) ([]byte, error) {
	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.Weight = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}

//=================================================================================================================================
//	 update_owner ** NOTE: This method may not be used as the ownership is update is handled by transfer functions defined 
// 					       below in this program.
//=================================================================================================================================

func (t *SimpleChaincode) update_owner(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value string) ([]byte, error) {
	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.Owner = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}

//=================================================================================================================================
//	 update_delivered
//=================================================================================================================================
func (t *SimpleChaincode) update_delivered(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value int) ([]byte, error) {
	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.Delivered = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}

//=================================================================================================================================
//	 update_status
//=================================================================================================================================

func (t *SimpleChaincode) update_status(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value int) ([]byte, error) {
	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.Status = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}

//=================================================================================================================================
//	 update_lastlocation
//=================================================================================================================================
func (t *SimpleChaincode) update_lastlocation(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value string) ([]byte, error) {
	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.LastLocation = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}

//=================================================================================================================================
//	 update_dispatchdate
//=================================================================================================================================

func (t *SimpleChaincode) update_dispatchdate(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value string) ([]byte, error) {
		if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner					== caller				&& 
			v.Delivered				== 0				{
			
					v.DispatchDate = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}

//=================================================================================================================================
//	 update_delivereddate
//=================================================================================================================================
func (t *SimpleChaincode) update_delivereddate(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value string) ([]byte, error) {
	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.DeliveredDate = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}

//=================================================================================================================================
//	 update_dimensions
//=================================================================================================================================

func (t *SimpleChaincode) update_dimensions(stub shim.ChaincodeStubInterface, v CargoPack, caller string, new_value string) ([]byte, error) {
	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
			
					v.Dimensions = new_value
					
	} else {
															return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }
	
	return nil, nil
}

//=================================================================================================================================
//	 deliver_package
//=================================================================================================================================
func (t *SimpleChaincode) deliver_package(stub shim.ChaincodeStubInterface, v CargoPack,caller string) ([]byte, error) {
fmt.Printf("Nihal Copy of chaincode running!")
	if		v.Status			== STATE_BEING_SCRAPPED	&& 
			v.Owner				== caller				&& 
			v.Delivered			== 0				{
		
					v.Delivered = 1
				
	} else {
		return nil, errors.New("Permission denied")
	}
	
	_, err := t.save_changes(stub, v)
	
															if err != nil { fmt.Printf("SCRAP_VEHICLE: Error saving changes: %s", err); return nil, errors.New("SCRAP_VEHICLError saving changes") }
	
	return nil, nil
	
}


//=================================================================================================================================
//	 check_unique_v5c
//=================================================================================================================================
func (t *SimpleChaincode) check_unique_v5c(stub shim.ChaincodeStubInterface, v5c string, caller string, caller_affiliation string) ([]byte, error) {
	_, err := t.retrieve_v5c(stub, v5c)
	if err == nil {
		return []byte("false"), errors.New("V5C is not unique")
	} else {
		return []byte("true"), nil
	}
}

//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))

															if err != nil { fmt.Printf("Error starting Chaincode: %s", err) }
}
