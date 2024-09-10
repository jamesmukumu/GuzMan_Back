package db

import (
	"fmt"
	"log"
	"os"

	mpesaexpress "github.com/jamesmukumu/guzman/work/models/mpesa_express"
	users "github.com/jamesmukumu/guzman/work/models/users"
	env "github.com/joho/godotenv"
	pg "gorm.io/driver/postgres"
	gorm "gorm.io/gorm"
)


var Connection *gorm.DB
func Db_connection(){
var Users users.Users
var Customer_Collections mpesaexpress.Customer_Details

var Mpesa_references mpesaexpress.Confirmation_Payment_Mpesa
env.Load()
var connection_Uri string = os.Getenv("connectionstringdb")
Connection_Db,err := gorm.Open(pg.Open(connection_Uri),&gorm.Config{})
if err != nil{
log.Fatal(err.Error())
return
}
Connection = Connection_Db

Connection.AutoMigrate(&Users,&Customer_Collections,&Mpesa_references)


fmt.Println("Connected to DB Successfully")
}