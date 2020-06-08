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

	router.HandleFunc("/admin", common.AdminHandler)
	router.HandleFunc("/deleteMosque", common.DeleteMosque)
	router.HandleFunc("/addMosque", common.AddMosque)
	router.HandleFunc("/show-hide", common.ShowMosque)
	router.HandleFunc("/activeRegistrations", common.ActiveRegistrations)
	router.HandleFunc("/registerAdmin", common.RegisterAdmin)
	router.HandleFunc("/registerMosqueAdmin", common.RegisterAdmin)

	router.HandleFunc("/mosqueIndex", common.MosqueHandler)
	router.HandleFunc("/mosqueAction", common.MosqueAction)
	router.HandleFunc("/getRegistrationsDate", common.GetRegistrations)
	router.HandleFunc("/getRegistrations", common.GetRegistrations)
	router.HandleFunc("/getPrayers", common.GetPrayers)
	router.HandleFunc("/editPrayers", common.EditPrayers)
	router.HandleFunc("/confirmVisitors", common.ConfirmVisitors)

	router.HandleFunc("/index", common.IndexPageHandler)

	router.HandleFunc("/logout", common.LogoutHandler)
	router.HandleFunc("/submit", common.SubmitPrayer)

	router.HandleFunc("/chooseMosque", common.Choose) // normally ChooseMosque
	router.HandleFunc("/choose", common.Choosen)
	router.HandleFunc("/chooseDate", common.ChooseDate)
	router.HandleFunc("/choosePrayer", common.ChoosePrayer)
	router.HandleFunc("/addMosque", common.AddMosque)
	router.HandleFunc("/signOut", common.SignOutPrayer)
	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}
