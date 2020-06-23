package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"pi-software/helpers"
	"pi-software/model"
	"pi-software/repos"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

const (
	dbConnectionError = "Error connecting to Database"
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

type chooseCookie struct {
	N string //Name
	S bool   //SetPrayer
	D int    //Date as int from today for time.Now add D
	P int    //Prayer
}

func chooseToCookie(choo choose) chooseCookie {
	var cookie chooseCookie
	cookie.N = choo.Name
	cookie.S = choo.SetPrayer
	for i := 0; i < choo.MaxFutureDate; i++ {
		if strings.Split(choo.Date.Date.String(), " ")[0] == strings.Split(time.Now().AddDate(0, 0, i).String(), " ")[0] {
			cookie.D = i
			break
		}
	}
	switch choo.PrayerName {
	case "Sabah":
		cookie.P = 1
	case "Ögle":
		cookie.P = 2
	case "Ikindi":
		cookie.P = 3
	case "Aksam":
		cookie.P = 4
	case "Yatsi":
		cookie.P = 5
	case "Cuma":
		cookie.P = 6
	case "Bayram":
		cookie.P = 7
	}
	return cookie
}

func cookieToChoose(cookie chooseCookie) choose {
	var choo choose
	choo.Name = cookie.N
	collection, _ := repos.GetDBCollection(1)
	var mosque model.Mosque
	collection.FindOne(context.TODO(), bson.M{"Name": cookie.N}).Decode(&mosque)
	choo.City = mosque.City
	choo.Street = mosque.Street
	choo.PLZ = mosque.PLZ
	choo.SetPrayer = cookie.S
	choo.MaxFutureDate = mosque.MaxFutureDate
	for _, dates := range mosque.Date {
		if strings.Split(dates.Date.String(), " ")[0] == strings.Split(time.Now().AddDate(0, 0, cookie.D).String(), " ")[0] {
			choo.Date = dates
			break
		}
	}
	choo.DateString = choo.Date.Date.Format(time.RFC3339)
	switch cookie.P {
	case 1:
		choo.PrayerName = "Sabah"
	case 2:
		choo.PrayerName = "Ögle"
	case 3:
		choo.PrayerName = "Ikindi"
	case 4:
		choo.PrayerName = "Aksam"
	case 5:
		choo.PrayerName = "Yatsi"
	case 6:
		choo.PrayerName = "Cuma"
	case 7:
		choo.PrayerName = "Bayram"
	}
	return choo
}

// Struct for easy handling the html template generation on the index page
type TempPrayer struct {
	Name      model.PrayerName
	Capacity  int
	Available bool
}

var emptyChoose = choose{"", "", "", 0, false, *new([]model.Mosque), *new(model.Date), "", *new([]TempPrayer), "", 0}

var emptyCookie = chooseCookie{"", false, 0, 0}

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
	t := check(response, request, err)
	if t != nil {
		t.Execute(response, errors.New(dbConnectionError))
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
		err := collection.FindOne(context.TODO(), bson.D{{"Phone", phone}})
		// If not found (throws exception/error) then we can proceed
		if err != nil {
			// Generate the hashed password with 14 as salt

			//use hash if you want
			//hash, err := bcrypt.GenerateFromPassword([]byte(phone), 14)

			// If there was an error generating the hash dont proceed

			// define a User model with typed first and last name, email and phone

			//usr := model.User{firstName, email, string(hash)}
			encF := repos.Encrypt(firstName)
			if encF == "" {
				fmt.Println("encF")
			}
			encL := repos.Encrypt(lastName)
			if encL == "" {
				fmt.Println("encL")
			}
			encE := repos.Encrypt(email)
			if encE == "" {
				fmt.Println("encE")
			}
			encP := repos.Encrypt(phone)
			if encP == "" {
				fmt.Println("encP")
			}
			usr := model.User{sex, encF, encL, encE, encP, false, []model.RegisteredPrayer{}}
			// Insert user to the table
			collection.InsertOne(context.TODO(), usr)
			// Change redirect target to LoginPage
			http.Redirect(response, request, "/", 302)
		} else {
			// TODO: checkError
			fmt.Fprintln(response, "User already exists")
		}
	} else {
		// TODO: checkError
		fmt.Fprintln(response, "This fields can not be blank!")
	}
}

