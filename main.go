package main
 
import (
    "net/http"
 
    "github.com/gorilla/mux"
 
    common "./common"
)
 
var router = mux.NewRouter()
 
func main() {
    router.HandleFunc("/", common.LoginPageHandler) // GET
    router.HandleFunc("/login", common.LoginHandler).Methods("POST")
    
    router.HandleFunc("/index", common.IndexPageHandler) // GET
    
    router.HandleFunc("/register", common.RegisterPageHandler).Methods("GET")
    router.HandleFunc("/register", common.RegisterHandler).Methods("POST")
    
    router.HandleFunc("/logout", common.LogoutHandler)//.Methods("POST")
	router.HandleFunc("/reset", common.ResetHandler)
	router.HandleFunc("/submit", common.SubmitHandler)
	
	router.HandleFunc("/appleHeadquarter", common.SteveJobsHandler)
	router.HandleFunc("/setFlag", common.SetFlag)
	
	
    http.Handle("/", router)
    

    http.ListenAndServe(":8080", nil)
}
