package adminhelper

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jamesmukumu/guzman/work/models/users"
	env "github.com/joho/godotenv"
)

var response = make(map[string]string, 0)


func Prevalidate_Admin_Creation(next http.HandlerFunc)http.HandlerFunc{
return func(res http.ResponseWriter, req *http.Request) {
env.Load()
var secretPayment string = os.Getenv("jwtSecret")
var tokenExtracted = req.Header.Get("Authorization")
if tokenExtracted == ""{
response["message"] = "Unauthorized"
databytes,_ := json.Marshal(response)
res.Write(databytes)
return
}
var tokenSlice []string = strings.Split(tokenExtracted," ")   
var actualToken = tokenSlice[1]


if len(tokenSlice) != 2 ||  actualToken == ""{
response["message"] = "Unauthorized"
databytes,_ := json.Marshal(response)
res.Write(databytes)
return
}



var tok,err = jwt.Parse(actualToken,func(token *jwt.Token) (interface{}, error) {
return []byte(secretPayment),nil
})

if err!= nil {
log.Fatal(err.Error())
response["message"] = "Unauthorized"
return
}
if !tok.Valid{
response["message"] ="Token is invalid"
databytes,_ := json.Marshal(response)
res.Write(databytes)
return
}

var claims = tok.Claims.(jwt.MapClaims)
var payloadString = claims["user"].(string)
var User users.Users
err1 := json.Unmarshal([]byte(payloadString),&User)
if err1 != nil {
panic(err1.Error())
}
if !User.Can_Access_All_Functions{
response["message"] = "Admin Not Authorized to handle These Functions"
databytes,_ := json.Marshal(response)
res.Write(databytes)
return
}
next.ServeHTTP(res,req)
    



}
}
