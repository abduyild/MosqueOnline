package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"pi-software/helpers"
	"pi-software/model"
	"pi-software/repos"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

const (
	dbConnectionError = "Veribanka hatasi | Datenbankfehler"
)

type Mosque model.Mosque

type mosques []model.Mosque

type users []model.User

type choose struct {
	Name          string
	City          string
	Street        string
	PLZ           int
	SetPrayer     bool
	Mosques       []model.Mosque
	Date          model.Date
	DateString    string
	Prayer        []TempPrayer
	PrayerName    string
	MaxFutureDate int
}

type ErrorMessage struct {
	Error string
	Link  string
}

type chooseCookie struct {
	N string //Name
}

// Struct for easy handling the html template generation on the index page
type TempPrayer struct {
	Name      model.PrayerName
	Capacity  int
	Available bool
}

var emptyChoose = choose{"", "", "", 0, false, *new([]model.Mosque), *new(model.Date), "", *new([]TempPrayer), "", 0}

var emptyCookie = chooseCookie{""}

var isAdmin bool
var mosquesList []model.Mosque

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

// Function for handling the Webpage call with GET
func RegisterPageHandler(response http.ResponseWriter, request *http.Request) {
	t, _ := template.ParseFiles("templates/register.gohtml", "templates/base.tmpl", "templates/footer.tmpl")
	t.Execute(response, nil)
}

// Function for handling the register action submitted by user
func RegisterHandler(response http.ResponseWriter, request *http.Request) {
	collection, err := repos.GetDBCollection(0)
	if err != nil {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError(dbConnectionError, "/register"))
		return
	}
	request.ParseForm()

	// Get data the User typen into the fields
	firstName := request.FormValue("firstname")
	lastName := request.FormValue("lastname")
	email := request.FormValue("email")
	phone := request.FormValue("phone")
	sex := request.FormValue("sex")

	// initializing as false for not filled
	_firstName, _lastName, _email, _phone := false, false, false, false
	_firstName = !helpers.IsEmpty(firstName)
	_lastName = !helpers.IsEmpty(lastName)
	_email = !helpers.IsEmpty(email)
	_phone = !helpers.IsEmpty(phone)
	// Check if fields are not empty
	if _firstName && _lastName && _email && _phone {
		// Look if the entered Username is already used
		result := collection.FindOne(context.TODO(), bson.D{{"Phone", repos.Encrypt(phone)}})
		// If not found (throws exception/error) then we can proceed
		if result.Err() != nil {
			encF := repos.Encrypt(firstName)
			if encF == "" {
				http.Redirect(response, request, "/register?format", 302)
				return
			}
			encL := repos.Encrypt(lastName)
			if encL == "" {
				http.Redirect(response, request, "/register?format", 302)
				return
			}
			encE := repos.Encrypt(email)
			if encE == "" {
				http.Redirect(response, request, "/register?format", 302)
				return
			}
			encP := repos.Encrypt(phone)
			if encP == "" {
				http.Redirect(response, request, "/register?format", 302)
				return
			}
			usr := model.User{sex, encF, encL, encE, encP, false, []model.RegisteredPrayer{}}
			// Insert user to the table
			collection.InsertOne(context.TODO(), usr)
			// Change redirect target to LoginPage
			http.Redirect(response, request, "/?success", 302)
		} else {
			http.Redirect(response, request, "/register?wrong", 302)
			return
		}
	} else {
		http.Redirect(response, request, "/register?empty", 302)
		return
	}
}

