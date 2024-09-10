package mpesaexpresscont

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

type Body_Express struct{    
BusinessShortCode string    
	Password string  
	Timestamp string  
	TransactionType string  
	Amount string 
	PartyA  string
	PartyB string
	PhoneNumber string   
	CallBackURL  string 
	AccountReference string
	TransactionDesc string
 }
type Payload_Request_Validation struct{    
BusinessShortCode string  
Password string    
Timestamp string    
CheckoutRequestID string   
 } 


var Access_Token = make(chan string, 1)

func GenerateAcess_Token(){
var client = http.Client{}
env.Load()
var consumerKey = os.Getenv("consumerKey")
var consumerSecret = os.Getenv("consumerSecret")
combined := consumerKey+":"+consumerSecret
databytes := base64.StdEncoding.EncodeToString([]byte(combined))

var request,err = http.NewRequest("GET","https://sandbox.safaricom.co.ke/oauth/v1/generate?grant_type=client_credentials",nil)
if err != nil{
log.Fatal(err.Error())
return
}
request.Header.Add("Authorization","Basic "+databytes)
var response,err2 = client.Do(request)
if err2 != nil{
log.Fatal(err2.Error())
return
}
defer response.Body.Close()
databytess,_ := ioutil.ReadAll(response.Body)
var responseMap map[string]string
err3 := json.Unmarshal(databytess,&responseMap)
if err3 != nil{
panic(err3.Error())
}
var access_string string = responseMap["access_token"]
Access_Token <- access_string

}



func Initiate_Mpesa_Express(res http.ResponseWriter,req *http.Request){
env.Load()
var paymentSecret = os.Getenv("paymentsecret")
// var userPaymentSecret = os.Getenv("userPaymentSecret")
var client = http.Client{}
var Request_Body Request_Express
var Response_Body mpesaexpress.Mpesa_Payment
env.Load()
json.NewDecoder(req.Body).Decode(&Request_Body)
if strings.HasPrefix(Request_Body.Customers_Phone_Number,"0") && len(Request_Body.Customers_Phone_Number) == 10{
Request_Body.Customers_Phone_Number = "254"+Request_Body.Customers_Phone_Number[1:]
}


var passKey = os.Getenv("passkey")
var Business_Shortcode string  = "174379"
go GenerateAcess_Token()
time.Sleep(time.Millisecond * 500)
access_grant_token := <- Access_Token
var timeStamp string = time.Now().Format("20060102150405")
mergedString := Business_Shortcode + passKey +timeStamp
var Password  string = base64.RawStdEncoding.EncodeToString([]byte(mergedString))

var PayloadData = Body_Express{    
	BusinessShortCode:"174379",    
	Password:Password,    
	Timestamp:timeStamp,    
	TransactionType:"CustomerPayBillOnline",    
	Amount: Request_Body.Amount,    
	PartyA:Request_Body.Customers_Phone_Number,    
	PartyB:"174379",    
	PhoneNumber:Request_Body.Customers_Phone_Number,    
	CallBackURL: "https://mydomain.com/path",    
	AccountReference:"Guzman Stores",    
	TransactionDesc:"Guzman Stores",
 }
var marshalledPayload,_= json.Marshal(PayloadData)
bytesPayload :=bytes.NewReader(marshalledPayload)
var request,err2 = http.NewRequest("POST","https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest",bytesPayload)
if err2 != nil{
log.Fatal(err2.Error())
return
}
request.Header.Add("Authorization","Bearer "+access_grant_token)
request.Header.Add("Content-Type","application/json")
Response,err5:= client.Do(request)
if err5 != nil{
panic(err5.Error())
}
defer Response.Body.Close()

bytesBody,_ := json.Marshal(Request_Body)


databytes,_ := ioutil.ReadAll(Response.Body)
errjson := json.Unmarshal(databytes,&Response_Body)
if errjson != nil {
log.Fatal(errjson.Error())
return
}
if Response_Body.CustomerMessage == "Success. Request accepted for processing"{
var token = jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
"merchant_id":string(databytes),
"exp":time.Now().Add(time.Hour * 4).Unix(),
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
var passkey = os.Getenv("passkey")
var BusinessShortCode = "174379"
var timeStamp = time.Now().Format("20060102150405")
mergedString := BusinessShortCode + passkey +timeStamp
var Password  string = base64.RawStdEncoding.EncodeToString([]byte(mergedString))


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
fmt.Println(actualToken[1])
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

var ResponseLoad mpesaexpress.Mpesa_Payment
errj := json.Unmarshal([]byte(payloadMap),&ResponseLoad)
if errj != nil{
panic(errj.Error())
}
errUser := json.Unmarshal([]byte(userPayloadmap),&userInfo)
if errUser != nil{
panic(errUser.Error())
}
go GenerateAcess_Token()
time.Sleep(time.Millisecond * 555)
access_token_string := <- Access_Token


var payLoadValidation = Payload_Request_Validation{
CheckoutRequestID:ResponseLoad.CheckoutRequestID,
Timestamp: timeStamp,
Password: Password,
BusinessShortCode: BusinessShortCode,
}

var client = http.Client{}
databytesPayload,_ := json.Marshal(payLoadValidation)
readerLoad := bytes.NewReader(databytesPayload)
Request,errRequest := http.NewRequest("POST","https://sandbox.safaricom.co.ke/mpesa/stkpushquery/v1/query",readerLoad)
if errRequest != nil {
panic(errRequest.Error())
}
Request.Header.Set("Authorization","Bearer "+access_token_string)
Request.Header.Set("Content-Type","application/json")
Response,_ := client.Do(Request)
defer Response.Body.Close()    

var Response_struct mpesaexpress.Confirmation_Payment_Mpesa
errDecoding := json.NewDecoder(Response.Body).Decode(&Response_struct)
msg := make(map[string]string, 0)

if errDecoding != nil{
msg["message"] ="Internal Server Error"	
bites,_ := json.Marshal(msg)
res.Write(bites)
panic(errDecoding.Error())
}
fmt.Println(Response_struct.ResultCode)
switch Response_struct.ResultCode {
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
userInfo.Checkout_Request_Id = ResponseLoad.CheckoutRequestID

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

}


   
      

}   