package usercontrollers

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"log"
	"net/http"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jamesmukumu/guzman/work/db"
	"github.com/jamesmukumu/guzman/work/models/users"
	env "github.com/joho/godotenv"
	bcrypt "golang.org/x/crypto/bcrypt"
)
type Pin struct{
Pin string `json:"Pin"`
User_Name string `json:"user_name"` 
}

type New_Admin_Name struct{
Admin_Name string `json:"admin_name"`
}
type Password struct{
Pass string `json:"password"`
}

    
func RegisterUser(res http.ResponseWriter,req *http.Request) {
var User users.Users
err := json.NewDecoder(req.Body).Decode(&User)
if err != nil{
log.Fatal(err.Error())
return
}
HashedBytes,err1 := bcrypt.GenerateFromPassword([]byte(User.Pin),11)
if err1 != nil{
panic(err1.Error())

}
var hashedPasswordstring string = string(HashedBytes)
User.Pin = hashedPasswordstring  
User.PresetTodefault()
result := db.Connection.Create(&User)
if result.RowsAffected !=0{
var mapString = make(map[string]string, 0)
mapString["message"] ="User Saved Successfully"   
mapString["rowsAffected"] = "1"
var databytes,_ = json.Marshal(mapString)
res.Write(databytes)
}else if result.Error != nil{
panic(result.Error.Error())
}
}


func Grant_Permission(res http.ResponseWriter,req *http.Request){
env.Load()
var secretJwt = os.Getenv("jwtSecret")
var pin Pin
var user users.Users
err := json.NewDecoder(req.Body).Decode(&pin)
if pin.Pin == "" || pin.User_Name == ""{
res.Write([]byte("Fill All fields"))
return
}
if err != nil{
log.Fatal(err.Error())
return
}

result := db.Connection.Where("users_name = ?",pin.User_Name).Find(&user)
if result.RowsAffected == 0{
jsonResp := map[string]string{
"message":"User Does Not Exist",
}
response,_ := json.Marshal(jsonResp)
res.Write([]byte(response))
return
}
fmt.Print(user.Pin)
err1 := bcrypt.CompareHashAndPassword([]byte(user.Pin),[]byte(pin.Pin))
if err1 != nil{

message := map[string]string{    
"message":"Pin mismatch",
}
databytes,_ := json.Marshal(message)
res.Write([]byte(databytes))
return
}else{
var UserInfo,_ = json.Marshal(user)
var token = jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
"user":string(UserInfo),
"exp":time.Now().Add(time.Hour * 1).Unix(),
})
var actualToken,errToken = token.SignedString([]byte(secretJwt))
if errToken != nil{
log.Fatal(errToken.Error())
return
}   

message := make(map[string]string, 0)
message["message"] = "Pin Accepted"
message["Token"] = actualToken
databytes,_ := json.Marshal(message)
res.Write([]byte(databytes))
}
if result.Error != nil{
messageResp := make(map[string]string,0)
messageResp["message"] ="Internal Server Error"
databytes,_ := json.Marshal(messageResp)
res.Write([]byte(databytes))    
}
}







func FecthallAdmins(res http.ResponseWriter,req *http.Request){
var Admins [] users.Users
db.Connection.Find(&Admins)
json.NewEncoder(res).Encode(map[string]interface{}{
"message":"Admins Fetched",
"data":Admins,
})
}


func Fetch_Admin_Primary_Key(res http.ResponseWriter,req *http.Request){

var admin users.Users

id_number := req.URL.Query().Get("id_number")

result := db.Connection.Find(&admin,id_number)
json.NewEncoder(res).Encode(map[string]interface{}{
"message":"Admin Fetched",
"data":admin,
"rowsaffected":result.RowsAffected,
})

}



func Adjust_Admins_Name(res http.ResponseWriter,req *http.Request){
var message = make(map[string]string,0)
var Admin users.Users
var adminLoad  New_Admin_Name
var admin_number string = req.URL.Query().Get("admin_number")
err := json.NewDecoder(req.Body).Decode(&adminLoad)
if err != nil{
log.Fatal(err.Error())
return
}

match := db.Connection.Table("users").Where("id = ?",admin_number).Find(&Admin)
if match.RowsAffected == 1 && Admin.Users_Name == adminLoad.Admin_Name {
message["message"] ="Admin Name cannot be same as old name"
message["content"] = "Try using a different admin name"
databytes,_ := json.Marshal(message)
res.WriteHeader(202)
res.Write(databytes)     
return
}else if match.RowsAffected == 0 && match.Error != nil {
message["message"] = "This Admin Does not exist"
databytes, _ :=  json.Marshal(message)
res.WriteHeader(202)
res.Write(databytes)
return
}
//

result := db.Connection.Table("users").Where("id = ?",admin_number).Update("users_name",adminLoad.Admin_Name)
if result.RowsAffected == 1  {
message["message"] = "Admin Name Accepted Successfully"
databytes,_ := json.Marshal(message)
res.WriteHeader(200)     
res.Write(databytes)
return
}
     
}