// Handler for Login Page used with POST by submitting Loginform
func LoginHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	mosque := request.URL.Query().Get("mosque")
	if mosque != "" {
		SetChoo(chooseCookie{mosque}, response)
		collection, _ := repos.GetDBCollection(1)
		var mosqueM model.Mosque
		collection.FindOne(context.TODO(), bson.M{"Name": mosque}).Decode(&mosqueM)
		ads := mosqueM.Ads
		t, _ := template.ParseFiles("templates/userlogin.gohtml", "templates/base.tmpl", "templates/footer.tmpl")
		t.Execute(response, ads)
	} else if len(request.PostForm) > 0 {
		if request.FormValue("type") == "admin" {
			adminLogin(response, request)
			return
		}
		email := request.FormValue("email")
		phone := request.FormValue("phone")
		redirectTarget := "/"
		if len(email) != 0 && len(phone) != 0 {
			collection, err := repos.GetDBCollection(0)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(dbConnectionError, "/"))
				return
			}
			var user model.User

			encP := repos.Encrypt(phone)

			err = collection.FindOne(context.TODO(), bson.D{{"Phone", encP}}).Decode(&user)
			if err != nil {
				http.Redirect(response, request, "/?wrong", 302)
				return
			}
			decE := repos.Decrypt(user.Email)
			if email != decE {
				http.Redirect(response, request, "/?wrong", 302)
				return
			}
			userCredentials, err := bcrypt.GenerateFromPassword([]byte(R(email+phone+"!")), 14)
			if err != nil {
				http.Redirect(response, request, "/?wrong", 302)
				return
			}
			cookie := R(email + "?" + phone + "!" + "&" + string(userCredentials))
			SetCookie(cookie, response)
			redirectTarget = "/index"
		}
		http.Redirect(response, request, redirectTarget, 302)
	} else {
		t, _ := template.ParseFiles("templates/login.gohtml", "templates/base.tmpl", "templates/footer.tmpl")
		t.Execute(response, getMosques(response, request, false))
	}
}

func decryptAdmin(admin model.Admin) model.Admin {
	dA := admin
	dA.Name = repos.Decrypt(admin.Name)
	dA.Email = repos.Decrypt(admin.Email)
	return dA
}

func adminLogin(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	if len(request.PostForm) > 0 {
		email := request.FormValue("email")
		password := request.FormValue("password")
		// Default redirect page is the login page, so if anything goes wrong, the program just redirects to the login page again
		redirectTarget := "/"
		if len(email) != 0 && len(password) != 0 {
			// Returns Table
			collection, err := repos.GetDBCollection(2)

			// if there was no error getting the table, te program does these operations
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(dbConnectionError, "/index"))
			}
			var admin model.Admin
			// Checking if typed in Username exists, if not redirect to register page
			encE := repos.Encrypt(email)
			err = collection.FindOne(context.TODO(), bson.D{{Key: "Email", Value: encE}}).Decode(&admin)
			// If there was an error getting an entry with matching username (no user with this username) redirect to faultpage
			if err != nil {
				http.Redirect(response, request, "/?wrong", 302)
				return
			}

			err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
			if err != nil {
				http.Redirect(response, request, "/?wrong", 302)
				return
			}

			name := admin.Name
			adminType := ""
			if admin.Admin {
				redirectTarget = "/admin"
				adminType = "!admin"
			} else {
				redirectTarget = "/mosqueIndex"
				adminType = "!mosque"
			}

			userCredentials, err := bcrypt.GenerateFromPassword([]byte(R(email+name)), 14)
			cookie := R(email+"?"+name+"&"+string(userCredentials)) + adminType
			SetCookie(cookie, response)
			// If the admin tries to login, change the redirect to the Adminpage

			// function for redirecting
			http.Redirect(response, request, redirectTarget, 302)
		} else {
			t, _ := template.ParseFiles("templates/login.gohtml", "templates/base.tmpl", "templates/footer.tmpl")
			t.Execute(response, nil)
		}
	} else {
		t, _ := template.ParseFiles("templates/login.gohtml", "templates/base.tmpl", "templates/footer.tmpl")
		t.Execute(response, nil)
	}
}

func decryptUser(encryptedUser model.User) model.User {
	dU := encryptedUser
	dU.FirstName = repos.Decrypt(encryptedUser.FirstName)
	dU.LastName = repos.Decrypt(encryptedUser.LastName)
	dU.Email = repos.Decrypt(encryptedUser.Email)
	dU.Phone = repos.Decrypt(encryptedUser.Phone)
	return dU
}

