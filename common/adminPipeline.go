package common

import (
	"context"
	"html/template"
	"net/http"
	"pi-software/model"
	"pi-software/repos"
	"strconv"
	"time"
)

type AdminPipeline struct {
	AddMosque    bool
	EditMosque   bool
	DeleteMosque bool
	ShowMosque   bool
	Mosques      []model.Mosque
}

func AdminHandler(response http.ResponseWriter, request *http.Request) {
	adminPipe := AdminPipeline{false, false, true, false, getMosques(response, request)}
	//rpId := mosque.Name + ":" + strconv.Itoa(index) + ":" + strconv.Itoa(prayer)
	t, _ := template.ParseFiles("templates/admin.html")
	t.Execute(response, adminPipe)
}

func Add(response http.ResponseWriter, request *http.Request) {
	t, err := template.ParseFiles("templates/addMosque.html")
	check(response, request, err)
	t.Execute(response, nil)
}

// Function for adding a VM to the Table
func AddMosque(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	if len(request.PostForm) > 0 {
		name := request.FormValue("name")
		plz, _ := strconv.Atoi(request.FormValue("plz"))
		street := request.FormValue("street")
		city := request.FormValue("city")
		cap_m, _ := strconv.Atoi(request.FormValue("cap-m"))
		cap_w, _ := strconv.Atoi(request.FormValue("cap-w"))

		collection, err := repos.GetDBCollection(1)
		check(response, request, err)
		var date model.Date
		var dates []model.Date
		var prayer model.Prayer
		var prayers []model.Prayer
		prayer.CapacityMen = cap_m
		prayer.CapacityWomen = cap_w
		prayer.Users = []model.User{}
		for i := 1; i < 6; i++ {
			prayer.Name = model.PrayerName(i)
			prayers = append(prayers, prayer)
		}
		// TODO: statt 10 einfach 100 nemhen oder so
		for i := 0; i < 10; i++ {
			currentDate := time.Now().AddDate(0, 0, i).Format(time.RFC3339)
			date.Date, _ = time.Parse(time.RFC3339, currentDate)
			date.Prayer = prayers
			dates = append(dates, date)
		}
		mosque := *new(Mosque)
		mosque.Name = name
		mosque.PLZ = plz
		mosque.Street = street
		mosque.City = city
		mosque.Date = dates

		collection.InsertOne(context.TODO(), mosque)

		http.Redirect(response, request, "/add", 302) // redirect back to Adminpage
	} else {

		t, err := template.ParseFiles("templates/addMosque.html")
		check(response, request, err)
		t.Execute(response, nil)
	}
}
