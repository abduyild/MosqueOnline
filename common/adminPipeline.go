package common

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
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

type Date struct { // custom date type for easier templating and format of date
	Date   string
	Prayer []model.Prayer
}

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
		adminPipe := AdminPipeline{delM, showM, getMosques(response, request, true)}
		t, _ := template.ParseFiles(target, "templates/base_adminloggedin.tmpl", "templates/footer.tmpl")
		t.Execute(response, adminPipe)
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func AddMosque(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		request.ParseForm()
		if len(request.PostForm) > 0 {
			name := R(request.FormValue("name"))
			mosque := getMosque(name)
			if mosque.Name != "" {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError("Camii Mevcut | Moschee bereits vorhanden", "/admin"))
			} else {
				plz, _ := strconv.Atoi(request.FormValue("plz"))
				street := request.FormValue("street")
				city := request.FormValue("city")
				maxdate, _ := strconv.Atoi(request.FormValue("maxdate"))
				cap_m, err := strconv.Atoi(request.FormValue("cap-m"))
				if err != nil {
					t, _ := template.ParseFiles("templates/errorpage.gohtml")
					t.Execute(response, GetError("Yanlis sayi boyutu | Falsches Zahlenformat", "/admin"))
				}
				cap_w, err := strconv.Atoi(request.FormValue("cap-w"))
				if err != nil {
					t, _ := template.ParseFiles("templates/errorpage.gohtml")
					t.Execute(response, GetError("Yanlis sayi boyutu | Falsches Zahlenformat", "/admin"))
				}

				collection, err := repos.GetDBCollection(1)
				if err != nil {
					t, _ := template.ParseFiles("templates/errorpage.gohtml")
					t.Execute(response, GetError(dbConnectionError, "/admin"))
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
				mosque := *new(Mosque)
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
				fridayIndices := []int{}
				for i := 0; i < addDates; i++ {
					var date model.Date
					currentDate := time.Now().AddDate(0, 0, i).Format(time.RFC3339)
					weekday := time.Now().AddDate(0, 0, i).Weekday()
					if cumaSet && int(weekday) == 5 { // cuma
						fridayIndices = append(fridayIndices, i)
					}
					eids := repos.GetEids()
					if bayramSet && containString(eids, strings.Split(currentDate, "T")[0]) {
						indicesBayram = append(indicesBayram, i)
					}
					date.Date, _ = time.Parse(time.RFC3339, currentDate)
					date.Prayer = prayers
					datesToAdd[i] = date
				}
				mosque.Name = name
				mosque.PLZ = plz
				mosque.Street = street
				mosque.City = city
				mosque.Date = datesToAdd
				mosque.MaxCapM = cap_m
				mosque.MaxCapW = cap_w
				mosque.Active = true
				mosque.Cuma = cumaSet
				mosque.Bayram = bayramSet
				if maxdate == 0 {
					maxdate = 5
				}
				mosque.MaxFutureDate = maxdate
				mosque.Ads = []model.Ad{}
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
			}
		} else {
			http.Redirect(response, request, "/admin", 302) // redirect back to Adminpage
		}
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func DeleteMosque(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		mosque := request.URL.Query().Get("mosque")
		collection, err := repos.GetDBCollection(1)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/admin"))
			return
		}
		collection.DeleteOne(context.TODO(), bson.M{"Name": mosque})
		collection, err = repos.GetDBCollection(0)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/admin"))
			return
		}
		update := bson.D{{Key: "$pull", Value: bson.D{{Key: "RegisteredPrayers", Value: bson.D{{Key: "MosqueName", Value: mosque}}}}}}
		collection.UpdateMany(context.TODO(), bson.M{}, update)
		response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func ShowMosque(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		mosqueName := request.URL.Query().Get("mosque")
		confirm := false
		var mosque model.Mosque
		if request.PostFormValue("confirm") == "yes" {
			confirm = true
		}
		if mosqueName != "" {
			mosque = getMosque(mosqueName)
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
							prayers = append(prayers, decryptPrayer(prayer))
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
			mosqueName := request.PostFormValue("mosque")
			mosque = getMosque(mosqueName)
			active := !mosque.Active
			collection, err := repos.GetDBCollection(1)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(dbConnectionError, "/admin"))
				return
			}
			collection.UpdateOne(context.TODO(), bson.M{"Name": mosqueName}, bson.M{"$set": bson.M{"Active": active}})
			response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
			mosque = *new(model.Mosque)
		} else {
			mosque = *new(model.Mosque)
			http.Redirect(response, request, "/admin", 300)
		}
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func ShowAllMosques(response http.ResponseWriter, request *http.Request) {
	mosques := getMosques(response, request, true)
	t, _ := template.ParseFiles("templates/show-mosques.gohtml", "templates/base_adminloggedin.tmpl", "templates/footer.tmpl")
	t.Execute(response, mosques)
}