func IndexPageHandler(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		encryptedUser, err := GetUserAsUser(response, request)
		if err != nil {
			fmt.Println("oh")
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError("Cerez hatasi | Cookiefehler", "/"))
			return
		}
		user := decryptUser(encryptedUser)
		var tUser model.User // for only adding the actual prayers, not past ones
		tUser = user
		var regP []model.RegisteredPrayer
		tUser.RegisteredPrayers = regP
		var mosque model.Mosque
		collection, err := repos.GetDBCollection(1)

		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/index"))
			return
		}

		for _, reg := range user.RegisteredPrayers {
			collection.FindOne(context.TODO(), bson.M{"Name": reg.MosqueName}).Decode(&mosque)
			if mosque.Active { // only show registrations if mosque is active
				if reg.Date != "" {
					timeNow := time.Now()
					if stringToTime(reg.Date).After(timeNow) {
						regP = append(regP, reg)
					}
				}
			}
		}
		sort.Slice(regP, func(i int, j int) bool {
			return stringToTime(regP[i].Date).Before(stringToTime(regP[j].Date))
		})
		tUser.RegisteredPrayers = regP
		t, err := template.ParseFiles("templates/index.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
		t.Execute(response, tUser)
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func stringToTime(date string) time.Time {
	regT := strings.Split(date, ".")
	m, _ := strconv.Atoi(regT[1])
	d, _ := strconv.Atoi(regT[0])
	regT[1] = fmt.Sprintf("%02d", m)
	regT[0] = fmt.Sprintf("%02d", d)
	regToday, _ := time.Parse(time.RFC3339, regT[2]+"-"+regT[1]+"-"+regT[0]+"T23:59:59Z")
	return regToday
}

type register struct {
	ShowMosqueSelect bool
	MosqueSelected   bool
	DateSelected     bool
	PrayerSelected   bool
	MaxFutureDate    int
	Favourite        string
	Mosques          []model.Mosque
	Mosque           model.Mosque
	Prayer           []TempPrayer
	Date             string // Date formatted to 01-12-31
	DateString       string // Date formatted to 31.12.01
	PrayerName       string
	PrayerID         int
}

func RegisterPrayer(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		var reg register
		reg.Favourite = GetChoo(request).N
		mosque := request.PostFormValue("mosque")
		date := request.PostFormValue("date")
		prayer := request.PostFormValue("prayer")
		confirm := request.PostFormValue("confirm")
		if len(request.PostForm) == 0 { // start
			reg.Mosques = getMosques(response, request, false)
			reg.ShowMosqueSelect = true
			t, _ := template.ParseFiles("templates/registerPrayer.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, reg)
		} else if mosque != "" && date == "" && prayer == "" { // mosque selected
			reg.Mosque = getMosque(mosque)
			if reg.Mosque.Name != "" {
				reg.MosqueSelected = true
				reg.MaxFutureDate = reg.Mosque.MaxFutureDate
				t, _ := template.ParseFiles("templates/registerPrayer.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
				t.Execute(response, reg)
			}
		} else if mosque != "" && date != "" && prayer == "" {
			reg.Mosque = getMosque(mosque)
			index := 0
			if reg.Mosque.Name != "" {
				for i, dates := range reg.Mosque.Date {
					if date == strings.Split(dates.Date.String(), " ")[0] {
						index = i
						reg.Date = date
						reg.DateString = strconv.Itoa(dates.Date.Day()) + "." + strconv.Itoa(int(dates.Date.Month())) + "." + strconv.Itoa(dates.Date.Year())
						break
					}
				}
			}
			cap := 0
			user, err := GetUserAsUser(response, request)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError("Cerez hatasi | Cookiefehler", "/"))
				return
			}
			male := user.Sex == "Men"
			var prayers []TempPrayer
			pray := *new(TempPrayer)
			for _, prayer := range reg.Mosque.Date[index].Prayer {
				if prayer.Available {
					if male {
						cap = prayer.CapacityMen
					} else {
						cap = prayer.CapacityWomen
					}
					if cap != 0 {
						pray.Name = prayer.Name
						pray.Capacity = cap
						pray.Available = true
						prayers = append(prayers, pray)
					}
				}
			}
			reg.Prayer = prayers
			reg.DateSelected = true
			t, _ := template.ParseFiles("templates/registerPrayer.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, reg)
		} else if mosque != "" && date != "" && prayer != "" && confirm == "" {
			prayerI, err := strconv.Atoi(prayer)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError("Yanlis sayi boyutu | Falsches Zahlenformat", "/index"))
			}
			reg.Mosque = getMosque(mosque)
			if reg.Mosque.Name != "" {
				setPrayer := prayerI > 0 && prayerI < 8
				if setPrayer {
					switch prayerI {
					case 1:
						reg.PrayerName = "Sabah"
					case 2:
						reg.PrayerName = "Ögle"
					case 3:
						reg.PrayerName = "Ikindi"
					case 4:
						reg.PrayerName = "Aksam"
					case 5:
						reg.PrayerName = "Yatsi"
					case 6:
						reg.PrayerName = "Cuma"
					case 7:
						reg.PrayerName = "Bayram"
					}
					reg.PrayerID = prayerI
					reg.Date = date
					reg.DateString = request.PostFormValue("dateString")
					reg.PrayerSelected = true
				}
				t, _ := template.ParseFiles("templates/registerPrayer.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
				t.Execute(response, reg)
			}
		} else if mosque != "" && date != "" && prayer != "" && confirm == "yes" {
			prayerN := request.PostFormValue("prayerName")
			prayerI, err := strconv.Atoi(prayer)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError("Yanlis sayi boyutu | Falsches Zahlenformat", "/index"))
				return
			}
			collection, err := repos.GetDBCollection(0)
			if err != nil {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError(dbConnectionError, "/index"))
				return
			}
			reg.Mosque = getMosque(mosque)
			if reg.Mosque.Name != "" {
				choosenMosque := reg.Mosque
				user, err := GetUserAsUser(response, request)
				if err != nil {
					t, _ := template.ParseFiles("templates/errorpage.gohtml")
					t.Execute(response, GetError("Cerez hatasi | Cookiefehler", "/"))
					return
				}
				registered := model.RegisteredPrayer{}
				var mosque = getMosque(choosenMosque.Name)
				index := 0
				for i, dates := range choosenMosque.Date {
					if date == strings.Split(dates.Date.String(), " ")[0] {
						registered.Date = strconv.Itoa(dates.Date.Day()) + "." + strconv.Itoa(int(dates.Date.Month())) + "." + strconv.Itoa(dates.Date.Year())
						index = i
						break
					}
				}
				registered.RpId = mosque.Name + ":" + strconv.Itoa(index) + ":" + prayer
				result := collection.FindOne(context.TODO(), bson.D{
					{"Phone", user.Phone},
					{"RegisteredPrayers.RpId", registered.RpId}})
				if result.Err() != nil {
					registered.PrayerName = prayerN
					registered.PrayerIndex = prayerI
					registered.MosqueName = mosque.Name
					registered.MosqueAddress = strconv.Itoa(mosque.PLZ) + " " + mosque.City + ", " + mosque.Street
					registered.DateIndex = index
					collection, err = repos.GetDBCollection(1)
					if err != nil {
						t, _ := template.ParseFiles("templates/errorpage.gohtml")
						t.Execute(response, GetError(dbConnectionError, "/index"))
						return
					}
					collection.UpdateOne(context.TODO(),
						bson.M{"Name": mosque.Name},
						bson.D{{"$inc", bson.D{
							{"Date." + strconv.Itoa(index) + ".Prayer." + strconv.Itoa(prayerI-1) + ".Capacity" + user.Sex, -1},
						},
						}})
					tempUser := user
					tempUser.RegisteredPrayers = []model.RegisteredPrayer{}
					collection.UpdateOne(context.TODO(),
						bson.M{"Name": mosque.Name}, bson.M{"$push": bson.M{"Date." + strconv.Itoa(index) + ".Prayer." + strconv.Itoa(prayerI-1) + ".Users": tempUser}})
					user.RegisteredPrayers = append(user.RegisteredPrayers, registered)
					collection, err = repos.GetDBCollection(0)
					if err != nil {
						t, _ := template.ParseFiles("templates/errorpage.gohtml")
						t.Execute(response, GetError(dbConnectionError, "/index"))
						return
					}
					phone, err := GetPhoneFromCookie(request)
					if err != nil {
						t, _ := template.ParseFiles("templates/errorpage.gohtml")
						t.Execute(response, GetError(err.Error(), "/"))
						return
					}
					encP := repos.Encrypt(phone)
					collection.UpdateOne(context.TODO(),
						bson.M{"Phone": encP}, bson.M{
							"$push": bson.M{"RegisteredPrayers": registered}})
					http.Redirect(response, request, "/index?success", 302)
				} else {
					http.Redirect(response, request, "/index?existent", 302)
				}
			} else {
				t, _ := template.ParseFiles("templates/errorpage.gohtml")
				t.Execute(response, GetError("Camii secilmedi | Keine Moschee asugewählt", "/index"))
			}
		} else {
			response.Write([]byte(`<script>window.location.href = "/index";</script>`))
		}
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func SignOutPrayer(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		name := request.FormValue("name")
		date := request.FormValue("date")
		prayer := request.FormValue("prayer")
		phone := request.FormValue("phone")
		prayerN, err := strconv.Atoi(prayer)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError("Bir hata olustu, birdaha deneyin | Ein Fehler ist aufgetreten, versuchen Sie es erneut", "/index"))
			return
		}
		prayer1 := strconv.Itoa(prayerN - 1)
		collection, err := repos.GetDBCollection(1)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/index"))
			return
		}
		encP := repos.Encrypt(phone)

		mosque := getMosque(name)
		dateIndex, err := strconv.Atoi(date)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError("Bir hata olustu, birdaha deneyin | Ein Fehler ist aufgetreten, versuchen Sie es erneut", "/index"))
			return
		}
		dateL := len(mosque.Date)
		dateP := len(mosque.Date[0].Prayer)
		if dateIndex > 0 && prayerN > 0 && dateIndex < dateL && prayerN-1 < dateP {
			users := mosque.Date[dateIndex].Prayer[prayerN-1].Users
			for _, user := range users {
				if user.Phone == encP && !user.Attended {
					collection.UpdateOne(context.TODO(),
						bson.M{"Name": name},
						bson.M{"$pull": bson.M{"Date" + "." + date + ".Prayer." + prayer1 + ".Users": bson.M{"Phone": encP}}})
					break
				}
			}
		} else {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError("Bir hata olustu, birdaha deneyin | Ein Fehler ist aufgetreten, versuchen Sie es erneut", "/index"))
			return
		}
		user, err := GetUserAsUser(response, request)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError("Cerez hatasi | Cookiefehler", "/"))
			return
		}
		collection.UpdateOne(context.TODO(),
			bson.M{"Name": name},
			bson.D{{"$inc", bson.D{
				{"Date." + date + ".Prayer." + prayer1 + ".Capacity" + user.Sex, 1},
			},
			}})
		collection, err = repos.GetDBCollection(0)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/index"))
			return
		}
		rpid := name + ":" + date + ":" + prayer
		filter := bson.D{{Key: "Phone", Value: encP}}
		update := bson.D{{Key: "$pull", Value: bson.D{{Key: "RegisteredPrayers", Value: bson.D{{Key: "RpId", Value: rpid}}}}}}
		collection.UpdateOne(context.TODO(), filter, update)
		http.Redirect(response, request, "/index", 302)
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func DeleteUser(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		user, err := GetUserAsUser(response, request)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(err.Error(), "/"))
			return
		}
		phone := user.Phone
		collection, err := repos.GetDBCollection(1)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/index"))
			return
		}
		for _, regP := range user.RegisteredPrayers {
			name := regP.MosqueName
			date := strconv.Itoa(regP.DateIndex)
			prayer := strconv.Itoa(regP.PrayerIndex - 1)
			collection.UpdateOne(context.TODO(),
				bson.M{"Name": name},
				bson.M{"$pull": bson.M{"Date" + "." + date + ".Prayer." + prayer + ".Users": bson.M{"Phone": phone}}})
			collection.UpdateOne(context.TODO(),
				bson.M{"Name": name},
				bson.M{"$inc": bson.M{"Date." + date + ".Prayer." + prayer + ".Capacity" + user.Sex: 1}})
		}
		collection, err = repos.GetDBCollection(0)
		if err != nil {
			t, _ := template.ParseFiles("templates/errorpage.gohtml")
			t.Execute(response, GetError(dbConnectionError, "/index"))
			return
		}
		collection.DeleteOne(context.TODO(), bson.M{"Phone": user.Phone})
		response.Write([]byte(`<script>window.location.href = "/?deleted";</script>`))
	} else {
		t, _ := template.ParseFiles("templates/errorpage.gohtml")
		t.Execute(response, GetError("Kayidiniz gecerli degil | Anmeldung nicht gültig", "/"))
	}
}

