package mpesaexpresscont

import (
	"bytes"
	"crypto/rand"

	"fmt"

	"encoding/hex"
	"encoding/json"

	"strings"

	"time"

	"io/ioutil"
	"log"
	"net/http"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/jamesmukumu/guzman/work/db"
	mpesaexpress "github.com/jamesmukumu/guzman/work/models/mpesa_express"
	env "github.com/joho/godotenv"
)

type Request_Express struct{
Customers_Phone_Number string `json:"customer_no"`
Amount string `json:"amount"`
}

type Ums_Load struct{
Api_Key string `json:"api_key"`
Email string `json:"email"`

Amount string `json:"amount"`
Msisdn string `json:"msisdn"`
Reference string `json:"reference"`
}   
type Ums_Response struct{
Success string `json:"success"`
Message string `json:"message"`
Transaction_request_Id string `json:"tranasaction_request_id"`
}
type Ums_Verification struct{
Api_Key string `json:"api_key"`
Email string `json:"email"`
Transaction_Id string `json:"tranasaction_request_id"`
}




var Access_Token = make(chan string, 1)
var Reference_String = make(chan string, 1)

func Generate_Hex(){
bytesSize := make([]byte,32)
rand.Read(bytesSize)
data := hex.EncodeToString(bytesSize)
Reference_String <- data
fmt.Println("generated ref is",data)
}

func Initiate_Mpesa_Ums(res http.ResponseWriter,req *http.Request){
env.Load()
var paymentSecret = os.Getenv("paymentsecret")

go Generate_Hex()
var ref string = <-Reference_String
time.Sleep(time.Millisecond * 6)
var client = http.Client{}
var Request_Body Request_Express
var Response_Body Ums_Response
json.NewDecoder(req.Body).Decode(&Request_Body)
if strings.HasPrefix(Request_Body.Customers_Phone_Number,"0") && len(Request_Body.Customers_Phone_Number) == 10{
Request_Body.Customers_Phone_Number = "254"+Request_Body.Customers_Phone_Number[1:]
}

  

var PayloadData = Ums_Load{     
Api_Key: os.Getenv("apisecret"),
Email:os.Getenv("email") ,
Amount:Request_Body.Amount ,
Msisdn: Request_Body.Customers_Phone_Number,
Reference:ref,
}
fmt.Println(PayloadData)
var marshalledPayload,_= json.Marshal(PayloadData)
bytesPayload :=bytes.NewReader(marshalledPayload)
var request,err2 = http.NewRequest("POST","https://api.umeskiasoftwares.com/api/v1/intiatestk",bytesPayload)
if err2 != nil{
log.Fatal(err2.Error())
return
}
request.Header.Add("Content-Type","application/json")
Response,err5:= client.Do(request)
if err5 != nil{
panic(err5.Error())
}
defer Response.Body.Close()

bytesBody,_ := json.Marshal(Request_Body)


databytes,_ := ioutil.ReadAll(Response.Body)
fmt.Println(string(databytes))
errjson := json.Unmarshal(databytes,&Response_Body)
if errjson != nil {
log.Fatal(errjson.Error())
return
}
if Response_Body.Success == "200"{
var token = jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
"merchant_id":string(databytes),
"exp":time.Now().Add(time.Minute * 5).Unix(),
"user_details":string(bytesBody),
})
var tokenString,errtok = token.SignedString([]byte(paymentSecret))
if errtok != nil{
panic(errtok.Error())
}
jsonMessage := map[string]string{
"message":"Payment Initiated",
"token":tokenString,
}
databites,_ := json.Marshal(jsonMessage)
res.Write(databites)
return
}
}







