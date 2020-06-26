package main

import (
	"log"
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

	router.HandleFunc("/", common.LoginHandler)
	router.HandleFunc("/index", common.IndexPageHandler)
	router.HandleFunc("/chooseMosque", common.Choose)
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
	router.HandleFunc("/edit", common.Edit)
	router.HandleFunc("/show-mosques", common.ShowAllMosques)
	router.HandleFunc("/show-admins", common.ShowAdmins)
	router.HandleFunc("/changeAdmin", common.ChangeAdmin)
	router.HandleFunc("/addBanner", common.AddBanner)
	router.HandleFunc("/removeBanner", common.RemoveBanner)

	router.HandleFunc("/mosqueIndex", common.MosqueHandler)
	router.HandleFunc("/getRegistrations", common.GetRegistrations)
	router.HandleFunc("/confirmVisitors", common.ConfirmVisitors)

	router.HandleFunc("/logout", common.LogoutHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/banner/", http.StripPrefix("/banner", http.FileServer(http.Dir("./banner"))))
	http.Handle("/", router)
	log.Fatal(http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/camii.online/fullchain.pem", "/etc/letsencrypt/live/camii.online/privkey.pem", nil))
	go http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
	}))
	log.Println("Started")
}
