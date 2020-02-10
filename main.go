package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"pi-software/common"
)

var router = mux.NewRouter()

// Functions for handling pagecalls like localhost:8080/login
func main() {
	router.HandleFunc("/", common.LoginPageHandler) // GET
	router.HandleFunc("/login", common.LoginHandler).Methods("POST")

	router.HandleFunc("/index", common.IndexPageHandler) // GET

	router.HandleFunc("/register", common.RegisterPageHandler).Methods("GET")
	router.HandleFunc("/register", common.RegisterHandler).Methods("POST")

	router.HandleFunc("/logout", common.LogoutHandler) //.Methods("POST")
	router.HandleFunc("/reset", common.ResetHandler)
	router.HandleFunc("/submit", common.SubmitHandler)

	router.HandleFunc("/appleHeadquarter", common.SteveJobsHandler)
	router.HandleFunc("/setFlag", common.SetFlag)
	router.HandleFunc("/setAllFlags", common.SetAllFlags)
	router.HandleFunc("/setAllFlagsForOne", common.SetAllFlagsForOne)
	router.HandleFunc("/addVm", common.AddVm)
	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}