func RegisterAdmin(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		collection, err := repos.GetDBCollection(2)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/admin"))
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
			name = repos.Encrypt(request.FormValue("name"))
		} else {
			name = request.FormValue("register-mosqueadmin")
		}
		encE := repos.Encrypt(email)
		var adminM model.Admin
		// Look if the entered Username is already used
		err = collection.FindOne(context.TODO(), bson.D{{"Email", encE}}).Decode(&adminM)
		// If not found (throws exception/error) then we can proceed, or if found but found one is not same admintype as found one we proceed
		if err != nil || adminM.Admin != admin {
			// Generate the hashed password with 14 as salt
			hash, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
			newAdmin := model.Admin{name, encE, string(hash), admin}
			// Insert user to the table
			collection.InsertOne(context.TODO(), newAdmin)
			// Change redirect target to LoginPage
			http.Redirect(response, request, "/admin", 302)
		} else {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError("Yönetici mevcut | Verwalter bereits vorhanden", "/admin"))
			return
		}
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func AddBayram(response http.ResponseWriter, request *http.Request) {
	date := request.URL.Query().Get("date")
	if date != "" {
		repos.AddEid(date)
		collection, err := repos.GetDBCollection(1)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/admin"))
			return
		}
		cur, _ := collection.Find(context.TODO(), bson.M{})
		for cur.Next(context.TODO()) {
			var mosque model.Mosque
			cur.Decode(&mosque)
			for i, dateM := range mosque.Date {
				if date == strings.Split(dateM.Date.String(), " ")[0] {
					if mosque.Bayram {
						collection.UpdateOne(context.TODO(), bson.M{"Name": mosque.Name}, bson.M{"$set": bson.M{"Date." + strconv.Itoa(i) + ".Prayer.6.Available": true}})
					}
				}
			}
		}
		response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
	}
}

func ChangeDate(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		days := request.URL.Query().Get("days")
		mosque := request.URL.Query().Get("mosque")
		daysI, _ := strconv.Atoi(days)
		if mosque != "" {
			collection, err := repos.GetDBCollection(1)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(dbConnectionError, "/admin"))
				return
			}
			collection.UpdateOne(context.TODO(), bson.M{"Name": mosque}, bson.M{"$set": bson.M{"MaxFutureDate": daysI}})
		}
	}
}

