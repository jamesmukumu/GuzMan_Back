package usercontrollers

import (
	"encoding/json"
	"os"
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
"exp":time.Now().Add(time.Hour * 1).UnixNano(),
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