func Delete_Admin(res http.ResponseWriter,req *http.Request){
var Admin users.Users
var msg  = make(map[string]string,0)
var admins_number string = req.URL.Query().Get("admin_number")
result := db.Connection.Table("users").Where("id = ?",admins_number).Delete(&Admin)

if result.RowsAffected == 1 && result.Error == nil{
msg["message"] ="Admin Has been Deleted successfully"
msg["rowsAffected"] ="1"
databytes,_ := json.Marshal(msg)
res.WriteHeader(202)
res.Write(databytes)
}else if result.RowsAffected == 0 && result.Error == nil{
msg["message"] = "Admin Does not exist"
msg["rowsAffected"] = "0"
databytes,_ := json.Marshal(msg)   
res.WriteHeader(202)
res.Write(databytes)
}   

}




func Generate_Reset_Token(res http.ResponseWriter,req *http.Request){
env.Load()
msg := make(map[string]string, 0)
var resetSecret string = os.Getenv("resetSecret")
var Admin users.Users
Admin_Name := req.URL.Query().Get("username")
result := db.Connection.Table("users").Where("users_name = ?",Admin_Name).Find(&Admin)
if result.RowsAffected == 1 && result.Error == nil {
databytes,_ := json.Marshal(Admin)
Token := jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
"admin":string(databytes),
"exp":time.Now().Add(time.Minute * 5).Unix(),
})

Token_string,_ := Token.SignedString([]byte(resetSecret))
msg["message"] ="Admin Found"
msg["reset_token"] = Token_string
data,_ := json.Marshal(msg)
res.Write(data)
return
}else if result.RowsAffected == 0 &&  result.Error == nil {
msg["message"] ="Admin Not Found"
databytes,_ := json.Marshal(msg)
res.Write(databytes)
return
  }else{
json.NewEncoder(res).Encode(map[string]string{
"message":"Internal server Error",
})
}
}



func Reset_Password(res http.ResponseWriter,req *http.Request){
env.Load()
var tokenString string = os.Getenv("resetSecret")
var msg = make(map[string]string,0)
var token = req.Header.Get("Authorization")
if token == "" || len(strings.Split(token,"")) < 1{
res.WriteHeader(401)
msg["message"] = "Unauthorized"
databytes,_ := json.Marshal(msg)
res.Write(databytes)
return
}
var actualToken = strings.Split(token," ")[1]
tok, err := jwt.Parse(actualToken,func(tok *jwt.Token) (interface{}, error) {
return []byte(tokenString),nil
})
if err!=nil && strings.Contains(err.Error(),"token has invalid claims: token is expired"){
msg["message"] = "Unauthorized"
databytes,_ := json.Marshal(msg)
res.Write(databytes)
return
}
var Admin users.Users
var tokenMap = tok.Claims.(jwt.MapClaims)
var admin = tokenMap["admin"].(string)
err1 := json.Unmarshal([]byte(admin),&Admin)
if err1 != nil {
log.Fatal(err1.Error())
return
}


var Password Password
err3 := json.NewDecoder(req.Body).Decode(&Password)
if err3 != nil {
log.Fatal(err3.Error())
return
}

db.Connection.Table("users").Where("users_name = ?",Admin.Users_Name).Find(&Admin)
errCompare := bcrypt.CompareHashAndPassword([]byte(Admin.Pin),[]byte(Password.Pass))
if errCompare ==  nil {
msg["message"] ="Old Password and New Password Match."
msg["content"] ="Try another passowrd,different from the old one"
databytes, _ := json.Marshal(msg)
res.Write(databytes)   
return
}       

encryptedPassword, _ := bcrypt.GenerateFromPassword([]byte(Password.Pass),11)
result := db.Connection.Table("users").Where("users_name = ?",Admin.Users_Name).Update("pin",string(encryptedPassword))
if result.RowsAffected == 1 && result.Error == nil {
msg["message"] = "Password Changes Successfully"
databytes, _ := json.Marshal(msg)
res.Write(databytes)
return
}else if !tok.Valid {
msg["message"] = "Unauthorized"
databytes,_ := json.Marshal(msg)
res.Write(databytes)
return
}


}

    

