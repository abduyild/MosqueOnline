package common

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"pi-software/model"
	"pi-software/repos"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type AdminPipeline struct {
	AddMosque    bool
	DeleteMosque bool
	ShowMosque   bool
	Mosques      []model.Mosque
}

var mosque model.Mosque

func AdminHandler(response http.ResponseWriter, request *http.Request) {
	adminPipe := AdminPipeline{false, false, false, getMosques(response, request)}
	t, _ := template.ParseFiles("templates/admin.html")
	t.Execute(response, adminPipe)
}

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
		t := check(response, request, err)
		if t != nil {
			t.Execute(response, errors.New(dbConnectionError))
			return
		}
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
		mosque.MaxCapM = cap_m
		mosque.MaxCapW = cap_w

		collection.InsertOne(context.TODO(), mosque)

		http.Redirect(response, request, "/admin", 302) // redirect back to Adminpage
	} else {
		t, _ := template.ParseFiles("templates/addMosque.html")
		t.Execute(response, nil)
	}
}

func DeleteMosque(response http.ResponseWriter, request *http.Request) {
	mosque := request.URL.Query().Get("mosque")
	collection, err := repos.GetDBCollection(1)
	t := check(response, request, err)
	if t != nil {
		t.Execute(response, errors.New(dbConnectionError))
		return
	}
	collection.DeleteOne(context.TODO(), bson.M{"Name": mosque})

	collection, err = repos.GetDBCollection(0)
	t = check(response, request, err)
	if t != nil {
		t.Execute(response, errors.New(dbConnectionError))
		return
	}
	update := bson.D{{Key: "$pull", Value: bson.D{{Key: "RegisteredPrayers", Value: bson.D{{Key: "MosqueName", Value: mosque}}}}}}
	collection.UpdateMany(context.TODO(), bson.D{{}}, update)

}
