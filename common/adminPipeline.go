package common

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"pi-software/model"
	"pi-software/repos"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type AdminPipeline struct {
	DeleteMosque bool
	ShowMosque   bool
	Mosques      []model.Mosque
}

var mosque model.Mosque

var fridayIndices []int

type Date struct { // custom date type for easier templating and format of date
	Date   string
	Prayer []model.Prayer
}

var dates []Date

func AdminHandler(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		action := request.URL.Query().Get("action")
		target := "templates/admin.gohtml"
		delM := false
		showM := false
		if action != "" {
			switch action {
			case "deletemosque":
				delM = true
				target = "templates/adminAction.gohtml"
			case "showmosque":
				showM = true
				target = "templates/adminAction.gohtml"
			}
		}
		adminPipe := AdminPipeline{delM, showM, getMosques(response, request)}
		t, _ := template.ParseFiles(target, "templates/base_adminloggedin.tmpl", "templates/footer.tmpl")
		t.Execute(response, adminPipe)
	} else {
		accessError(response, request)
	}
}

func AddMosque(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		request.ParseForm()

		if len(request.PostForm) > 0 {
			name := R(request.FormValue("name"))
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
			addDates := 100 // set how many dates you want to add to the future
			var prayer model.Prayer
			var prayers []model.Prayer
			prayer.CapacityMen = cap_m
			prayer.CapacityWomen = cap_w
			prayer.Users = []model.User{}
			cumaSet := false
			bayramSet := false
			form := request.Form["prayer"]
			for i := 1; i < 8; i++ {
				switch i {
				case 1:
					prayer.Available = containString(form, "fajr")
				case 2:
					prayer.Available = containString(form, "dhuhr")
				case 3:
					prayer.Available = containString(form, "asr")
				case 4:
					prayer.Available = containString(form, "maghrib")
				case 5:
					prayer.Available = containString(form, "ishaa")
				case 6:
					cumaSet = containString(form, "cuma")
				case 7:
					bayramSet = containString(form, "bayram")
				}
				prayer.Name = model.PrayerName(i)
				prayers = append(prayers, prayer)
				prayer.Available = false
			}
			var indicesBayram []int
			datesToAdd := make([]model.Date, addDates)
			fridayIndices = []int{}
			for i := 0; i < addDates; i++ {
				var date model.Date
				currentDate := time.Now().AddDate(0, 0, i).Format(time.RFC3339)
				weekday := time.Now().AddDate(0, 0, i).Weekday()
				if cumaSet && int(weekday) == 5 { // cuma
					fridayIndices = append(fridayIndices, i)
				}
				if bayramSet && containString(model.GetBayrams(), strings.Split(currentDate, "T")[0]) {
					indicesBayram = append(indicesBayram, i)
				}
				date.Date, _ = time.Parse(time.RFC3339, currentDate)
				date.Prayer = prayers
				datesToAdd[i] = date
			}
			mosque := *new(Mosque)
			mosque.Name = name
			mosque.PLZ = plz
			mosque.Street = street
			mosque.City = city
			mosque.Date = datesToAdd
			mosque.MaxCapM = cap_m
			mosque.MaxCapW = cap_w
			mosque.Active = true
			collection.InsertOne(context.TODO(), mosque)
			for _, i := range fridayIndices {
				collection.UpdateOne(context.TODO(),
					bson.M{"Name": name},
					bson.M{"$set": bson.M{
						"Date." + strconv.Itoa(i) + ".Prayer.1.Available": false,
						"Date." + strconv.Itoa(i) + ".Prayer.5.Available": true}})
			}
			for _, i := range indicesBayram {
				collection.UpdateOne(context.TODO(),
					bson.M{"Name": name},
					bson.M{"$set": bson.M{"Date." + strconv.Itoa(i) + ".Prayer.6.Available": true}})
			}
			http.Redirect(response, request, "/admin", 302) // redirect back to Adminpage
		} else {
			http.Redirect(response, request, "/admin", 302) // redirect back to Adminpage
		}
	} else {
		accessError(response, request)
	}
}