func Create_Favorites(res http.ResponseWriter,req *http.Request){
env.Load()
msg := make(map[string]string,0)
var secretAdmin = os.Getenv("jwtSecret")

var Admin users.Users
var fav users.Favourites_Customers
var token = req.Header.Get("Authorization")

if token ==""{
msg["message"] = "Unauthorized"
databytes,_ := json.Marshal(msg)
res.Write(databytes)
return
}
var actualToken = strings.Split(token," ")[1]  

tok,err := jwt.Parse(actualToken,func(t *jwt.Token) (interface{}, error) {
return []byte(secretAdmin),nil
})

if err != nil && strings.Contains(err.Error(),"token has invalid claims: token is expired") {
msg["message"] = "Unauthorized"
databytes,_ := json.Marshal(msg)
res.WriteHeader(401)  
res.Write(databytes)
return
}


var tokenMap = tok.Claims.(jwt.MapClaims)
var Admin_User = tokenMap["user"].(string)
json.Unmarshal([]byte(Admin_User),&Admin)


json.NewDecoder(req.Body).Decode(&fav)
fav.Created_By = Admin.Users_Name
result := db.Connection.Table("favourites_customers").Create(&fav)
if result.RowsAffected == 1 && result.Error == nil {
msg["message"] ="Customer Saved To Favourites"
databytes,_ := json.Marshal(msg)
res.WriteHeader(200)
res.Write(databytes)
}else if result.RowsAffected == 0{
msg["message"] = "This Customer already exists"
databytes, _ := json.Marshal(msg)
res.WriteHeader(202)
res.Write(databytes)   
}}







func Fetch_Favourites(res http.ResponseWriter,req *http.Request){
env.Load()
msg := make(map[string]string, 0)
var secretToken string = os.Getenv("jwtSecret")
var token  = req.Header.Get("Authorization")
if token == ""{
msg["message"] ="Unauthorized" 
databytes,_ := json.Marshal(msg)
res.WriteHeader(401)
res.Write(databytes)
return  
}

var actualTok = strings.Split(token," ")[1]
tok,err := jwt.Parse(actualTok,func(t *jwt.Token) (interface{}, error) {
return []byte(secretToken),nil
})

if err != nil && strings.Contains(err.Error(),"token has invalid claims: token is expired") {
res.WriteHeader(401)
json.NewEncoder(res).Encode(map[string]string{
"message":"Token Is expired",
})
return
}

var user string  = tok.Claims.(jwt.MapClaims)["user"].(string)
var Admin users.Users
json.Unmarshal([]byte(user),&Admin)

var Favs []users.Favourites_Customers

result := db.Connection.Table("favourites_customers").Where("created_by",Admin.Users_Name).Find(&Favs)
if result.RowsAffected != 0 && result.Error == nil{
json.NewEncoder(res).Encode(map[string]interface{}{   
"message":"Favourites Fetched",
"data":Favs,
})
}
}




func Delete_Favs(res http.ResponseWriter,req *http.Request){
env.Load()
var secret string = os.Getenv("jwtSecret")
var token string = req.Header.Get("Authorization")
var resp = make(map[string]interface{},0)
var customersID string = req.URL.Query().Get("customer")
var actualToken =  strings.Split(token," ")[1]
if actualToken == "" || len(strings.Split(token," ")) < 0{
resp["message"] ="Unaruthorized"
databytes,_ := json.Marshal(resp)
res.WriteHeader(401)  
res.Write(databytes)
return
}

tok,err := jwt.Parse(actualToken,func(t *jwt.Token) (interface{}, error) {
return []byte(secret),nil
})
if err != nil && strings.Contains(err.Error(),"token is invalid"){
resp["message"] ="Unaruthorized"
databytes,_ := json.Marshal(resp)
res.WriteHeader(401)
res.Write(databytes)
return
}
customerIIintform,_ := strconv.Atoi(customersID)
var payload = tok.Claims.(jwt.MapClaims)["user"].(string)
var Admin users.Users
var Fav users.Favourites_Customers
json.Unmarshal([]byte(payload),&Admin)
result := db.Connection.Table("favourites_customers").Where("id  = ?",customerIIintform).Where("created_by = ?",Admin.Users_Name).Delete(&Fav)
if result.RowsAffected != 0 && result.Error == nil {
resp["message"] = "Customer has been deleted"
databytes, _ := json.Marshal(resp)
res.WriteHeader(200)
res.Write(databytes)
return
}else if result.RowsAffected == 0 && result.Error == nil{
resp["message"] = "Error in deleting"
databytes, _ := json.Marshal(resp)
res.WriteHeader(200)
res.Write(databytes)
}
}