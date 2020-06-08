package common

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"pi-software/model"
	"pi-software/repos"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type AdminPipeline struct {
	AddMosque           bool
	DeleteMosque        bool
	ShowMosque          bool
	RegisterAdmin       bool
	RegisterMosqueAdmin bool
	Mosques             []model.Mosque
}

var mosque model.Mosque

type Date struct { // custom date type for easier templating and format of date
	Date   string
	Prayer []model.Prayer
}

var dates []Date

func AdminHandler(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		pathUrl, _ := url.ParseQuery(request.URL.RawQuery)
		path := pathUrl.Encode()
		if path != "" {
			action := strings.Split(path, "=")[1]
			addM := false
			delM := false
			showM := false
			regA := false
			regMA := false
			switch action {
			case "addmosque":
				addM = true
			case "deletemosque":
				delM = true
			case "showmosque":
				showM = true
			case "registeradmin":
				regA = true
			case "registermosqueadmin":
				regMA = true
			}
			adminPipe := AdminPipeline{addM, delM, showM, regA, regMA, getMosques(response, request)}
			t, _ := template.ParseFiles("templates/admin.html")
			t.Execute(response, adminPipe)
		} else {
			t, _ := template.ParseFiles("templates/adminIndex.html")
			t.Execute(response, nil)
		}
	} else {
		accessError(response, request)
	}
}

func accessError(response http.ResponseWriter, request *http.Request) {
	t, _ := template.ParseFiles("templates/errorpage.html")
	t.Execute(response, errors.New("Illegal Access"))
	response.Write([]byte(`<script>window.location.href = "/error"</script>`))
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
			var date model.Date
			var dates []model.Date
			var prayer model.Prayer
			var prayers []model.Prayer
			prayer.CapacityMen = cap_m
			prayer.CapacityWomen = cap_w
			prayer.Users = []model.User{}
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
				}
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
			mosque.Active = true

			collection.InsertOne(context.TODO(), mosque)

			http.Redirect(response, request, "/admin", 302) // redirect back to Adminpage
		} else {
			t, _ := template.ParseFiles("templates/addMosque.html")
			t.Execute(response, nil)
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
		if mosqueName != "" && !confirm {
			collection, _ := repos.GetDBCollection(1)
			collection.FindOne(context.TODO(),
				bson.D{
					{"Name", mosqueName},
				}).Decode(&mosque)
			t, _ := template.ParseFiles("templates/show-hide.html")
			t.Execute(response, mosque)
		} else if confirm {
			fmt.Println("confirm")
			var prayers []model.Prayer
			mosque.Active = !mosque.Active
			if !mosque.Active { // if old status active, then list active registrations
				for _, date := range mosque.Date {
					for _, prayer := range date.Prayer {
						if len(prayer.Users) > 0 {
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
			collection, _ := repos.GetDBCollection(1)
			collection.UpdateOne(context.TODO(), bson.M{"Name": mosque.Name}, bson.M{"$set": bson.M{"Active": mosque.Active}})
			response.Write([]byte(`<script>window.location.href = "/activeRegistrations";</script>`))
		} else {
			http.Redirect(response, request, "/admin", 300)

		}
	} else {
		accessError(response, request)
	}
}

func ActiveRegistrations(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		if len(dates) > 0 {
			t, _ := template.ParseFiles("templates/activeRegistrations.html")
			t.Execute(response, dates)
			dates = []Date{}
		} else {
			response.Write([]byte(`<script>window.location.href = "/admin";</script>`))

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
			fmt.Fprintln(response, "User already exists")
		}
	} else {
		accessError(response, request)
	}
}

func containString(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}