func LogoutHandler(response http.ResponseWriter, request *http.Request) {
	ClearCookie(response)
	ClearMosque(response)
	http.Redirect(response, request, "/", 302)
}

func SetChoo(choosenMosque chooseCookie, response http.ResponseWriter) {
	choose, err := json.Marshal(choosenMosque)
	if err != nil {
		panic(err)
	}
	cookie := &http.Cookie{
		Name:  "choosenMosque",
		Value: repos.Encode(choose),
		Path:  "/",
	}
	http.SetCookie(response, cookie)
}

func GetChoo(request *http.Request) chooseCookie {
	var choosenOne chooseCookie
	cookie, err := request.Cookie("choosenMosque")
	if err != nil {
		return choosenOne
	}
	json.Unmarshal([]byte(repos.Decode(cookie.Value)), &choosenOne)
	return choosenOne
}

// Function for setting the Cookie
func SetCookie(user string, response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:  "cookie",
		Value: user,
		Path:  "/",
	}
	http.SetCookie(response, cookie)
}

func ClearMosque(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "choosenMosque",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

// Function for deletion of the Cookie
func ClearCookie(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "cookie",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

// If user phone, if admin username, if mosqueadmin mosquename
func GetPhoneFromCookie(request *http.Request) (string, error) {
	phone := ""
	// Check if there is an active Cookie
	cookie, err := request.Cookie("cookie")
	if err != nil {
		return "", errors.New("Giris yapilmadi | Nicht eingeloggt")
	}
	if !strings.Contains(cookie.Value, "!") {
		return "", errors.New("Cerez hatasi | Cookiefehler")
	}
	if !strings.Contains(cookie.Value, "&") {
		return "", errors.New("Cerez hatasi | Cookiefehler")
	}
	if !strings.Contains(cookie.Value, "?") {
		return "", errors.New("Cerez hatasi | Cookiefehler")
	}
	cookieValue := cookie.Value
	cookieVal := strings.Split(cookieValue, "&")[0]
	values := strings.Split(cookieVal, "?")
	cookieHash := strings.Split(cookieValue, "&")[1]
	err = bcrypt.CompareHashAndPassword([]byte(cookieHash), []byte(values[0]+values[1]))
	if err != nil {
		return "", errors.New("Cerez hatasi | Cookiefehler")
	}
	val := strings.Split(values[1], "!")[0]
	phone = val
	return phone, nil
}

func R(input string) string {
	replacer := strings.NewReplacer("ä", "ae", "ö", "oe", "ü", "ue", "ß", "ss")
	return replacer.Replace(input)
}

func GetUserAsUser(response http.ResponseWriter, request *http.Request) (model.User, error) {
	var user model.User
	phone, err := GetPhoneFromCookie(request)
	if err != nil {
		return user, err
	}
	collection, err := repos.GetDBCollection(0)
	if err != nil {
		return user, errors.New(dbConnectionError)
	}
	encP := repos.Encrypt(phone)
	collection.FindOne(context.TODO(), bson.M{"Phone": encP}).Decode(&user)
	return user, nil
}

func getMosques(response http.ResponseWriter, request *http.Request, all bool) mosques {
	mosquess := []model.Mosque{}
	collection, err := repos.GetDBCollection(1)
	if err != nil {
		log.Println(dbConnectionError)
		return mosquess
	}
	cur, err := collection.Find(context.TODO(), bson.D{})
	for cur.Next(context.TODO()) {
		var mosque model.Mosque
		cur.Decode(&mosque)
		if all || mosque.Active {
			mosquess = append(mosquess, mosque)
		}
	}
	return mosquess
}

func getMosque(name string) model.Mosque {
	var mosque model.Mosque
	collection, _ := repos.GetDBCollection(1)
	collection.FindOne(context.TODO(),
		bson.D{
			{"Name", name},
		}).Decode(&mosque)
	return mosque
}

func GetError(err string, link string) ErrorMessage {
	return ErrorMessage{err, link}
}

func adminLoggedin(response http.ResponseWriter, request *http.Request, adminType string) bool {
	cookie, err := request.Cookie("cookie")
	if err != nil {
		return false
	}
	values := cookie.Value
	if !strings.Contains(values, "!") {
		return false
	}
	if !strings.Contains(values, "&") {
		return false
	}
	if !strings.Contains(values, "?") {
		return false
	}
	if strings.Split(values, "!")[1] == adminType {
		cookieValue := strings.Split(R(cookie.Value), "!")[0]
		cookieVal := strings.Split(cookieValue, "&")[0]
		valuesCookie := strings.Split(cookieVal, "?")
		cookieHash := strings.Split(cookieValue, "&")[1]
		err = bcrypt.CompareHashAndPassword([]byte(cookieHash), []byte(valuesCookie[0]+valuesCookie[1]))
		if err != nil {
			return false
		}
		return true
	}
	return false
}

// check every method with this
func loggedin(response http.ResponseWriter, request *http.Request) bool {
	if _, err := GetPhoneFromCookie(request); err != nil {
		return false
	}
	return true
}