func EditPrayers(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		name := request.URL.Query().Get("mosque")
		mosque := getMosque(name)
		if name == "" {
			tod := time.Now().Format(time.RFC3339)
			today, _ := time.Parse(time.RFC3339, tod)
			fromDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
			pathUrl, _ := url.ParseQuery(request.URL.RawQuery)
			path := pathUrl.Encode()
			prayer := strings.Split(path, "=")[0]
			// TODO: If prayer = 6/7 attention! only activate on fridays/bayrams, see: adminPipieline line 150+
			value := strings.Split(path, "=")[1]
			collection, err := repos.GetDBCollection(1)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(dbConnectionError, "/admin"))
				return
			}
			available, err := strconv.ParseBool(value)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError("Yanlis boyut, yine deneyin | Falsches Format, versuchen Sie es erneut", "/admin"))
				return
			}
			dates := mosque.Date
			if prayer == "5" { // cuma
				for i, date := range dates {
					weekday := date.Date.Weekday()
					if int(weekday) == 5 { // cuma
						collection.UpdateOne(context.TODO(),
							bson.M{"Name": mosque.Name},
							bson.M{"$set": bson.M{
								"Bayram": true,
								"Date." + strconv.Itoa(i) + ".Prayer.5.Available": available}})
					}
				}
			} else if prayer == "6" { // bayram
				for i, date := range dates {
					eids := repos.GetEids()
					if containString(eids, strings.Split(date.Date.String(), " ")[0]) {
						collection.UpdateOne(context.TODO(),
							bson.M{"Name": name},
							bson.M{"$set": bson.M{"Date." + strconv.Itoa(i) + ".Prayer.6.Available": available}})
					}
				}
			} else {
				collection.UpdateMany(context.TODO(),
					bson.M{"Name": name, "Date.Date": bson.M{"$gt": fromDate}},
					bson.M{"$set": bson.M{"Date.$[].Prayer." + prayer + ".Available": available}})
			}
			mosque = *new(model.Mosque)
			response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
		} else {
			var mosqueC model.Mosque
			collection, err := repos.GetDBCollection(1)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(dbConnectionError, "/admin"))
				return
			}
			collection.FindOne(context.TODO(), bson.M{"Name": name}).Decode(&mosqueC)
			mosque = mosqueC
			dates := mosqueC.Date
			var status [7]string
			reachedToday := false
			reachedFriday := false
			reachedBayram := false
			setF := false
			setB := false
			today := strings.Split(time.Now().String(), " ")[0]
			for _, date := range dates {
				if !reachedToday && today == strings.Split(date.Date.String(), " ")[0] {
					reachedToday = true
					for i, prayer := range date.Prayer {
						if prayer.Available {
							status[i] = "Acik | Offen"
						} else {
							status[i] = "Kapali | Geschlossen"
						}
					}
				}
				if !reachedFriday && int(date.Date.Weekday()) == 5 {
					reachedFriday = true
					if date.Prayer[5].Available {
						setF = true
					}
				}
				eids := repos.GetEids()
				if !reachedBayram && containString(eids, strings.Split(date.Date.String(), " ")[0]) {
					reachedBayram = true
					if date.Prayer[6].Available {
						setB = true
					}
				}
				if reachedToday && reachedFriday && reachedBayram {
					break
				}
			}
			if setF {
				status[5] = "Acik | Offen"
			} else {
				status[5] = "Kapali | Geschlossen"
			}
			if setB {
				status[6] = "Acik | Offen"
			} else {
				status[6] = "Kapali | Geschlossen"
			}
			t, _ := template.ParseFiles("templates/editPrayers.gohtml", "templates/base_adminloggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, status)
		}
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func Edit(response http.ResponseWriter, request *http.Request) {
	fmt.Println("edit")
	if mosqueName := request.PostFormValue("mosque"); mosqueName != "" {
		mosque := getMosque(mosqueName)
		if mosque.Name != "" {
			capm := request.PostFormValue("capm")
			capw := request.PostFormValue("capw")
			capmI := 0
			capwI := 0
			pM := mosque.MaxCapM // previous capacity
			pW := mosque.MaxCapW
			if capm == "" && capw == "" {
				return
			}
			if capm == "" {
				capmI = pM
			} else {
				capmI, _ = strconv.Atoi(capm)
			}
			if capw == "" {
				capwI = pW
			} else {
				capwI, _ = strconv.Atoi(capw)
			}
			collection, err := repos.GetDBCollection(1)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(dbConnectionError, "/admin"))
				return
			}
			collection.UpdateOne(context.TODO(), bson.M{"Name": mosque.Name}, bson.M{"$set": bson.M{"MaxCapM": capmI, "MaxCapW": capwI}})
			for i, date := range mosque.Date { // check and update only if no registrations were made that day
				for j, prayer := range date.Prayer {
					newCapM := capmI - (pM - prayer.CapacityMen)
					newCapW := capwI - (pW - prayer.CapacityWomen)
					collection.UpdateMany(context.TODO(), bson.M{"Name": mosque.Name}, bson.M{"$set": bson.M{"Date." + strconv.Itoa(i) + ".Prayer." + strconv.Itoa(j) + ".CapacityMen": newCapM}})
					collection.UpdateMany(context.TODO(), bson.M{"Name": mosque.Name}, bson.M{"$set": bson.M{"Date." + strconv.Itoa(i) + ".Prayer." + strconv.Itoa(j) + ".CapacityWomen": newCapW}})
				}
			}
		} else {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError("Camii ismi gecerli degil | Moscheename ungültig", "/admin"))
			return
		}
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Camii ismi gecerli degil | Moscheename ungültig", "/admin"))
		return
	}
	response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
}

