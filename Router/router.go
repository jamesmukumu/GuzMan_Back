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
Router.HandleFunc("/create/new/user",usercontrollers.RegisterUser).Methods("POST")
Router.HandleFunc("/validate/pin",usercontrollers.Grant_Permission).Methods("POST")
Router.HandleFunc("/initiate/payment",mpesaexpresscont.Initiate_Mpesa_Ums).Methods("POST")
 Router.HandleFunc("/validate/payment",mpesaexpresscont.Validate_Payment).Methods("POST")
Router.HandleFunc("/fetch/all/admins",adminhelper.Prevalidate_Admin_Creation(usercontrollers.FecthallAdmins)).Methods("GET")
Router.HandleFunc("/admin/fetch/id",adminhelper.Prevalidate_Admin_Creation(usercontrollers.Fetch_Admin_Primary_Key)).Methods("GET")
Router.HandleFunc("/update/admins/name",adminhelper.Prevalidate_Admin_Creation(usercontrollers.Adjust_Admins_Name)).Methods("PUT")
Router.HandleFunc("/delete/admin",adminhelper.Prevalidate_Admin_Creation(usercontrollers.Delete_Admin)).Methods("DELETE")
Router.HandleFunc("/create/new/fav/client",usercontrollers.Create_Favorites).Methods("POST")
Router.HandleFunc("/initiate/reset/password",usercontrollers.Generate_Reset_Token).Methods("POST")
Router.HandleFunc("/update/new/password",usercontrollers.Reset_Password).Methods("PUT")
Router.HandleFunc("/fetch/favourites",usercontrollers.Fetch_Favourites).Methods("GET")
Router.HandleFunc("/delete/favourite/customer",usercontrollers.Delete_Favs).Methods("DELETE")
Router.HandleFunc("/fetch/todays/payments",mpesaexpresscont.Fetch_todays_payments).Methods("GET")
Router.HandleFunc("/fetch/payments/on/filter",mpesaexpresscont.Filter_Time_Range_Payments).Methods("POST")

     

var ActualHandler = corsHandler.Handler(Router)   
fmt.Printf("Server Listening for Requests at port %s",port)
log.Fatal(http.ListenAndServe(":9900",ActualHandler))
      
   


}