package common

import (
	"context"
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
	Name       string
	City       string
	SetMosque  bool
	SetDate    bool
	SetPrayer  bool
	Mosques    []model.Mosque
	Date       model.Date
	DateString string
	Prayer     []TempPrayer
	PrayerName string
}

// Struct for easy handling the html template generation on the index page
type TempPrayer struct {
	Name      model.PrayerName
	Capacity  int
	Available bool
}

var isAdmin bool

var choo choose

var choosenMosque model.Mosque

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

// Function for handling the Webpage call with GET
func RegisterPageHandler(response http.ResponseWriter, request *http.Request) {
	var body, _ = helpers.LoadFile("templates/register.html")
	fmt.Fprintf(response, body)
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

			usr := model.User{sex, firstName, lastName, email, phone, false, []model.RegisteredPrayer{}}
			// Insert user to the table
			collection.InsertOne(context.TODO(), usr)
			// Change redirect target to LoginPage
			http.Redirect(response, request, "/", 302)
		} else {
			fmt.Fprintln(response, "User already exists")
		}
	} else {
		fmt.Fprintln(response, "This fields can not be blank!")
	}
}

// Handler for Login Page used with POST by submitting Loginform
func LoginHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	if len(request.PostForm) > 0 {
		if request.FormValue("type") == "admin" {
			adminLogin(response, request)
			return
		}
		email := request.FormValue("email")
		phone := request.FormValue("phone")
		// Default redirect page is the login page, so if anything goes wrong, the program just redirects to the login page again
		redirectTarget := "/"
		if len(email) != 0 && len(phone) != 0 {
			// Returns Table
			collection, err := repos.GetDBCollection(0)

			// if there was no error getting the table, te program does these operations
			if err != nil {
				http.Redirect(response, request, "/register", 302)
			}
			var user model.User
			// Checking if typed in Username exists, if not redirect to register page
			err = collection.FindOne(context.TODO(), bson.D{{"Phone", phone}}).Decode(&user)
			// If there was an error getting an entry with matching username (no user with this username) redirect to faultpage
			if err != nil {
				http.Redirect(response, request, "/register", 302)
			}
			// Checking if typed in password is equivalent to the password typed in registry process, if not redirect to faultpage
			/* Use encryption if you want
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass

			if check(err) {
				http.Redirect(response, request, "/register", 302)
			}
			*/

			if user.Email != email {
				http.Redirect(response, request, "/register", 302)
			}

			userCredentials, err := bcrypt.GenerateFromPassword([]byte(R(email+phone)), 14)
			cookie := R(email+"?"+phone+"&"+string(userCredentials)) + "!"
			SetCookie(cookie, response)
			// If the admin tries to login, change the redirect to the Adminpage
			if email == "steveJobs@apple.de" {
				redirectTarget = "/appleHeadquarter"
				// Else redirect to the normal indexpage
			} else {
				redirectTarget = "/"
			}
		}
		// function for redirecting
		http.Redirect(response, request, redirectTarget, 302)
	} else {
		t, _ := template.ParseFiles("templates/login.html")
		t.Execute(response, nil)
	}
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
			err = collection.FindOne(context.TODO(), bson.D{{Key: "Email", Value: email}}).Decode(&admin)
			// If there was an error getting an entry with matching username (no user with this username) redirect to faultpage
			if err != nil {
				http.Redirect(response, request, "/register", 302)
			}

			err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))

			if err != nil {
				http.Redirect(response, request, "/register", 302)
			}

			if admin.Email != email {
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
			t, _ := template.ParseFiles("templates/login.html")
			t.Execute(response, nil)
		}
		// function for redirecting
		http.Redirect(response, request, redirectTarget, 302)
	} else {
		t, _ := template.ParseFiles("templates/login.html")
		t.Execute(response, nil)
	}
}

func IndexPageHandler(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		user, err := GetUserAsUser(response, request)
		if err != nil {
			http.Redirect(response, request, "/login", 302)
			reset()
		} else {
			t, _ := template.ParseFiles("templates/index.html")
			t.Execute(response, user)
		}
	} else {
		http.Redirect(response, request, "/login", 401)
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
		if mosque == "" {
			choo.Mosques = getMosques(response, request)
			t, _ := template.ParseFiles("templates/choose.html")
			t.Execute(response, choo)
			choo.SetMosque = true
		} else {
			choo.Name = mosque
			if len(choo.Mosques) > 0 && choo.Name != "" {
				for _, mosq := range choo.Mosques {
					if mosq.Name == mosque {
						choosenMosque = mosq
						t, _ := template.ParseFiles("templates/chooseDate.html")
						t.Execute(response, choo)
						return
					}
				}
			} else {
				reset()
				Choose(response, request)
			}
		}
	} else {
		http.Redirect(response, request, "/login", 401)
		response.Write([]byte(`<script>window.location.href = "/login";</script>`))
	}
}