func DeleteMosque(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
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
		response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
	} else {
		accessError(response, request)
	}
}

func ShowMosque(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		mosqueName := request.URL.Query().Get("mosque")
		confirm := false
		if request.URL.Query().Get("confirm") == "yes" {
			confirm = true
		}
		request.ParseForm()
		if mosqueName != "" {
			var mosque model.Mosque
			collection, _ := repos.GetDBCollection(1)
			collection.FindOne(context.TODO(),
				bson.D{
					{"Name", mosqueName},
				}).Decode(&mosque)
			var dates []Date
			type tMos struct {
				Name   string
				Date   []Date
				Active bool
				PLZ    int
				Street string
				City   string
			}
			var newMosque tMos
			newMosque.Name = mosque.Name
			newMosque.Active = mosque.Active
			newMosque.PLZ = mosque.PLZ
			newMosque.Street = mosque.Street
			newMosque.City = mosque.City
			if mosque.Active {
				reachedToday := false
				today := strings.Split(time.Now().String(), " ")[0]
				var prayers []model.Prayer
				for _, date := range mosque.Date {
					for _, prayer := range date.Prayer {
						if today == strings.Split(date.Date.String(), " ")[0] {
							reachedToday = true
						}
						if reachedToday && len(prayer.Users) > 0 {
							prayers = append(prayers, prayer)
						}
					}
					if len(prayers) > 0 {
						var dat Date
						dateS := strconv.Itoa(date.Date.Day()) + ". " + date.Date.Month().String() + " " + strconv.Itoa(date.Date.Year())
						dat.Date = dateS
						dat.Prayer = prayers
						prayers = []model.Prayer{}
						dates = append(dates, dat)
					}
				}
			}
			newMosque.Date = dates
			t, _ := template.ParseFiles("templates/show-hide.gohtml", "templates/base_adminloggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, newMosque)
		} else if confirm {
			mosque.Active = !mosque.Active
			collection, _ := repos.GetDBCollection(1)
			collection.UpdateOne(context.TODO(), bson.M{"Name": mosque.Name}, bson.M{"$set": bson.M{"Active": mosque.Active}})
			response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
		} else {
			http.Redirect(response, request, "/admin", 300)
		}
	} else {
		accessError(response, request)
	}
}

func RegisterAdmin(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		collection, err := repos.GetDBCollection(2)
		t := check(response, request, err)
		if t != nil {
			t.Execute(response, errors.New(dbConnectionError))
			return
		}
		request.ParseForm()
		// Get data the User typen into the fields
		name := ""
		email := request.FormValue("email")
		password := request.FormValue("password")
		admin := false
		if request.URL.Path == "/registerAdmin" {
			admin = true
			name = request.FormValue("name")
		} else {
			name = request.FormValue("register-mosqueadmin")
		}
		// Look if the entered Username is already used
		result := collection.FindOne(context.TODO(), bson.D{{"Email", email}})
		// If not found (throws exception/error) then we can proceed
		if result.Err() != nil {

			// Generate the hashed password with 14 as salt
			hash, _ := bcrypt.GenerateFromPassword([]byte(password), 14)

			newAdmin := model.Admin{name, email, string(hash), admin}
			// Insert user to the table
			collection.InsertOne(context.TODO(), newAdmin)
			// Change redirect target to LoginPage
			http.Redirect(response, request, "/admin", 302)
		} else {
			// TODO: checkError
			fmt.Fprintln(response, "User already exists")
		}
	} else {
		accessError(response, request)
	}
}

func accessError(response http.ResponseWriter, request *http.Request) {
	t, _ := template.ParseFiles("templates/errorpage.gohtml", "templates/base_adminloggedin.tmpl", "templates/footer.tmpl")
	t.Execute(response, errors.New("Illegal Access"))
}

func containString(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}
