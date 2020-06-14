package common

import (
	"context"
	"html/template"
	"net/http"
	"net/url"
	"pi-software/model"
	"pi-software/repos"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type MosquePipeline struct {
	Mosque       model.Mosque
	Date         Date
	EditPrayers  bool // add/delete prayer, f.ex. add cuma prayer
	EditCapacity bool // edit capacity for mosque, for women and men
	GetDate      bool // get users registered for a choosen date
}

type prayerBool struct {
	Fajr    bool
	Dhuhr   bool
	Asr     bool
	Maghrib bool
	Ishaa   bool
}

type tempMosque struct {
	Name string
	Date Date
}

var prayerB prayerBool

var mosquePipe MosquePipeline

func MosqueHandler(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		collection, _ := repos.GetDBCollection(1)
		var mosque model.Mosque
		name, _ := GetPhoneFromCookie(request)
		collection.FindOne(context.TODO(), bson.M{"Name": name}).Decode(&mosque)
		date := time.Now()
		today := strconv.Itoa(date.Day()) + "." + strconv.Itoa(int(date.Month())) + "." + strconv.Itoa(date.Year())
		var prayers []model.Prayer
		for _, date := range mosque.Date {
			if today == strconv.Itoa(date.Date.Day())+"."+strconv.Itoa(int(date.Date.Month()))+"."+strconv.Itoa(date.Date.Year()) {
				for _, prayer := range date.Prayer {
					if len(prayer.Users) > 0 {
						prayers = append(prayers, prayer)
					}
				}
				break
			}
		}
		var tmpDate Date
		tmpDate.Date = today
		tmpDate.Prayer = prayers
		mosquePipe = MosquePipeline{mosque, tmpDate, false, false, false}
		t, _ := template.ParseFiles("templates/mosque.gohtml", "templates/base_mosqueloggedin.tmpl", "templates/footer.tmpl")
		t.Execute(response, mosquePipe)
	} else {
		accessError(response, request)
	}
}

func GetRegistrations(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		request.ParseForm()
		dateG := request.URL.Query().Get("date")
		// For a specific Date
		if dateG != "" {
			var choosenDate Date
			var prayers []model.Prayer
			mosque := mosquePipe.Mosque
			dates := mosque.Date
			for _, date := range dates {
				if dateG == strings.Split(date.Date.String(), " ")[0] {
					choosenDate.Date = strconv.Itoa(date.Date.Day()) + "." + strconv.Itoa(int(date.Date.Month())) + "." + strconv.Itoa(date.Date.Year())
					choosenDate.Prayer = date.Prayer
					break
				}
			}
			for _, prayer := range choosenDate.Prayer {
				if len(prayer.Users) > 0 {
					prayers = append(prayers, prayer)
				}
			}
			choosenDate.Prayer = prayers
			tmpDate := make([]Date, 1)
			tmpDate[0].Date = dateG
			if choosenDate.Date != "" {
				tmpDate[0] = choosenDate
			}
			t, _ := template.ParseFiles("templates/getRegistrations.gohtml", "templates/base_mosqueloggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, tmpDate)
		} else { // For all dates
			var datesMosque []Date
			var prayers []model.Prayer
			for _, date := range mosquePipe.Mosque.Date {
				for _, prayer := range date.Prayer {
					if len(prayer.Users) > 0 {
						prayers = append(prayers, prayer)
					}
				}
				if len(prayers) > 0 {
					var dat Date
					dateS := strconv.Itoa(date.Date.Day()) + "." + strconv.Itoa(int(date.Date.Month())) + "." + strconv.Itoa(date.Date.Year())
					dat.Date = dateS
					dat.Prayer = prayers
					prayers = []model.Prayer{}
					datesMosque = append(datesMosque, dat)
				}
			}
			t, _ := template.ParseFiles("templates/getRegistrations.gohtml", "templates/base_mosqueloggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, datesMosque)
		}
	} else {
		accessError(response, request)
	}
}

func ConfirmVisitors(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		request.ParseForm()
		visitors := request.Form["visitor"]

		data := strings.Split(request.URL.Query().Get("data"), "!")
		for _, phone := range visitors {
			today := strings.Split(time.Now().String(), " ")[0]
			index := 0
			for i, dateI := range mosquePipe.Mosque.Date {
				if today == strings.Split(dateI.Date.String(), " ")[0] {
					index = i
				}
			}
			collection, _ := repos.GetDBCollection(1)
			in, _ := strconv.Atoi(data[1])
			ind := strconv.Itoa(in - 1)
			collection.UpdateOne(context.TODO(),
				bson.M{"Name": data[0], "Date." + strconv.Itoa(index) + ".Prayer." + ind + ".Users.Phone": phone},
				bson.M{"$set": bson.M{"Date." + strconv.Itoa(index) + ".Prayer." + ind + ".Users.$.Attended": true}})
			response.Write([]byte(`<script>window.location.href = "/mosqueIndex";</script>`))
		}
	} else {
		accessError(response, request)
	}
}