func EditCapacity(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		mosqueName := request.URL.Query().Get("mosque")
		if mosqueName != "" {
			mosque := getMosque(mosqueName)
			type nMosque struct {
				Name     string
				CurrentM int
				CurrentW int
				MinM     int
				MinW     int
			}

			var newMosque nMosque
			newMosque.Name = mosque.Name
			newMosque.CurrentM = mosque.MaxCapM
			newMosque.CurrentW = mosque.MaxCapW
			minM := 0
			minW := 0

			for _, date := range mosque.Date {
				for _, prayer := range date.Prayer {
					usedM := mosque.MaxCapM - prayer.CapacityMen
					usedW := mosque.MaxCapW - prayer.CapacityWomen
					if usedM > minM {
						minM = usedM
					}
					if usedW > minW {
						minW = usedW
					}
				}
			}
			newMosque.MinM = minM
			newMosque.MinW = minW
			t, _ := template.ParseFiles("templates/editCapacity.gohtml", "templates/base_adminloggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, newMosque)
		} else {
			response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
		}
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func ShowAdmins(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		admin, _ := strconv.ParseBool(request.URL.Query().Get("admin"))
		t, _ := template.ParseFiles("templates/show-admins.gohtml", "templates/base_adminloggedin.tmpl", "templates/footer.tmpl")
		var admins []model.Admin
		if admin {
			admins = getAdmins()
		} else {
			admins = getMosqueAdmins()
		}
		t.Execute(response, admins)
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func ChangeAdmin(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		name := request.URL.Query().Get("name")
		oldEmail := request.URL.Query().Get("email")
		email := request.URL.Query().Get("nemail")
		password := request.URL.Query().Get("password")
		admin, _ := strconv.ParseBool(request.URL.Query().Get("admin"))
		encN := ""
		if admin {
			encN = repos.Encrypt(name)
		} else {
			encN = name
		}
		encOe := repos.Encrypt(oldEmail)
		var adminModel model.Admin
		collection, err := repos.GetDBCollection(2)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/admin"))
			return
		}
		collection.FindOne(context.TODO(), bson.M{"Name": encN}).Decode(&adminModel)
		encE := ""
		if email == "" {
			encE = adminModel.Email
		} else {
			encE = repos.Encrypt(email)
		}
		hash := []byte{}
		if password == "" {
			hash = []byte(adminModel.Password)
		} else {
			hash, _ = bcrypt.GenerateFromPassword([]byte(password), 14)
		}
		newAdmin := model.Admin{encN, encE, string(hash), admin}
		collection.ReplaceOne(context.TODO(), bson.M{"Email": encOe}, newAdmin)
		response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func AddBanner(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		request.ParseMultipartForm(10 << 20)
		file, handler, err := request.FormFile("file") //retrieve the file from form data
		name := request.PostFormValue("mosque")
		link := request.PostFormValue("link")
		//replace file with the key your sent your image with
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError("Fotograf yüklenilemedi, birdaha deneyin | Foto konnte nicht geladen werden, versuchen Sie es erneut", "/admin"))
		}

		defer file.Close() //close the file when we finish
		//this is path which  we want to store the file
		f, err := os.OpenFile("banner/"+name+" "+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
		}
		defer f.Close()
		io.Copy(f, file)
		var ad model.Ad
		ad.Path = name + " " + handler.Filename
		ad.Link = link
		collection, err := repos.GetDBCollection(1)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/admin"))
			return
		}
		collection.UpdateOne(context.TODO(), bson.M{"Name": name}, bson.M{"$push": bson.M{"Ads": ad}})
		response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func RemoveBanner(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "admin") {
		path := request.PostFormValue("path")
		name := request.PostFormValue("mosque")
		if path != "" {
			collection, err := repos.GetDBCollection(1)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(dbConnectionError, "/admin"))
				return
			}
			collection.UpdateOne(context.TODO(), bson.M{"Name": name}, bson.M{"$pull": bson.M{"Ads": bson.M{"Path": path}}})
			response.Write([]byte(`<script>window.location.href = "/admin";</script>`))
		} else {
			var mosque = getMosque(name)
			type adM struct {
				Ads  []model.Ad
				Name string
			}
			var ads adM
			ads.Ads = mosque.Ads
			ads.Name = name
			t, _ := template.ParseFiles("templates/removeBanner.gohtml", "templates/base_adminloggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, ads)
		}
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func getAdmins() []model.Admin {
	var admins []model.Admin
	collection, err := repos.GetDBCollection(2)
	if err != nil {
		return admins
	}
	cur, _ := collection.Find(context.TODO(), bson.M{})
	for cur.Next(context.TODO()) {
		var adminModel model.Admin
		cur.Decode(&adminModel)
		if adminModel.Admin {
			admins = append(admins, decryptAdmin(adminModel))
		}
	}
	return admins
}

func getMosqueAdmins() []model.Admin {
	var admins []model.Admin
	collection, err := repos.GetDBCollection(2)
	if err != nil {
		return admins
	}
	cur, _ := collection.Find(context.TODO(), bson.M{})
	for cur.Next(context.TODO()) {
		var adminModel model.Admin
		cur.Decode(&adminModel)
		if !adminModel.Admin {
			admins = append(admins, decryptMosque(adminModel))
		}
	}
	return admins
}

func decryptMosque(admin model.Admin) model.Admin {
	dA := admin
	dA.Name = admin.Name
	dA.Email = repos.Decrypt(admin.Email)
	return dA
}

func containString(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}
