package main

import (
	"net/http"
	"pi-software/common"

	"github.com/gorilla/mux"
)

var router = mux.NewRouter()

// Functions for handling pagecalls like localhost:8080/login
func main() {
	router.HandleFunc("/", common.LoginPageHandler)
	router.HandleFunc("/login", common.LoginHandler).Methods("POST")

	router.HandleFunc("/index", common.IndexPageHandler)

	router.HandleFunc("/register", common.RegisterPageHandler).Methods("GET")
	router.HandleFunc("/register", common.RegisterHandler).Methods("POST")

	router.HandleFunc("/logout", common.LogoutHandler)
	//router.HandleFunc("/reset", common.ResetHandler)
	router.HandleFunc("/submit", common.SubmitHandler)

	router.HandleFunc("/chooseMosque", common.Choose) // normally ChooseMosque
	router.HandleFunc("/choose", common.Choosen)
	router.HandleFunc("/chooseDate", common.ChooseDate)
	router.HandleFunc("/choosePrayer", common.ChoosePrayer)

	//router.HandleFunc("/appleHeadquarter", common.SteveJobsHandler)
	router.HandleFunc("/setFlag", common.SetFlag)
	router.HandleFunc("/setAllFlags", common.SetAllFlags)
	router.HandleFunc("/setAllFlagsForOne", common.SetAllFlagsForOne)
	router.HandleFunc("/addMosque", common.AddMosque)
	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}
