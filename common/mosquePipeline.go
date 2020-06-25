package common

import (
	"context"
	"html/template"
	"net/http"
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

//var mosquePipe MosquePipeline

func decryptPrayer(prayer model.Prayer) model.Prayer {
	dP := prayer
	dP.Users = []model.User{}
	for _, user := range prayer.Users {
		dP.Users = append(dP.Users, decryptUser(user))
	}
	return dP
}

func MosqueHandler(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		collection, err := repos.GetDBCollection(1)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/mosqueIndex"))
			return
		}
		var mosque model.Mosque
		name, err := GetPhoneFromCookie(request)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(err.Error(), "/"))
			return
		}
		collection.FindOne(context.TODO(), bson.M{"Name": name}).Decode(&mosque)
		date := time.Now()
		today := strconv.Itoa(date.Day()) + "." + strconv.Itoa(int(date.Month())) + "." + strconv.Itoa(date.Year())
		var prayers []model.Prayer
		for _, date := range mosque.Date {
			if today == strconv.Itoa(date.Date.Day())+"."+strconv.Itoa(int(date.Date.Month()))+"."+strconv.Itoa(date.Date.Year()) {
				for _, prayer := range date.Prayer {
					if len(prayer.Users) > 0 {
						prayers = append(prayers, decryptPrayer(prayer))
					}
				}
				break
			}
		}
		var tmpDate Date
		tmpDate.Date = today
		tmpDate.Prayer = prayers
		mosquePipe := MosquePipeline{mosque, tmpDate, false, false, false}
		t, _ := template.ParseFiles("templates/mosque.gohtml", "templates/base_mosqueloggedin.tmpl", "templates/footer.tmpl")
		t.Execute(response, mosquePipe)
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
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
			mosqueName, err := GetPhoneFromCookie(request)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(err.Error(), "/"))
				return
			}
			mosque := getMosque(mosqueName)
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
					nprayer := prayer
					nprayer.Users = []model.User{}
					for _, user := range prayer.Users {
						if user.Attended {
							nprayer.Users = append(nprayer.Users, user)
						}
					}
					prayers = append(prayers, decryptPrayer(nprayer))
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
			mosqueName, err := GetPhoneFromCookie(request)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(err.Error(), "/"))
				return
			}
			mosque := getMosque(mosqueName)
			if mosque.Name != "" {
				for _, date := range mosque.Date {
					for _, prayer := range date.Prayer {
						if len(prayer.Users) > 0 {
							nprayer := prayer
							nprayer.Users = []model.User{}
							for _, user := range prayer.Users {
								if user.Attended {
									nprayer.Users = append(nprayer.Users, user)
								}
							}
							prayers = append(prayers, decryptPrayer(nprayer))
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
			} else {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError("Camii bulunamadi | Moschee konnte nicht gefunden werden", "/mosqueIndex"))
			}
		}
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func ConfirmVisitors(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		request.ParseForm()
		visitors := request.Form["visitor"]
		mosqueName, err := GetPhoneFromCookie(request)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(err.Error(), "/"))
			return
		}
		mosque := getMosque(mosqueName)
		if mosque.Name != "" {
			if len(visitors) > 0 {
				if request.URL.Query().Get("type") == "add" {
					data := strings.Split(request.URL.Query().Get("data"), "!")
					for _, phone := range visitors {
						today := strings.Split(time.Now().String(), " ")[0]
						index := 0
						for i, dateI := range mosque.Date {
							if today == strings.Split(dateI.Date.String(), " ")[0] {
								index = i
							}
						}
						collection, err := repos.GetDBCollection(1)
						if err != nil {
							t, _ := template.ParseFiles("templates/errorpage.gohtml")
							t.Execute(response, GetError(dbConnectionError, "/mosqueIndex"))
							return
						}
						in, _ := strconv.Atoi(data[1])
						ind := strconv.Itoa(in - 1)
						encP := repos.Encrypt(phone)
						collection.UpdateOne(context.TODO(),
							bson.M{"Name": data[0], "Date." + strconv.Itoa(index) + ".Prayer." + ind + ".Users.Phone": encP},
							bson.M{"$set": bson.M{"Date." + strconv.Itoa(index) + ".Prayer." + ind + ".Users.$.Attended": true}})
						response.Write([]byte(`<script>window.location.href = "/mosqueIndex";</script>`))
					}
				} else {
					data := strings.Split(request.URL.Query().Get("data"), "!")
					for _, phone := range visitors {
						today := strings.Split(time.Now().String(), " ")[0]
						index := 0
						for i, dateI := range mosque.Date {
							if today == strings.Split(dateI.Date.String(), " ")[0] {
								index = i
							}
						}
						collection, err := repos.GetDBCollection(1)
						if err != nil {
							t, _ := template.ParseFiles("templates/errorpage.gohtml")
							t.Execute(response, GetError(dbConnectionError, "/mosqueIndex"))
							return
						}
						in, err := strconv.Atoi(data[1])
						if err != nil {
							t, _ := template.ParseFiles("templates/errorpage.gohtml")
							t.Execute(response, GetError("Sayi dönüsümde hata | Fehler beim umwandeln", "/mosqueIndex"))
							return
						}
						ind := strconv.Itoa(in - 1)
						encP := repos.Encrypt(phone)
						collection.UpdateOne(context.TODO(),
							bson.M{"Name": data[0], "Date." + strconv.Itoa(index) + ".Prayer." + ind + ".Users.Phone": encP},
							bson.M{"$set": bson.M{"Date." + strconv.Itoa(index) + ".Prayer." + ind + ".Users.$.Attended": false}})
						response.Write([]byte(`<script>window.location.href = "/mosqueIndex";</script>`))
					}
				}
			} else {
				response.Write([]byte(`<script>window.location.href = "/mosqueIndex";</script>`))
			}
		} else {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError("Camii bulunamadi | Moschee konnte nicht gefunden werden", "/mosqueIndex"))
		}
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}