func Validate_Payment(res http.ResponseWriter,req *http.Request){
env.Load()
var userInfo mpesaexpress.Customer_Details
var secretPayment = os.Getenv("paymentsecret")
var secretPaymentByteForm = []byte(secretPayment)
var token = req.Header.Get("Authorization")
if token ==  ""{
jsonMessage := map[string]string{
"message":"Token Could not be found.",
}
Databytes,_ := json.Marshal(jsonMessage)
res.Write(Databytes)
return
}


var actualToken []string = strings.Split(token," ")

Token := actualToken[1]

validityToken,err := jwt.Parse(Token,func(tok *jwt.Token) (interface{}, error) {
return secretPaymentByteForm,nil
})
if err != nil{
log.Fatal(err.Error())   
messageValidity := make(map[string]string, 0)
messageValidity["message"] ="Internal Server Error"
databytes,_ := json.Marshal(messageValidity)
res.Write([]byte(databytes))
return
}
if !validityToken.Valid{
responseJson := map[string]string{
"message":"Token Expired.Cannot Process Payment",
}
jsonResp,_ := json.Marshal(responseJson)
res.Write([]byte(jsonResp))
return
}
Payload := validityToken.Claims.(jwt.MapClaims)
var payloadMap = Payload["merchant_id"].(string)
var userPayloadmap = Payload["user_details"].(string)

var ResponseLoad Ums_Response
var userStruct Request_Express
errj := json.Unmarshal([]byte(payloadMap),&ResponseLoad)
if errj != nil{
panic(errj.Error())
}
errUser := json.Unmarshal([]byte(userPayloadmap),&userStruct)
if errUser != nil{
panic(errUser.Error())
}    
fmt.Println(ResponseLoad)



var payLoadValidation = Ums_Verification{
Transaction_Id: ResponseLoad.Transaction_request_Id,
Email: os.Getenv("email"),
Api_Key: os.Getenv("apisecret"),
}

var client = http.Client{}
databytesPayload,_ := json.Marshal(payLoadValidation)
readerLoad := bytes.NewReader(databytesPayload)
Request,errRequest := http.NewRequest("POST","https://api.umeskiasoftwares.com/api/v1/transactionstatus",readerLoad)
if errRequest != nil {
panic(errRequest.Error())
}



Request.Header.Set("Content-Type","application/json")
Response,errResponse := client.Do(Request)
defer Response.Body.Close()    
if errResponse != nil {
log.Fatal(errResponse.Error())
return   

}

 var dataBittes,_ = ioutil.ReadAll(Response.Body)
 fmt.Println(string(dataBittes))
     

var Response_struct mpesaexpress.Confirmation_Payment_Mpesa
json.Unmarshal(dataBittes,&Response_struct)
msg := make(map[string]string, 0)


fmt.Println(Response_struct.TransactionCode)
switch Response_struct.TransactionCode {
case "1037":
msg["message"] = "User Taking too long to respond.Try again"
databytes,_ := json.Marshal(msg)
res.Write(databytes)
break;	
case "1":
msg["message"] = "Customer has insufficient balance to complete this transaction"
databytes,_ := json.Marshal(msg)
res.Write(databytes)
break
case "1032":
msg["message"] = "Customer has Cancelled this transaction"
databytes,_ := json.Marshal(msg)
res.Write(databytes)
break   

case "0":
Date_Pay := time.Now().Format("2006-01-02")
userInfo.Date_Pay = Date_Pay
userInfo.Checkout_Request_Id = Response_struct.TransactionID
userInfo.Amount = Response_struct.TransactionAmount
userInfo.Customer_Number = Response_struct.Msisdn 
findings := db.Connection.Where("Checkout_Request_Id  = ?",userInfo.Checkout_Request_Id).Find(&userInfo)
if findings.RowsAffected == 1{  
msg["message"] = "Payment was already processed"
databytes,_ := json.Marshal(msg)
res.Write(databytes)
return
}


result := db.Connection.Create(&userInfo)
result1 := db.Connection.Create(&Response_struct)
if result.RowsAffected == 1 && result1.RowsAffected == 1 {
msg["message"] = "Payment Process Successfully and Saved to DB"
msg["customer_number"] = userInfo.Customer_Number
msg["amount"] = userInfo.Amount
databytes,_ := json.Marshal(msg)
res.Write(databytes)
return
}else if result.RowsAffected == 0 || result.Error != nil || result1.RowsAffected == 0 {
msg["message"] = "Payment Process Successfully but not Saved"
databytes,_ := json.Marshal(msg)
res.Write(databytes)   
log.Fatal(result.Error.Error())
}    
default :
msg["message"] ="Internal Server Error"
databytes,_ := json.Marshal(msg)
res.Write(databytes)

}   


   
      

}   