package router

import (
	"fmt"
	"log"
	"net/http"
	"os"

	mux "github.com/gorilla/mux"
	"github.com/jamesmukumu/guzman/work/controllers/mpesaexpresscont"
	"github.com/jamesmukumu/guzman/work/controllers/usercontrollers"
	"github.com/jamesmukumu/guzman/work/helpers/adminhelper"
	env "github.com/joho/godotenv"
	cors "github.com/rs/cors"
)


func ServerSetup(){
env.Load()
port := os.Getenv("port")


Router := mux.NewRouter()
var corsOptions cors.Options = cors.Options{
AllowedOrigins:[]string{"*"},
AllowedMethods: []string{"POST","GET","DELETE","PATCH","PUT"},
AllowedHeaders: []string{"*"},    
}
corsHandler := cors.New(corsOptions)
Router.HandleFunc("/create/new/user",adminhelper.Prevalidate_Admin_Creation(usercontrollers.RegisterUser)).Methods("POST")
Router.HandleFunc("/validate/pin",usercontrollers.Grant_Permission).Methods("POST")
Router.HandleFunc("/initiate/payment",mpesaexpresscont.Initiate_Mpesa_Express).Methods("POST")
Router.HandleFunc("/validate/payment",mpesaexpresscont.Validate_Payment).Methods("POST")
Router.HandleFunc("/fetch/all/admins",usercontrollers.FecthallAdmins).Methods("GET")
Router.HandleFunc("/admin/fetch/id",usercontrollers.Fetch_Admin_Primary_Key).Methods("GET")


var ActualHandler = corsHandler.Handler(Router)
fmt.Printf("Server Listening for Requests at port %s",port)
log.Fatal(http.ListenAndServe(":9900",ActualHandler))

   


}