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

type Mosque struct {
	Name   string
	PLZ    int
	Street string
	City   string
	Date   []model.Date
}

type Mosques struct {
	Mosques []Mosque
}

var choosenMosque model.Mosque

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

// Handler for Login Page used with GET, shows login page
func LoginPageHandler(response http.ResponseWriter, request *http.Request) {
	var body, _ = helpers.LoadFile("templates/login.html")
	fmt.Fprintf(response, body)
}

// Handler for Login Page used with POST by submitting Loginform
func LoginHandler(response http.ResponseWriter, request *http.Request) {
	email := request.FormValue("email")
	phone := request.FormValue("phone")
	// Default redirect page is the login page, so if anything goes wrong, the program just redirects to the login page again
	redirectTarget := "/"
	if !helpers.IsEmpty(email) && !helpers.IsEmpty(phone) {
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

		if err != nil {
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
			redirectTarget = "/chooseMosque"
		}
	}
	// function for redirecting
	http.Redirect(response, request, redirectTarget, 302)
}

// Function for handling the Webpage call with GET
func RegisterPageHandler(response http.ResponseWriter, request *http.Request) {
	var body, _ = helpers.LoadFile("templates/register.html")
	fmt.Fprintf(response, body)
}

// Function for handling the register action submitted by user
func RegisterHandler(response http.ResponseWriter, request *http.Request) {
	collection, err := repos.GetDBCollection(0)
	if err != nil {
		fmt.Fprintf(response, "There was an error")
		return
	}
	request.ParseForm()

	// Get data the User typen into the fields
	firstName := request.FormValue("firstname")
	lastName := request.FormValue("lastname")
	email := request.FormValue("email")
	phone := request.FormValue("phone")

	// initializing as false for not filled
	_firstName, _lastName, _email, _phone := false, false, false, false
	_firstName = !helpers.IsEmpty(firstName)
	_lastName = !helpers.IsEmpty(lastName)
	_email = !helpers.IsEmpty(email)
	_phone = !helpers.IsEmpty(phone)
	// Check if fields are not empty
	if _firstName && _lastName && _email && _phone {
		// Look if the entered Username is already used
		user := collection.FindOne(context.TODO(), bson.D{{"Phone", phone}})
		// If not found (throws exception/error) then we can proceed
		if user.Err() != nil {
			// Generate the hashed password with 14 as salt

			//use hash if you want
			//hash, err := bcrypt.GenerateFromPassword([]byte(phone), 14)

			// If there was an error generating the hash dont proceed
			if err != nil {
				return
			}
			// define a User model with typed first and last name, email and phone

			//usr := model.User{firstName, email, string(hash)}

			usr := model.User{firstName, lastName, email, phone, false}
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

// Function for Handling the Pagecall of the Indexpage
func IndexPageHandler(response http.ResponseWriter, request *http.Request) {
	user, err := GetUser(request)
	if err != nil {
		fmt.Fprintln(response, err)
		return
	}

	phone := GetPhone(user)
	if !helpers.IsEmpty(phone) {
		var mosqueItem model.Mosque
		collection, _ := repos.GetDBCollection(1)
		// Search for group with given ID as group and decode, if not possibe to decode erro != nil
		erro := collection.FindOne(context.TODO(), bson.D{{"Name", choosenMosque.Name}}).Decode(&mosqueItem)
		// If there was an error decoding the item with the Databasequery, throw an error
		if erro != nil {
			fmt.Fprintf(response, "There was an Error getting your Mosque!")
			return
		}
		// Parse the templatefile, changes all Placeholders {{ }} with appropiate Values
		tpl, _ := template.ParseFiles("templates/index.html")
		// Inserts the groups to the Template as we are using the ID and points of the group
		tpl.Execute(response, mosqueItem)

	} else {
		http.Redirect(response, request, "/", 302)
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
	if err != nil {
		fmt.Fprintf(response, "Your Mosque couldn't be found: "+err.Error())
		return
	}
	t, _ := template.ParseFiles("templates/chooseDate.html")
	t.Execute(response, choosenMosque)
	//http.Redirect(response, request, "/chooseDate", 302)
}

type choose struct {
	Name       string
	City       string
	SetMosque  bool
	SetDate    bool
	SetPrayer  bool
	Mosques    []model.Mosque
	Date       model.Date
	DateString string
	Prayer     []model.Prayer
	PrayerName string
}

var choo choose

func Choose(response http.ResponseWriter, request *http.Request) {
	mosque := request.URL.Query().Get("mosque")

	if !choo.SetMosque {
		fmt.Println("OOH")
		dataBase, err := repos.GetDBCollection(1)
		if err != nil {
			fmt.Println(response, "error")
			return
		}
		cur, _ := dataBase.Find(context.TODO(), bson.D{})
		// Iterate through whole Collection and append the Array consisting of Groups
		for cur.Next(context.TODO()) {
			var mosque model.Mosque
			erro := cur.Decode(&mosque)
			if erro != nil {
				fmt.Println(erro.Error())
			}
			choo.Mosques = append(choo.Mosques, mosque)
		}
		t, _ := template.ParseFiles("templates/choose.html")
		t.Execute(response, choo)
		choo.SetMosque = true
	} else {
		fmt.Println(mosque)
		for _, mosq := range choo.Mosques {
			if mosq.Name == mosque {
				choosenMosque = mosq
				t, _ := template.ParseFiles("templates/choose.html")
				t.Execute(response, choo)
			}
		}
	}
}

func ChooseDate(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	date := request.Form.Get("date")
	fmt.Println("input date: " + date)
	var dateF model.Date
	index := 0
	dateF.Date, _ = time.Parse(time.RFC3339, date)
	for i, dates := range choosenMosque.Date {
		if dateF.Date == dates.Date {
			dateF = dates
			index = i
			choo.Date = dateF
			break
		}
	}
	for _, prayer := range choosenMosque.Date[index].Prayer {
		choo.Prayer = append(choo.Prayer, prayer)
	}
	t, _ := template.ParseFiles("templates/choosePrayer.html")
	t.Execute(response, choo)
	choo.SetDate = true
}

func ChoosePrayer(response http.ResponseWriter, request *http.Request) {
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
	fmt.Println("Name: " + choo.Name)
	fmt.Println("date: " + choo.Date.Date.String())
	fmt.Println("PrayerName: " + choo.PrayerName)
	t, _ := template.ParseFiles("templates/confirm.html")
	t.Execute(response, choo)
}

func SubmitPrayer(response http.ResponseWriter, request *http.Request) {
	pray, _ := strconv.Atoi(request.URL.Query().Get("prayer"))
	collection, err := repos.GetDBCollection(1)
	if err != nil {
		fmt.Fprintf(response, "Error trying to connect to DB")
		return
	}
	var mosque model.Mosque
	//var date model.Date
	//var prayer model.Prayer
	// Search for the group in the Table with the groupID equivalent to the users groupID and decode it
	err = collection.FindOne(context.TODO(),
		bson.D{
			{"Name", choosenMosque.Name},
		}).Decode(&mosque)
	if err != nil {
		fmt.Fprintf(response, "Your Mosque couldn't be found")
		return
	}
	index := 0
	for i, dates := range choosenMosque.Date {
		if choo.Date.Date == dates.Date {
			index = i
			break
		}
	}
	collection.UpdateOne(context.TODO(),
		bson.D{
			{"Name", choosenMosque.Name},
		},
		bson.D{
			{"$set", bson.D{
				{"Capacity", choosenMosque.Date[index].Prayer[pray].Capacity - 1},
			}},
		})
	choosenMosque.Date[index].Prayer[pray].Capacity--
	// TODO: check if works
	// TODO reset choosenMosque
	http.Redirect(response, request, "/index", 302)
}

func ChooseMosque(response http.ResponseWriter, request *http.Request) Mosques {
	var mosq Mosques
	dataBase, err := repos.GetDBCollection(1)
	if err != nil {
		http.Redirect(response, request, "/404", 302)
	}
	cur, _ := dataBase.Find(context.TODO(), bson.D{})
	for cur.Next(context.TODO()) {
		var mosque Mosque
		cur.Decode(&mosque)
		mosq.Mosques = append(mosq.Mosques, mosque)
	}
	return mosq
}

func SubmitHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	prayer := request.FormValue("prayer")
	date := request.URL.Query().Get("date")
	fmt.Println(date)
	user := GetUserAsUser(request)

	collection, err := repos.GetDBCollection(1)
	if err != nil {
		fmt.Fprintf(response, "Error trying to connect to DB")
		return
	}
	var prayerModel model.Prayer
	prayerModel.Name = model.Sabah
	var mosque Mosque
	if helpers.IsEmpty(prayer) {
		// Search for the group in the Table with the groupID equivalent to the users groupID and decode it
		err := collection.FindOne(context.TODO(),
			bson.D{
				{"Name", choosenMosque.Name},
			}).Decode(&mosque)
		if err != nil {
			fmt.Fprintf(response, "Your Mosque couldn't be found")
			return
		}
		for dates := range mosque.Date {
			if dates == dates {

				break
			}
		}
		/*collection.UpdateOne(context.TODO(),
		bson.D{
			{"Name", choosenMosque.Name},
		},
		bson.D{
			{"$set", bson.D{
				{"Capacity", choosenMosque.Capacity - 1},
			}},
		})*/
		collection.UpdateOne(context.TODO(), bson.D{}, bson.D{{"$push", bson.M{"Users": user}}})
		// TODO: check if works
		// TODO reset choosenMosque
		http.Redirect(response, request, "/index", 302)
	} else {
		fmt.Fprintln(response, "Date and prayer selection may not be empty")
		return
	}
}

/*
// Function for Handling the Submitforms
func SubmitHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	userFlag := request.FormValue("user")
	rootFlag := request.FormValue("root")

	user, err := GetUser(request)
	if err != nil {
		fmt.Fprintln(response, err)
		return
	}
	group := GetPhone(user)
	machine := request.URL.Query().Get("machine")
	m, _ := strconv.Atoi(machine)

	collection, err := repos.GetDBCollection(1)
	var groupItem Group
	var machines []Machine

	if userFlag != "" && CheckInput(userFlag) {
		// Search for the group in the Table with the groupID equivalent to the users groupID and decode it
		err := collection.FindOne(context.TODO(),
			bson.D{
				{"ID", group},
			}).Decode(&groupItem)
		if err != nil {
			fmt.Fprintf(response, "Your Group couldn't be found")
			return
		}
		machines = groupItem.Machines
		if machines[m].SolvedUser {
			fmt.Fprintf(response, "UserFlag already submitted")
			return
		}
		if machines[m].UserFlag != userFlag {
			fmt.Fprintf(response, "Wrong flag")
			return
		}
		// Update entry by changing the flag for a solved Userflag and add 500 Points to the groups points
		collection.UpdateOne(context.TODO(),
			bson.D{
				{"ID", group},
			},
			bson.D{
				{"$set", bson.D{
					{"Machines." + machine + ".SolvedUser", true},
					{"Points", groupItem.Points + 500},
				}},
			})
	} else if userFlag!= "" && !CheckInput(userFlag) {
		fmt.Fprintf(response, "Invalid Characters found")
	}

	if rootFlag != ""  && CheckInput(rootFlag) {
		err := collection.FindOne(context.TODO(),
			bson.D{
				{"ID", group},
			}).Decode(&groupItem)
		if err != nil {
			fmt.Fprintf(response, "Your Group couldn't be found")
			return
		}
		machines = groupItem.Machines
		if machines[m].SolvedRoot {
			fmt.Fprintf(response, "RootFlag already submitted")
			return
		}
		if machines[m].RootFlag != rootFlag {
			fmt.Fprintf(response, "Wrong flag")
			return
		}
		collection.UpdateOne(context.TODO(),
			bson.D{
				{"ID", group},
			},
			bson.D{
				{"$set", bson.D{
					{"Machines." + machine + ".SolvedRoot", true},
					{"Points", groupItem.Points + 1500},
				}},
			})
	} else if rootFlag!= "" && !CheckInput(rootFlag) {
		fmt.Fprintf(response, "Invalid Characters found")
	}
	http.Redirect(response, request, "/index", 302)
}

func SteveJobsHandler(response http.ResponseWriter, request *http.Request) {
	var groups []Group
	dataBase, err := repos.GetDBCollection(1)
	if err != nil {
		fmt.Println(response, "error")
		return
	}

	cur, _ := dataBase.Find(context.TODO(), bson.D{})
	// Iterate through whole Collection and append the Array consisting of Groups
	for cur.Next(context.TODO()) {
		var group Group
		cur.Decode(&group)
		groups = append(groups, group)
	}

	t, _ := template.ParseFiles("templates/appleHeadquarter.gohtml")
	t.Execute(response, groups)
}
*/
func SetFlag(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	userFlag := request.FormValue("user")
	rootFlag := request.FormValue("root")
	machines := request.FormValue("machine")
	group, _ := strconv.Atoi(request.FormValue("group"))
	collection, err := repos.GetDBCollection(1)
	if err != nil {
		fmt.Println(response, "error")
		return
	}

	if userFlag != "" {
		// Update the Userflag of one groups machine with given machineID to given userFlag
		collection.UpdateOne(context.TODO(),
			bson.D{
				{"ID", group},
			},
			bson.D{
				{"$set", bson.D{
					{"Machines." + machines + ".UserFlag", userFlag},
				}},
			})
	}
	if rootFlag != "" {
		collection.UpdateOne(context.TODO(),
			bson.D{
				{"ID", group},
			},
			bson.D{
				{"$set", bson.D{
					{"Machines." + machines + ".RootFlag", rootFlag},
				}},
			})
	}
	http.Redirect(response, request, "/appleHeadquarter", 302)
}

// Function for changing  the Flags for all Machines
func SetAllFlags(response http.ResponseWriter, request *http.Request) {
	userFlag := request.FormValue("user")
	rootFlag := request.FormValue("root")

	collection, err := repos.GetDBCollection(1)
	if err != nil {
		fmt.Println(response, "Error getting the DB")
	}

	if userFlag != "" {
		// Update all Machines for all Groups
		_, err := collection.UpdateMany(context.TODO(),
			bson.D{},
			bson.D{
				{"$set", bson.D{
					{"Machines.$[].UserFlag", userFlag},
				}},
			})
		if err != nil {
			fmt.Fprintf(response, "Error")
		}
	}
	if rootFlag != "" {
		_, err := collection.UpdateMany(context.TODO(),
			bson.D{},
			bson.D{
				{"$set", bson.D{
					{"Machines.$[].RootFlag", rootFlag},
				}},
			})
		if err != nil {
			fmt.Fprintf(response, "Error")
		}
	}
	http.Redirect(response, request, "/appleHeadquarter", 302)
}

// Function for changing  the Flags for one Machine for all Groups
func SetAllFlagsForOne(response http.ResponseWriter, request *http.Request) {
	userFlag := request.FormValue("user")
	rootFlag := request.FormValue("root")
	machines := request.FormValue("machine")

	collection, err := repos.GetDBCollection(1)
	if err != nil {
		fmt.Println(response, "Error getting the DB")
	}

	if userFlag != "" {
		// Update one particlular Machine for all Groups
		_, err := collection.UpdateMany(context.TODO(),
			bson.D{},
			bson.D{
				{"$set", bson.D{
					{"Machines." + machines + ".UserFlag", userFlag},
				}},
			})
		if err != nil {
			fmt.Fprintf(response, "Error")
		}
	}
	if rootFlag != "" {
		_, err := collection.UpdateMany(context.TODO(),
			bson.D{},
			bson.D{
				{"$set", bson.D{
					{"Machines." + machines + ".RootFlag", rootFlag},
				}},
			})
		if err != nil {
			fmt.Fprintf(response, "Error")
		}
	}
	http.Redirect(response, request, "/appleHeadquarter", 302)
}

// Function for adding a VM to the Table
func AddMosque(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	machines := request.FormValue("machine")
	userFlag := request.FormValue("user")
	rootFlag := request.FormValue("root")

	collection, err := repos.GetDBCollection(2)
	if err != nil {
		fmt.Println(response, "error getting DataBase")
		return
	}
	if userFlag == "" {
		userFlag = "user"
	}
	if rootFlag == "" {
		rootFlag = "root"
	}
	if machines != "" {
		// Update the Userflag of one groups machine with given machineID to given userFlag
		machineID, _ := strconv.Atoi(machines)
		collection.UpdateMany(context.TODO(),
			bson.D{},
			bson.D{
				{"$push", bson.M{"Machines": bson.M{
					"ID_Machine": machineID,
					"SolvedUser": false,
					"SolvedRoot": false,
					"UserFlag":   userFlag,
					"RootFlag":   rootFlag,
				},
				},
				}},
		)
	}
	http.Redirect(response, request, "/appleHeadquarter", 302)
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

// Function for getting the groupID from the Cookie (cookievalue as String)
func GetPhone(user string) string {
	groupA := strings.Split(user, "?")[1]  // String after "?"
	phone := strings.Split(groupA, "&")[0] // String before "&"

	return phone
}

func CheckInput(input string) bool {
	if match, _ := regexp.MatchString(`\w+`, input); match == true {
		return true
	}
	return false
}

func GetUser(request *http.Request) (string, error) {
	user := ""
	// Check if there is an active Cookie
	if cookie, err := request.Cookie("cookie"); err == nil {
		cookieValue := cookie.Value
		cookieVal := strings.Split(cookieValue, "&")[0]
		values := strings.Split(cookieVal, "?")
		cookieHash := strings.Split(cookieValue, "&")[1]
		err = bcrypt.CompareHashAndPassword([]byte(cookieHash), []byte(values[0]+values[1]))
		if err != nil {
			return "", errors.New("Wrong or Modified Cookie")
		}
		user = cookieValue
	}
	return user, nil
}

func GetUserAsUser(request *http.Request) model.User {
	var user model.User
	phone, _ := GetUser(request)
	collection, _ := repos.GetDBCollection(0)
	collection.FindOne(context.TODO(), bson.D{{"Phone", phone}}).Decode(&user)
	return user
}
