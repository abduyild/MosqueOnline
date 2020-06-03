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

// Struct for easy handling the htm,l template generation on the index page
type TempPrayer struct {
	Name     model.PrayerName
	Capacity int
}

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
	check(response, request, err)
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

			userCredentials, err := bcrypt.GenerateFromPassword([]byte(email+phone), 14)
			cookie := email + "?" + phone + "&" + string(userCredentials)
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
		redirectTarget := "/"
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
			userCredentials, err := bcrypt.GenerateFromPassword([]byte(email+name), 14)
			cookie := email + "?" + name + "&" + string(userCredentials)
			SetCookie(cookie, response)
			// If the admin tries to login, change the redirect to the Adminpage
			if admin.Admin {
				redirectTarget = "/admin"
				// Else redirect to the normal indexpage
			} else {
				redirectTarget = "/mosqueoverview"
			}
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

func DeleteMosque(response http.ResponseWriter, request *http.Request) {
	mosque := request.URL.Query().Get("mosque")
	collection, err := repos.GetDBCollection(1)
	check(response, request, err)
	collection.DeleteOne(context.TODO(), bson.M{"Name": mosque})

	collection, err = repos.GetDBCollection(0)
	check(response, request, err)
	update := bson.D{{Key: "$pull", Value: bson.D{{Key: "RegisteredPrayers", Value: bson.D{{Key: "MosqueName", Value: mosque}}}}}}
	collection.UpdateMany(context.TODO(), bson.D{{}}, update)

}

func MosqueHandler(response http.ResponseWriter, request *http.Request) {
	t, _ := template.ParseFiles("templates/mosqueOverview.html")
	t.Execute(response, nil)
}

// Function for Handling the Pagecall of the Indexpage
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
	}
}

// Function for handling the call to logout, simply deletes the cookie associated with the session and redirects to loginpage
func LogoutHandler(response http.ResponseWriter, request *http.Request) {
	ClearCookie(response)
	http.Redirect(response, request, "/", 302)
}

func Choosen(response http.ResponseWriter, request *http.Request) {
	mosque := request.URL.Query().Get("mosque")
	collection, _ := repos.GetDBCollection(1)
	err := collection.FindOne(context.TODO(),
		bson.D{
			{"Name", mosque},
		}).Decode(&choosenMosque)
	check(response, request, err)
	t, _ := template.ParseFiles("templates/chooseDate.html")
	t.Execute(response, choosenMosque)
	//http.Redirect(response, request, "/chooseDate", 302)
}

func Choose(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		mosque := request.URL.Query().Get("mosque")
		if !choo.SetMosque {
			dataBase, err := repos.GetDBCollection(1)
			check(response, request, err)
			cur, _ := dataBase.Find(context.TODO(), bson.D{})
			for cur.Next(context.TODO()) {
				var mosque model.Mosque
				err = cur.Decode(&mosque)
				check(response, request, err)
				choo.Mosques = append(choo.Mosques, mosque)
			}
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
					}
				}
			} else {
				reset()
				Choose(response, request)
			}
		}
	} else {
		http.Redirect(response, request, "/login", 401)
	}
}

func getMosques(response http.ResponseWriter, request *http.Request) mosques {
	mosquess := []model.Mosque{}
	dataBase, err := repos.GetDBCollection(1)
	check(response, request, err)
	cur, _ := dataBase.Find(context.TODO(), bson.D{})
	for cur.Next(context.TODO()) {
		var mosque model.Mosque
		err = cur.Decode(&mosque)
		check(response, request, err)
		mosquess = append(mosquess, mosque)
	}
	return mosquess
}

func getUsers(response http.ResponseWriter, request *http.Request) users {
	users := []model.User{}
	dataBase, err := repos.GetDBCollection(0)
	check(response, request, err)
	cur, _ := dataBase.Find(context.TODO(), bson.D{})
	for cur.Next(context.TODO()) {
		var user model.User
		err = cur.Decode(&user)
		check(response, request, err)
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
		male := true
		user, err := GetUserAsUser(response, request)
		if err != nil {
			http.Redirect(response, request, "/login", 302)
			reset()
		}
		if user.Sex == "Women" {
			male = false
		}
		for _, prayer := range choosenMosque.Date[index].Prayer {
			if male {
				cap = prayer.CapacityMen
			} else {
				cap = prayer.CapacityWomen
			}
			pray := *new(TempPrayer)
			pray.Name = prayer.Name
			pray.Capacity = cap
			choo.Prayer = append(choo.Prayer, pray)
		}
		t, _ := template.ParseFiles("templates/choosePrayer.html")
		t.Execute(response, choo)
		choo.SetDate = true
	} else {
		http.Redirect(response, request, "/login", 401)
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
			err = collection.FindOne(context.TODO(),
				bson.D{
					{"Name", choosenMosque.Name},
				}).Decode(&mosque)
			check(response, request, err)
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
			if err1.Err() != nil {
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
				collection, _ = repos.GetDBCollection(0)
				phone, err := GetPhoneFromCookie(request)
				check(response, request, err)
				collection.UpdateOne(context.TODO(),
					bson.M{"Phone": phone}, bson.M{
						"$push": bson.M{"RegisteredPrayers": registered}})
				choo = *new(choose)
				http.Redirect(response, request, "/", 302)
			} else {
				PrintError(response, request, errors.New("You are already Signed in for this prayer"))
			}
		} else {
			//TODO delete all temp files method
			reset()
			http.Redirect(response, request, "/", 302)
		}
	} else {
		http.Redirect(response, request, "/login", 401)
	}
}

func SignOutPrayer(response http.ResponseWriter, request *http.Request) {
	if loggedin(response, request) {
		name := request.FormValue("name")
		date := request.FormValue("date")
		prayer := request.FormValue("prayer")
		phone := request.FormValue("phone")
		prayerN, err := strconv.Atoi(prayer)
		check(response, request, err)
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
	match, err := regexp.MatchString("[^s]+?[^s]+&[^s]+", cookie.Value)
	if err == nil && match {
		cookieValue := cookie.Value
		cookieVal := strings.Split(cookieValue, "&")[0]
		values := strings.Split(cookieVal, "?")
		cookieHash := strings.Split(cookieValue, "&")[1]
		err = bcrypt.CompareHashAndPassword([]byte(cookieHash), []byte(values[0]+values[1]))
		if err != nil {
			return "", errors.New("Wrong or Modified Cookie")
		}
		phone = values[1]
	} else {
		return "", errors.New("Invalid Cookie")
	}
	return phone, nil
}

func GetUserAsUser(response http.ResponseWriter, request *http.Request) (model.User, error) {
	var user model.User
	phone, err := GetPhoneFromCookie(request)
	check(response, request, err)
	collection, err := repos.GetDBCollection(0)
	check(response, request, err)
	err = collection.FindOne(context.TODO(), bson.M{"Phone": phone}).Decode(&user)
	return user, err
}

func check(response http.ResponseWriter, request *http.Request, err error) {
	if err != nil {
		PrintError(response, request, err)
	}
}

// TODO: switch with error messages because of mongo etc.
func PrintError(response http.ResponseWriter, request *http.Request, err error) {
	t, _ := template.ParseFiles("templates/errorpage.html")
	if strings.Contains(err.Error(), "<") {
		err = errors.New(strings.Split(err.Error(), "<")[0])
	}
	t.Execute(response, err)
	reset()
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
