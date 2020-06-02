package main

import (
	"net/http"
	"pi-software/common"

	"github.com/gorilla/mux"
)

var router = mux.NewRouter()

// Functions for handling pagecalls like localhost:8080/login
func main() {
	router.HandleFunc("/register", common.RegisterPageHandler).Methods("GET")
	router.HandleFunc("/register", common.RegisterHandler).Methods("POST")

	router.HandleFunc("/", common.IndexPageHandler)
	router.HandleFunc("/login", common.LoginHandler)

	router.HandleFunc("/index", common.IndexPageHandler)

	router.HandleFunc("/logout", common.LogoutHandler)
	//router.HandleFunc("/reset", common.ResetHandler)
	router.HandleFunc("/submit", common.SubmitPrayer)

	router.HandleFunc("/chooseMosque", common.Choose) // normally ChooseMosque
	router.HandleFunc("/choose", common.Choosen)
	router.HandleFunc("/chooseDate", common.ChooseDate)
	router.HandleFunc("/choosePrayer", common.ChoosePrayer)
	router.HandleFunc("/add", common.Add)
	router.HandleFunc("/addMosque", common.AddMosque)

	//router.HandleFunc("/appleHeadquarter", common.SteveJobsHandler)
	router.HandleFunc("/addMosque", common.AddMosque)
	router.HandleFunc("/signOut", common.SignOutPrayer)
	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}
