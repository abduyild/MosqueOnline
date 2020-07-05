package common

import (
	"context"
	"html/template"
	"net/http"
	"pi-software/helpers"
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
	Found        bool
	Register     bool
	User         model.User
	Prayers      []model.Prayer
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
		var prayersA []model.Prayer
		for _, date := range mosque.Date {
			if today == strconv.Itoa(date.Date.Day())+"."+strconv.Itoa(int(date.Date.Month()))+"."+strconv.Itoa(date.Date.Year()) {
				for _, prayer := range date.Prayer {
					if prayer.Available {
						if len(prayer.Users) > 0 {
							prayers = append(prayers, decryptPrayer(prayer))
						}
						prayersA = append(prayersA, decryptPrayer(prayer))
					}
				}
				break
			}
		}
		var tmpDate Date
		tmpDate.Date = today
		tmpDate.Prayer = prayers
		var user model.User
		registered := false
		found := false
		if phone := request.PostFormValue("phone"); phone != "" {
			registered = true
			_, err := strconv.Atoi(phone)
			if err != nil {
				http.Redirect(response, request, "/mosqueIndex?format", 302)
				return
			}
			user.Phone = phone
			if len(request.PostForm) > 1 {
				collection, err := repos.GetDBCollection(0)
				if err != nil {
					t, _ := template.ParseFiles("templates/errorpage.gohtml")
					t.Execute(response, GetError(dbConnectionError, "/mosqueIndex"))
					return
				}
				firstName := request.FormValue("firstname")
				lastName := request.FormValue("lastname")
				email := request.FormValue("email")
				sex := request.FormValue("sex")
				_firstName, _lastName, _email, _phone := false, false, false, false
				_firstName = !helpers.IsEmpty(firstName)
				_lastName = !helpers.IsEmpty(lastName)
				_email = !helpers.IsEmpty(email)
				_phone = !helpers.IsEmpty(phone)
				if _firstName && _lastName && _email && _phone {
					result := collection.FindOne(context.TODO(), bson.D{{"Phone", repos.Encrypt(phone)}})
					if result.Err() != nil {
						encF := repos.Encrypt(firstName)
						if encF == "" {
							http.Redirect(response, request, "/mosqueIndex?format", 302)
							return
						}
						encL := repos.Encrypt(lastName)
						if encL == "" {
							http.Redirect(response, request, "/mosqueIndex?format", 302)
							return
						}
						encE := repos.Encrypt(email)
						if encE == "" {
							http.Redirect(response, request, "/mosqueIndex?format", 302)
							return
						}
						encP := repos.Encrypt(phone)
						if encP == "" {
							http.Redirect(response, request, "/mosqueIndex?format", 302)
							return
						}
						usr := model.User{sex, encF, encL, encE, encP, false, []model.RegisteredPrayer{}}

						//test autoattent start
						var reg register
						reg.Mosque = getMosque(name)
						if reg.Mosque.Name != "" {
							choosenMosque := reg.Mosque
							if err != nil {
								t, _ := template.ParseFiles("templates/errorpage.gohtml")
								t.Execute(response, GetError("Cerez hatasi | Cookiefehler", "/"))
								return
							}
							registered := model.RegisteredPrayer{}
							var mosque = getMosque(choosenMosque.Name)
							index := 0
							for i, dates := range choosenMosque.Date {
								if strings.Split(date.String(), " ")[0] == strings.Split(dates.Date.String(), " ")[0] {
									registered.Date = strconv.Itoa(dates.Date.Day()) + "." + strconv.Itoa(int(dates.Date.Month())) + "." + strconv.Itoa(dates.Date.Year())
									index = i
									break
								}
							}
							prayer := request.PostFormValue("prayer")
							prayerI, err := strconv.Atoi(prayer)
							if err != nil {
								t, _ := template.ParseFiles("templates/errorpage.gohtml")
								t.Execute(response, GetError("Yanlis sayi boyutu | Falsches Zahlenformat", "/mosqueIndex"))
								return
							}
							registered.RpId = mosque.Name + ":" + strconv.Itoa(index) + ":" + prayer
							result := collection.FindOne(context.TODO(), bson.D{
								{"Phone", usr.Phone},
								{"RegisteredPrayers.RpId", registered.RpId}})
							if result.Err() != nil {
								registered.MosqueName = mosque.Name
								registered.MosqueAddress = strconv.Itoa(mosque.PLZ) + " " + mosque.City + ", " + mosque.Street
								registered.DateIndex = index
								switch prayerI {
								case 1:
									registered.PrayerName = "Sabah"
								case 2:
									registered.PrayerName = "Ögle"
								case 3:
									registered.PrayerName = "Ikindi"
								case 4:
									registered.PrayerName = "Aksam"
								case 5:
									registered.PrayerName = "Yatsi"
								case 6:
									registered.PrayerName = "Cuma"
								case 7:
									registered.PrayerName = "Bayram"
								}
								registered.PrayerIndex = prayerI
								collection, err = repos.GetDBCollection(1)
								if err != nil {
									t, _ := template.ParseFiles("templates/errorpage.gohtml")
									t.Execute(response, GetError(dbConnectionError, "/mosqueIndex"))
									return
								}
								collection.UpdateOne(context.TODO(),
									bson.M{"Name": mosque.Name},
									bson.D{{"$inc", bson.D{
										{"Date." + strconv.Itoa(index) + ".Prayer." + strconv.Itoa(prayerI-1) + ".Capacity" + usr.Sex, -1},
									},
									}})
								tempUser := usr
								tempUser.RegisteredPrayers = []model.RegisteredPrayer{}
								collection.UpdateOne(context.TODO(),
									bson.M{"Name": mosque.Name}, bson.M{"$push": bson.M{"Date." + strconv.Itoa(index) + ".Prayer." + strconv.Itoa(prayerI-1) + ".Users": tempUser}})

								collection.UpdateOne(context.TODO(),
									bson.M{"Name": mosque.Name, "Date." + strconv.Itoa(index) + ".Prayer." + strconv.Itoa(prayerI-1) + ".Users.Phone": encP},
									bson.M{"$set": bson.M{"Date." + strconv.Itoa(index) + ".Prayer." + strconv.Itoa(prayerI-1) + ".Users.$.Attended": true}})

								usr.RegisteredPrayers = append(usr.RegisteredPrayers, registered)
								collection, err = repos.GetDBCollection(0)
								if err != nil {
									t, _ := template.ParseFiles("templates/errorpage.gohtml")
									t.Execute(response, GetError(dbConnectionError, "/mosqueIndex"))
									return
								}
								if err != nil {
									t, _ := template.ParseFiles("templates/errorpage.gohtml")
									t.Execute(response, GetError(err.Error(), "/"))
									return
								}
								usr.RegisteredPrayers = []model.RegisteredPrayer{}
								usr.RegisteredPrayers = append(usr.RegisteredPrayers, registered)
								collection.InsertOne(context.TODO(), usr) // test if this works
								/*collection.UpdateOne(context.TODO(),
								bson.M{"Phone": usr.Phone}, bson.M{
									"$push": bson.M{"RegisteredPrayers": registered}})*/
								http.Redirect(response, request, "/mosqueIndex?success", 302)
							}
						} else {
							t, _ := template.ParseFiles("templates/errorpage.gohtml")
							t.Execute(response, GetError("Camii secilmedi | Keine Moschee asugewählt", "/mosqueIndex"))
						}
						//test autoattent end

						http.Redirect(response, request, "/mosqueIndex?success", 302)
					} else {
						http.Redirect(response, request, "/mosqueIndex", 302)
					}
				}
			} else {
				collection, err := repos.GetDBCollection(0)
				if err != nil {
					t, _ := template.ParseFiles("templates/errorpage.gohtml")
					t.Execute(response, GetError(dbConnectionError, "/mosqueIndex"))
					return
				}
				encP := repos.Encrypt(phone)
				if encP == "" {
					http.Redirect(response, request, "/mosqueIndex?format", 302)
					return
				}
				result := collection.FindOne(context.TODO(), bson.D{{"Phone", encP}})
				if result.Err() != nil {
					found = false
				} else {
					found = true
					collection.FindOne(context.TODO(), bson.D{{"Phone", encP}}).Decode(&user)
					user = decryptUser(user)
				}
			}
		}
		mosquePipe := MosquePipeline{mosque, tmpDate, false, false, false, found, registered, user, prayersA}
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
			today := strings.Split(time.Now().String(), " ")[0]
			mosque := getMosque(mosqueName)
			if mosque.Name != "" {
				for _, date := range mosque.Date {
					if today == strings.Split(date.Date.String(), " ")[0] {
						break
					}
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