func EditPrayers(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		if request.URL.RawQuery != "" {
			tod := time.Now().Format(time.RFC3339)
			today, _ := time.Parse(time.RFC3339, tod)
			fromDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
			pathUrl, _ := url.ParseQuery(request.URL.RawQuery)
			path := pathUrl.Encode()
			prayer := strings.Split(path, "=")[0]
			// TODO: If prayer = 6/7 attention! only activate on fridays/bayrams, see: adminPipieline line 150+
			value := strings.Split(path, "=")[1]
			collection, _ := repos.GetDBCollection(1)
			available, err := strconv.ParseBool(value)
			if err != nil {
				http.Error(response, "Wrong parameter", 402)
			}
			var mosque model.Mosque
			name, _ := GetPhoneFromCookie(request)
			collection.FindOne(context.TODO(), bson.M{"Name": name}).Decode(&mosque)
			dates := mosque.Date
			if prayer == "5" { // cuma
				for i, date := range dates {
					weekday := date.Date.Weekday()
					if int(weekday) == 5 { // cuma
						collection.UpdateOne(context.TODO(),
							bson.M{"Name": name},
							bson.M{"$set": bson.M{"Date." + strconv.Itoa(i) + ".Prayer.5.Available": available}})
					}
				}
			} else if prayer == "6" { // bayram
				for i, date := range dates {
					if containString(model.GetBayrams(), strings.Split(date.Date.String(), " ")[0]) {
						collection.UpdateOne(context.TODO(),
							bson.M{"Name": name},
							bson.M{"$set": bson.M{"Date." + strconv.Itoa(i) + ".Prayer.6.Available": available}})
					}
				}
			} else {
				collection.UpdateMany(context.TODO(),
					bson.M{"Name": mosquePipe.Mosque.Name, "Date.Date": bson.M{"$gt": fromDate}},
					bson.M{"$set": bson.M{"Date.$[].Prayer." + prayer + ".Available": available}})
			}
			response.Write([]byte(`<script>window.location.href = "/mosqueIndex";</script>`))
		} else {
			mosque := mosquePipe.Mosque
			dates := mosque.Date
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
				if !reachedBayram && containString(model.GetBayrams(), strings.Split(date.Date.String(), " ")[0]) {
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
			t, _ := template.ParseFiles("templates/editPrayers.gohtml", "templates/base_mosqueloggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, status)
		}
	} else {
		accessError(response, request)
	}
}

func EditCapacity(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		capm := request.FormValue("capm")
		capw := request.FormValue("capw")
		capmI := 0
		capwI := 0
		if capm == "" && capw == "" {
			return
		}
		if capm == "" {
			capmI = mosquePipe.Mosque.MaxCapM
		} else {
			capmI, _ = strconv.Atoi(capm)
		}
		if capw == "" {
			capwI = mosquePipe.Mosque.MaxCapW
		} else {
			capwI, _ = strconv.Atoi(capw)
		}
		collection, _ := repos.GetDBCollection(1)
		var mosque model.Mosque
		collection.FindOne(context.TODO(), bson.M{"Name": mosquePipe.Mosque.Name}).Decode(&mosque)
		today := strings.Split(time.Now().String(), " ")[0]
		dates := mosque.Date
		reachedToday := false
		for _, date := range dates {
			for _, prayer := range date.Prayer {
				if today == strings.Split(date.Date.String(), " ")[0] {
					reachedToday = true
				}
				if reachedToday && len(prayer.Users) > 0 {
					http.Error(response, "Kayitlar var, Kapasite degisitirelemez | Anmeldungen vorhanden, keine Änderung der Kapazität möglich", 402)
					return
				}
			}
		}
		collection.UpdateOne(context.TODO(), bson.M{"Name": mosquePipe.Mosque.Name}, bson.M{"$set": bson.M{"MaxCapM": capmI, "MaxCapW": capwI}})
		// attention: also changes available capacity in past  actually no problem, as not used
		collection.UpdateMany(context.TODO(), bson.M{"Name": mosquePipe.Mosque.Name}, bson.M{"$set": bson.M{"Date.$[].Prayer.$[].CapacityMen": capmI, "Date.$[].Prayer.$[].CapacityWomen": capwI}})
		response.Write([]byte(`<script>window.location.href = "/mosqueIndex";</script>`))
	}
}