// Handler for Login Page used with POST by submitting Loginform
func LoginHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	mosque := request.URL.Query().Get("mosque")
	if mosque != "" {
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
		redirectTarget := "/login"
		if len(email) != 0 && len(phone) != 0 {
			collection, err := repos.GetDBCollection(0)
			if err != nil {
				http.Redirect(response, request, "/register", 302)
			}
			var user model.User

			encP := repos.Encrypt(phone)

			err = collection.FindOne(context.TODO(), bson.D{{"Phone", encP}}).Decode(&user)
			if err != nil {
				fmt.Println(err, "***", encP, phone)
				http.Redirect(response, request, "/register", 302)
			}
			encE := repos.Encrypt(email)
			if user.Email != encE {
				fmt.Println("email: ***", encE, user.Email)
				http.Redirect(response, request, "/register", 302)
			}
			userCredentials, err := bcrypt.GenerateFromPassword([]byte(R(email+phone)), 14)
			if err != nil {
				fmt.Println(err.Error())
			}
			cookie := R(email+"?"+phone+"&"+string(userCredentials)) + "!"
			SetCookie(cookie, response)
			redirectTarget = "/"
		}
		// function for redirecting
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
		redirectTarget := "/login"
		if len(email) != 0 && len(password) != 0 {
			// Returns Table
			collection, err := repos.GetDBCollection(2)

			// if there was no error getting the table, te program does these operations
			if err != nil {
				http.Redirect(response, request, "/register", 302)
			}
			var admin model.Admin
			// Checking if typed in Username exists, if not redirect to register page
			encE := repos.Encrypt(email)
			err = collection.FindOne(context.TODO(), bson.D{{Key: "Email", Value: encE}}).Decode(&admin)
			// If there was an error getting an entry with matching username (no user with this username) redirect to faultpage
			if err != nil {
				http.Redirect(response, request, "/register", 302)
			}

			err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
			if err != nil {
				http.Redirect(response, request, "/register", 302)
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
		// function for redirecting
		http.Redirect(response, request, redirectTarget, 302)
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
		SetChoo(emptyChoose, response)
		if err != nil {
			response.Write([]byte(`<script>window.location.href = "/login";</script>`))
		} else {
			user := decryptUser(encryptedUser)
			var tUser model.User // for only adding the actual prayers, not past ones
			tUser = user
			var regP []model.RegisteredPrayer
			tUser.RegisteredPrayers = regP
			var mosque model.Mosque
			collection, _ := repos.GetDBCollection(1)
			for _, reg := range user.RegisteredPrayers {
				collection.FindOne(context.TODO(), bson.M{"Name": reg.MosqueName}).Decode(&mosque)
				if mosque.Active { // only show registrations if mosque is active
					if reg.Date != "" { // delöte this later
						regT := strings.Split(reg.Date, ".")
						m, _ := strconv.Atoi(regT[1])
						d, _ := strconv.Atoi(regT[0])
						regT[1] = fmt.Sprintf("%02d", m)
						regT[0] = fmt.Sprintf("%02d", d)
						regToday, _ := time.Parse(time.RFC3339, regT[2]+"-"+regT[1]+"-"+regT[0]+"T23:59:59Z")
						timeNow := time.Now()
						if regToday.After(timeNow) {
							regP = append(regP, reg)
						}
					}
				}
			}
			tUser.RegisteredPrayers = regP
			t, _ := template.ParseFiles("templates/index.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, tUser)
		}
	} else {
		response.Write([]byte(`<script>window.location.href = "/login";</script>`))
	}
}

func LogoutHandler(response http.ResponseWriter, request *http.Request) {
	ClearCookie(response)
	http.Redirect(response, request, "/", 302)
}

func Choose(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		mosque := request.URL.Query().Get("mosque")
		var choo = emptyChoose
		if mosque == "" {
			mosquesList = getMosques(response, request, false)
			choo.Mosques = mosquesList
			t, _ := template.ParseFiles("templates/choose.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
			t.Execute(response, choo)
			choo.Mosques = *new([]model.Mosque)
			SetChoo(choo, response)
		} else {
			choo.Name = mosque
			if len(mosquesList) > 0 && choo.Name != "" {
				for _, mosq := range mosquesList {
					if mosq.Name == mosque {
						choo.MaxFutureDate = mosq.MaxFutureDate
						choo.Street = mosq.Street
						choo.PLZ = mosq.PLZ
						choo.City = mosq.City
						t, _ := template.ParseFiles("templates/chooseDate.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
						SetChoo(choo, response)
						t.Execute(response, getMosque(mosque))
						return
					}
				}
			} else {
				SetChoo(emptyChoose, response)
				response.Write([]byte(`<script>window.location.href = "/";</script>`))
			}
		}
	} else {
		http.Redirect(response, request, "/login", 401)
		response.Write([]byte(`<script>window.location.href = "/login";</script>`))
	}
}

func Choosen(response http.ResponseWriter, request *http.Request) {
	mosque := request.URL.Query().Get("mosque")
	t, _ := template.ParseFiles("templates/chooseDate.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
	t.Execute(response, getMosque(mosque))
}

func ChooseDate(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		date := request.PostFormValue("date")
		choosenMosque := getMosque(request.PostFormValue("mosque"))
		var choo = GetChoo(request)
		index := 0
		for i, dates := range choosenMosque.Date {
			if date == strings.Split(dates.Date.String(), " ")[0] {
				index = i
				choo.Date.Date = dates.Date
				break
			}
		}
		cap := 0
		user, err := GetUserAsUser(response, request)
		if err != nil {
			http.Redirect(response, request, "/login", 302)
			SetChoo(emptyChoose, response)
			return
		}
		male := user.Sex == "Men"
		for _, prayer := range choosenMosque.Date[index].Prayer {
			if prayer.Available {
				if male {
					cap = prayer.CapacityMen
				} else {
					cap = prayer.CapacityWomen
				}
				if cap != 0 {
					pray := *new(TempPrayer)
					pray.Name = prayer.Name
					pray.Capacity = cap
					pray.Available = true
					choo.Prayer = append(choo.Prayer, pray)
				}
			}
		}
		t, _ := template.ParseFiles("templates/choosePrayer.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
		SetChoo(choo, response)
		t.Execute(response, choo)
	} else {
		http.Redirect(response, request, "/login", 401)
		response.Write([]byte(`<script>window.location.href = "/login";</script>`))
	}
}

func ChoosePrayer(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		prayer, _ := strconv.Atoi(request.URL.Query().Get("prayer"))
		var choo = GetChoo(request)
		choo.SetPrayer = prayer > 0 && prayer < 8
		if choo.SetPrayer {
			switch prayer {
			case 1:
				choo.PrayerName = "Sabah"
			case 2:
				choo.PrayerName = "Ögle"
			case 3:
				choo.PrayerName = "Ikindi"
			case 4:
				choo.PrayerName = "Aksam"
			case 5:
				choo.PrayerName = "Yatsi"
			case 6:
				choo.PrayerName = "Cuma"
			case 7:
				choo.PrayerName = "Bayram"
			}
			date := choo.Date.Date
			choo.DateString = strconv.Itoa(date.Day()) + "." + strconv.Itoa(int(date.Month())) + "." + strconv.Itoa(date.Year())
			t, _ := template.ParseFiles("templates/confirm.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
			SetChoo(choo, response)
			t.Execute(response, choo)
		} else {
			http.Error(response, "no valid prayer selected", 402)
		}
	} else {
		http.Redirect(response, request, "/login", 401)
		response.Write([]byte(`<script>window.location.href = "/login";</script>`))
	}
}

func SubmitPrayer(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		if request.URL.Query().Get("confirm") == "yes" {
			collection, _ := repos.GetDBCollection(1)
			var choo = GetChoo(request)
			choosenMosque := getMosque(choo.Name)
			prayer := 0
			switch choo.PrayerName {
			case "Sabah":
				prayer = 1
			case "Ögle":
				prayer = 2
			case "Ikindi":
				prayer = 3
			case "Aksam":
				prayer = 4
			case "Yatsi":
				prayer = 5
			case "Cuma":
				prayer = 6
			case "Bayram":
				prayer = 7
			}
			user, err := GetUserAsUser(response, request)
			if err != nil {
				http.Redirect(response, request, "/login", 302)
				return
			}
			registered := model.RegisteredPrayer{}
			var mosque model.Mosque
			collection.FindOne(context.TODO(),
				bson.D{
					{"Name", choosenMosque.Name},
				}).Decode(&mosque)
			registered.PrayerName = choo.PrayerName
			registered.PrayerIndex = prayer
			registered.MosqueName = mosque.Name
			registered.MosqueAddress = strconv.Itoa(mosque.PLZ) + " " + mosque.City + ", " + mosque.Street
			index := 0
			for i, dates := range choosenMosque.Date {
				if choo.Date.Date == dates.Date {
					registered.Date = strconv.Itoa(dates.Date.Day()) + "." + strconv.Itoa(int(dates.Date.Month())) + "." + strconv.Itoa(dates.Date.Year())
					index = i
					break
				}
			}
			collection, _ = repos.GetDBCollection(0)
			registered.RpId = mosque.Name + ":" + strconv.Itoa(index) + ":" + strconv.Itoa(prayer)
			encP := repos.Encrypt(user.Phone)
			result := collection.FindOne(context.TODO(), bson.D{
				{"Phone", encP},
				{"RegisteredPrayers.RpId", registered.RpId}})
			if result.Err() != nil {
				collection, _ = repos.GetDBCollection(1)
				registered.DateIndex = index
				_, error := collection.UpdateOne(context.TODO(),
					bson.M{"Name": mosque.Name},
					bson.D{{"$inc", bson.D{
						{"Date." + strconv.Itoa(index) + ".Prayer." + strconv.Itoa(prayer-1) + ".Capacity" + user.Sex, -1},
					},
					}})
				if error != nil {
					http.Redirect(response, request, "/404", 302)
				}
				tempUser := user
				tempUser.RegisteredPrayers = []model.RegisteredPrayer{}
				collection.UpdateOne(context.TODO(),
					bson.M{"Name": mosque.Name}, bson.M{"$push": bson.M{"Date." + strconv.Itoa(index) + ".Prayer." + strconv.Itoa(prayer-1) + ".Users": tempUser}})
				user.RegisteredPrayers = append(user.RegisteredPrayers, registered)
				collection, err = repos.GetDBCollection(0)
				t := check(response, request, err)
				if t != nil {
					t.Execute(response, errors.New(dbConnectionError))
					return
				}
				phone, err := GetPhoneFromCookie(request)
				t = check(response, request, err)
				if t != nil {
					t.Execute(response, errors.New(err.Error()))
					return
				}
				encP := repos.Encrypt(phone)
				collection.UpdateOne(context.TODO(),
					bson.M{"Phone": encP}, bson.M{
						"$push": bson.M{"RegisteredPrayers": registered}})
				SetChoo(emptyChoose, response)
				http.Redirect(response, request, "/", 302)
			} else {
				t, _ := template.ParseFiles("templates/errorpage.gohtml", "templates/base_loggedin.tmpl", "templates/footer.tmpl")
				t.Execute(response, errors.New("Bu namaz icin gecerli bir kayidiniz bulunmakta! Sie besitzen bereits eine gültige Anmeldung für dieses Gebet"))
			}
		} else {
			SetChoo(emptyChoose, response)
			http.Redirect(response, request, "/", 302)
		}
	} else {
		http.Redirect(response, request, "/login", 401)
		response.Write([]byte(`<script>window.location.href = "/login";</script>`))
	}
}

func SignOutPrayer(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		name := request.FormValue("name")
		date := request.FormValue("date")
		prayer := request.FormValue("prayer")
		phone := request.FormValue("phone")
		prayerN, err := strconv.Atoi(prayer)
		t := check(response, request, err)
		if t != nil {
			t.Execute(response, errors.New("Wrong Input format"))
			return
		}
		prayer1 := strconv.Itoa(prayerN - 1)
		collection, _ := repos.GetDBCollection(1)
		encP := repos.Encrypt(phone)
		collection.UpdateOne(context.TODO(),
			bson.M{"Name": name},
			bson.M{"$pull": bson.M{"Date" + "." + date + ".Prayer." + prayer1 + ".Users": bson.M{"Phone": encP}}})
		user, err := GetUserAsUser(response, request)
		if err != nil {
			http.Redirect(response, request, "/login", 302)
			return
		}
		collection.UpdateOne(context.TODO(),
			bson.M{"Name": name},
			bson.D{{"$inc", bson.D{
				{"Date." + date + ".Prayer." + prayer1 + ".Capacity" + user.Sex, 1},
			},
			}})
		collection, _ = repos.GetDBCollection(0)
		rpid := name + ":" + date + ":" + prayer
		filter := bson.D{{Key: "Phone", Value: encP}}
		update := bson.D{{Key: "$pull", Value: bson.D{{Key: "RegisteredPrayers", Value: bson.D{{Key: "RpId", Value: rpid}}}}}}
		collection.UpdateOne(context.TODO(), filter, update)
		http.Redirect(response, request, "/", 302)
	} else {
		http.Redirect(response, request, "/login", 401)
		response.Write([]byte(`<script>window.location.href = "/login";</script>`))
	}
}

func SetChoo(choosenMosque choose, response http.ResponseWriter) {
	choose, err := json.Marshal(chooseToCookie(choosenMosque))
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

func GetChoo(request *http.Request) choose {
	var choosenOne chooseCookie
	cookie, err := request.Cookie("choosenMosque")
	if err != nil {
		panic(err)
	}
	json.Unmarshal([]byte(repos.Decode(cookie.Value)), &choosenOne)
	return cookieToChoose(choosenOne)
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
		Name:   "mosque",
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

// TODO: Filter braces, sql commands etc.
func CheckInput(input string) bool {
	if match, _ := regexp.MatchString(`\w+`, input); match == true {
		return true
	}
	return false
}

func GetPhoneFromCookie(request *http.Request) (string, error) {
	phone := ""
	// Check if there is an active Cookie
	cookie, err := request.Cookie("cookie")
	if err != nil {
		return "", errors.New("Not logged in")
	}
	if !strings.Contains(cookie.Value, "!") {
		return "", errors.New("Invalid Cookie")
	}
	if !strings.Contains(cookie.Value, "&") {
		return "", errors.New("Invalid Cookie")
	}
	if !strings.Contains(cookie.Value, "?") {
		return "", errors.New("Invalid Cookie")
	}
	cookieValue := strings.Split(R(cookie.Value), "!")[0]
	cookieVal := strings.Split(cookieValue, "&")[0]
	values := strings.Split(cookieVal, "?")
	cookieHash := strings.Split(cookieValue, "&")[1]
	err = bcrypt.CompareHashAndPassword([]byte(cookieHash), []byte(values[0]+values[1]))
	if err != nil {
		return "", errors.New("Wrong or Modified Cookie")
	}
	phone = values[1]
	return phone, nil
}

func R(input string) string {
	replacer := strings.NewReplacer("ä", "ae", "ö", "oe", "ü", "ue", "ß", "ss")
	return replacer.Replace(input)
}

func GetUserAsUser(response http.ResponseWriter, request *http.Request) (model.User, error) {
	var user model.User
	phone, err := GetPhoneFromCookie(request)
	t := check(response, request, err)
	if t != nil {
		return user, errors.New(err.Error())
	}
	collection, err := repos.GetDBCollection(0)
	t = check(response, request, err)
	if t != nil {
		return user, errors.New(dbConnectionError)
	}
	encP := repos.Encrypt(phone)
	err = collection.FindOne(context.TODO(), bson.M{"Phone": encP}).Decode(&user)
	return user, err
}

func getMosques(response http.ResponseWriter, request *http.Request, all bool) mosques {
	mosquess := []model.Mosque{}
	collection, err := repos.GetDBCollection(1)
	t := check(response, request, err)
	if t != nil {
		t.Execute(response, errors.New(dbConnectionError))
		return nil
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

func getUsers(response http.ResponseWriter, request *http.Request) users {
	var users []model.User
	dataBase, err := repos.GetDBCollection(0)
	t := check(response, request, err)
	if t != nil {
		t.Execute(response, errors.New(dbConnectionError))
		return nil
	}
	cur, _ := dataBase.Find(context.TODO(), bson.D{})
	for cur.Next(context.TODO()) {
		var user model.User
		cur.Decode(&user)
		users = append(users, decryptUser(user))
	}
	return users
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

func check(response http.ResponseWriter, request *http.Request, err error) *template.Template {
	if err != nil {
		t, _ := template.ParseFiles("templates/errorpage.html")
		http.Redirect(response, request, "/error", 402)
		response.Write([]byte(`<script>window.location.href = "/error"</script>`))
		SetChoo(emptyChoose, response)
		return t
	}
	return nil
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
