package mpesaexpresscont

import (
	"bytes"
	"crypto/rand"
	"strconv"

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

type Range_Date struct {
Start_Date string `json:"start"`
End_Date string `json:"end"`
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

}}   








func Fetch_todays_payments(res http.ResponseWriter,req *http.Request){
var Pays []mpesaexpress.Customer_Details
var msg = make(map[string]interface{}, 0)
now := time.Now()
startDay := time.Date(now.Year(),now.Month(),now.Day(),0,0,0,0,now.Location())
endDate := time.Date(now.Year(),now.Month(),now.Day(),23,59,59, 999999999,now.Location())

result := db.Connection.Table("customer_details").Where("created_at BETWEEN ? AND ?",startDay,endDate).Preload("Confirmation_Payment_Mpesa").Find(&Pays)
if result.RowsAffected > 0 && result.Error == nil {
var totalSalestoday int
for _,pay := range Pays{
tt,_ := strconv.Atoi(pay.Amount)
totalSalestoday += tt

}
msg["message"] = "Todays Payments fetched"  
msg["data"] = Pays  
msg["total_Sales_today"] =  totalSalestoday       
databytes,_ := json.Marshal(msg)
res.WriteHeader(202)
res.Write(databytes)
return
}else if result.RowsAffected == 0 && result.Error == nil {
msg["message"] = "You have not collected any payments today"
databytes,_ := json.Marshal(msg)   
res.Write(databytes)
}
}



func Filter_Time_Range_Payments(res http.ResponseWriter,req *http.Request){
var Day_Range Range_Date
var pays []mpesaexpress.Customer_Details
msg := make(map[string]interface{},0)
err := json.NewDecoder(req.Body).Decode(&Day_Range)
if err != nil {
panic(err.Error())
}
Start_Date,errStart := time.Parse(time.RFC3339,Day_Range.Start_Date)
if errStart != nil {
panic(errStart.Error())
}
EndDate,errEnd := time.Parse(time.RFC3339,Day_Range.End_Date)
if errEnd != nil {   
panic(errEnd.Error())
}
start := time.Date(Start_Date.Year(),Start_Date.Month(),Start_Date.Day(),23,59,59, 999999999,Start_Date.Location())
end :=  time.Date(EndDate.Year(),EndDate.Month(),EndDate.Day(),23,59,59, 999999999,EndDate.Location())
result := db.Connection.Table("customer_details").Where("created_at  BETWEEN ? AND ?",start,end).Preload("Confirmation_Payment_Mpesa").Find(&pays)
if result.RowsAffected > 0 && result.Error == nil {
var totals int

for _,Pays := range pays{
amounts,_ := strconv.Atoi(Pays.Amount)
totals += amounts
}

msg["message"] = "Results fetched for these filter"
msg["data"] = pays
msg["amount"] = totals
databytes,_ := json.Marshal(msg)
res.WriteHeader(202)
res.Write(databytes)
return
}else if result.RowsAffected == 0 && result.Error == nil {
msg["message"] ="No Results matching this filter"
databytes,_ := json.Marshal(msg)
res.WriteHeader(202)
res.Write(databytes)
}
}


func Fetch_Payments_Analysis(res http.ResponseWriter,req *http.Request){
var Result []map[string]interface{}
  
var pays []mpesaexpress.Customer_Details
var months   = []string{"January","February","March","April","May","June","July","August","September","October","November","December"}
now := time.Now()  

for i := 0; i < len(months); i++ {
    
    referenceDate := time.Date(now.Year(), time.Month(i+1), 1, 0, 0, 0, 0, now.Location())

    nextDate := referenceDate.AddDate(0, 1, 0)
    endDate := nextDate.AddDate(0, 0, -1).Add(23 * time.Hour).Add(59 * time.Minute).Add(59 * time.Second).Add(999999999 * time.Nanosecond)
    
    startDate := referenceDate
    
  
    resul := db.Connection.Table("customer_details").Where("created_at BETWEEN ? AND ?", startDate, endDate).Preload("Confirmation_Payment_Mpesa").Find(&pays)
    var totals int

    if resul.Error == nil {
     for _,p := range pays {
	tt,_ := strconv.Atoi(p.Amount)
	totals +=  tt
	 }
	 days := int(nextDate.Sub(referenceDate).Hours()/24)
        var mapResp = make(map[string]interface{}, 0)
        mapResp["month"] = months[i]
        mapResp["amount"] = totals
		mapResp["averagePaymentMonthly"] =  totals/days   
        Result = append(Result, mapResp)
    } else if resul.Error != nil {
        panic(resul.Error.Error())
        break
    }
}

databytes,_ := json.Marshal(Result)
res.WriteHeader(202)
res.Write(databytes)
}




func Fetch_Weekly_Analysis(res http.ResponseWriter,req *http.Request){
var pays []mpesaexpress.Customer_Details
var msg = make(map[string]interface{},0)
now := time.Now()
startDate := time.Date(now.Year(),time.Month(now.Month()),now.Day()-7,23,59,59,9999999,now.Location())
endDate := time.Date(now.Year(),time.Month(now.Month()),now.Day(),23,59,59,9999999,now.Location())
result := db.Connection.Table("customer_details").Where("created_at BETWEEN ? AND ?",startDate,endDate).Preload("Confirmation_Payment_Mpesa").Find(&pays)
if result.RowsAffected != 0 && result.Error == nil {
var totals int
for _,pay := range pays {
num,_ := strconv.Atoi(pay.Amount)
totals += num
}
msg["message"] ="Weekly payments fetched successfully"
msg["data"] = pays
msg["totals"] = totals
databytes,_ := json.Marshal(msg)
res.WriteHeader(200)
res.Write(databytes)
return
}else if result.RowsAffected == 0 && result.Error == nil{
msg["message"] = "You have not collected any payments"
databytes,_ := json.Marshal(msg)
res.WriteHeader(200)
res.Write(databytes)
}
}



func Fetch_montly_analysis(res http.ResponseWriter,req *http.Request){
var pays []mpesaexpress.Customer_Details
var msg = make(map[string]interface{},0)
now := time.Now()
initialRference := time.Date(now.Year(),now.Month(),1,1,1,1,1,now.Location())
nextMonth := now.AddDate(0,1,0)
diff := nextMonth.Sub(initialRference).Hours()/24

startDate := time.Date(now.Year(),now.Month(),1,1,1,1,9999999,now.Location())
endDate  := time.Date(now.Year(),now.Month(),int(diff),23,59,59,9999999,now.Location())

result := db.Connection.Table("customer_details").Where("created_at BETWEEN ? AND ?",startDate,endDate).Preload("Confirmation_Payment_Mpesa").Find(&pays)
if result.RowsAffected != 0 && result.Error == nil {
var totals int
for i:=0;i < len(pays);i++{
amt,_ := strconv.Atoi(pays[i].Amount)
totals += amt

}


msg["message"] = "Monthly Analysis payments fetched"
msg["data"] = pays
msg["totals"] = totals
databytes,_ := json.Marshal(msg)
res.WriteHeader(202)
res.Write(databytes)
return
}else if result.RowsAffected == 0 && result.Error == nil{
msg["message"] ="You have no monthly payments"
databytes,_ := json.Marshal(msg)
res.WriteHeader(202)
res.Write(databytes)
}


}