package main

import (
	"fmt"
	"net/http"
	"pi-software/common"
	"pi-software/repos"

	"github.com/gorilla/mux"
)

var router = mux.NewRouter()

// Functions for handling pagecalls like localhost:8080/login
func main() {
	repos.StartCronjob()
	router.HandleFunc("/register", common.RegisterPageHandler).Methods("GET")
	router.HandleFunc("/register", common.RegisterHandler).Methods("POST")

	router.HandleFunc("/login", common.LoginHandler)
	router.HandleFunc("/index", common.IndexPageHandler)
	router.HandleFunc("/", common.IndexPageHandler)
	router.HandleFunc("/chooseMosque", common.Choose)
	router.HandleFunc("/choose", common.Choosen)
	router.HandleFunc("/chooseDate", common.ChooseDate)
	router.HandleFunc("/choosePrayer", common.ChoosePrayer)
	router.HandleFunc("/submit", common.SubmitPrayer)
	router.HandleFunc("/signOut", common.SignOutPrayer)

	router.HandleFunc("/admin", common.AdminHandler)
	router.HandleFunc("/deleteMosque", common.DeleteMosque)
	router.HandleFunc("/addMosque", common.AddMosque)
	router.HandleFunc("/show-hide", common.ShowMosque)
	router.HandleFunc("/registerAdmin", common.RegisterAdmin)
	router.HandleFunc("/registerMosqueAdmin", common.RegisterAdmin)
	router.HandleFunc("/addBayram", common.AddBayram)
	router.HandleFunc("/changeFutureDate", common.ChangeDate)
	router.HandleFunc("/editPrayers", common.EditPrayers)
	router.HandleFunc("/editCapacity", common.EditCapacity)
	router.HandleFunc("/show-mosques", common.ShowAllMosques)
	router.HandleFunc("/show-admins", common.ShowAdmins)
	router.HandleFunc("/changeAdmin", common.ChangeAdmin)
	router.HandleFunc("/addBanner", common.AddBanner)

	router.HandleFunc("/mosqueIndex", common.MosqueHandler)
	router.HandleFunc("/getRegistrations", common.GetRegistrations)

	router.HandleFunc("/confirmVisitors", common.ConfirmVisitors)

	router.HandleFunc("/logout", common.LogoutHandler)
	http.Handle("/banner/", http.StripPrefix("/banner", http.FileServer(http.Dir("./banner"))))
	http.Handle("/", router)
	fmt.Println("Server is up and running at Port :8080")
	http.ListenAndServe(":8080", nil)
}