func Choosen(response http.ResponseWriter, request *http.Request) {
	mosque := request.URL.Query().Get("mosque")
	collection, _ := repos.GetDBCollection(1)
	collection.FindOne(context.TODO(),
		bson.D{
			{"Name", mosque},
		}).Decode(&choosenMosque)
	t, _ := template.ParseFiles("templates/chooseDate.html")
	t.Execute(response, choosenMosque)
	//http.Redirect(response, request, "/chooseDate", 302)
}

func getMosques(response http.ResponseWriter, request *http.Request) mosques {
	mosquess := []model.Mosque{}
	dataBase, err := repos.GetDBCollection(1)
	t := check(response, request, err)
	if t != nil {
		t.Execute(response, errors.New(dbConnectionError))
		return nil
	}
	cur, _ := dataBase.Find(context.TODO(), bson.D{})
	for cur.Next(context.TODO()) {
		var mosque model.Mosque
		cur.Decode(&mosque)
		if mosque.Active {
			mosquess = append(mosquess, mosque)
		}
	}
	return mosquess
}

func getUsers(response http.ResponseWriter, request *http.Request) users {
	users := []model.User{}
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
		users = append(users, user)
	}
	return users
}

func ChooseDate(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		date := request.PostFormValue("date")
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
			reset()
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
				pray := *new(TempPrayer)
				pray.Name = prayer.Name
				pray.Capacity = cap
				pray.Available = true
				choo.Prayer = append(choo.Prayer, pray)
			}
		}
		t, _ := template.ParseFiles("templates/choosePrayer.html")
		t.Execute(response, choo)
		choo.SetDate = true
	} else {
		http.Redirect(response, request, "/login", 401)
		response.Write([]byte(`<script>window.location.href = "/login";</script>`))
	}
}

// TODO: test invalid prayer as query
func ChoosePrayer(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		prayer, _ := strconv.Atoi(request.URL.Query().Get("prayer"))
		choo.SetPrayer = prayer != 0
		for _, prayers := range choo.Prayer {
			if int(prayers.Name) == prayer {
				choo.SetPrayer = true
			}
		}
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
		}
		choo.DateString = choo.Date.Date.Format(time.RFC3339)
		t, _ := template.ParseFiles("templates/confirm.html")
		t.Execute(response, choo)
	} else {
		http.Redirect(response, request, "/login", 401)
		response.Write([]byte(`<script>window.location.href = "/login";</script>`))
	}
}

func SubmitPrayer(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		if request.URL.Query().Get("confirm") == "yes" {
			collection, _ := repos.GetDBCollection(1)
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
			err1 := collection.FindOne(context.TODO(), bson.D{
				{"Phone", user.Phone},
				{"RegisteredPrayers.RpId", registered.RpId}})
			if err1 != nil {
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
				collection.UpdateOne(context.TODO(),
					bson.M{"Phone": phone}, bson.M{
						"$push": bson.M{"RegisteredPrayers": registered}})
				choo = *new(choose)
				http.Redirect(response, request, "/", 302)
			} else {
				t, _ := template.ParseFiles("templates/errorpage.html")
				err = errors.New("Bu namaz icin gecerli bir kayidiniz bulunmakta! Sie besitzen bereits eine gültige Anmeldung für dieses Gebet")
				t.Execute(response, err)
			}
		} else {
			//TODO delete all temp files method
			reset()
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
		collection.UpdateOne(context.TODO(),
			bson.M{"Name": name},
			bson.M{"$pull": bson.M{"Date" + "." + date + ".Prayer." + prayer1 + ".Users": bson.M{"Phone": phone}}})
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
		filter := bson.D{{Key: "Phone", Value: phone}}
		update := bson.D{{Key: "$pull", Value: bson.D{{Key: "RegisteredPrayers", Value: bson.D{{Key: "RpId", Value: rpid}}}}}}
		collection.UpdateOne(context.TODO(), filter, update)
		http.Redirect(response, request, "/", 302)
	} else {
		http.Redirect(response, request, "/login", 401)
		response.Write([]byte(`<script>window.location.href = "/login";</script>`))
	}
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
	err = collection.FindOne(context.TODO(), bson.M{"Phone": phone}).Decode(&user)
	return user, err
}

func check(response http.ResponseWriter, request *http.Request, err error) *template.Template {
	if err != nil {
		t, _ := template.ParseFiles("templates/errorpage.html")
		http.Redirect(response, request, "/error", 402)
		response.Write([]byte(`<script>window.location.href = "/error"</script>`))
		reset()
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

func reset() {
	choo = *new(choose)
	choosenMosque = *new(model.Mosque)
}
