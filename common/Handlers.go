package handlers
 
import (
	"context"
    "fmt"
    "strings"
    "net/http"
    "html/template"
    helpers "../helpers"
    //"os/exec"
	"../model"
	"../repos"
    "github.com/gorilla/securecookie"
    
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"

)

// Mapping groups to an Integer which is storing the Points of the Group
var groupPoints = map[string]int{
    "1": 0,
    "2": 0,
    "3": 0,
    "4": 0,
    "5": 0,
    "6": 0,
}

// Type in your User Flags for the equivalent Virtual Machine instead of "user1" etc.
var userFlags = map[string]string{
    "1": "user1",
    "2": "user2",
    "3": "user3",
    "4": "user4",
    "5": "user5",
    "6": "user6",
}

// Type in your Root Flags for the equivalent Virtual Machine "root1" etc.
var rootFlags = map[string]string{
    "1": "root1",
    "2": "root2",
    "3": "root3",
    "4": "root4",
    "5": "root5",
    "6": "root6",
}

var cookieHandler = securecookie.New(
    securecookie.GenerateRandomKey(64),
    securecookie.GenerateRandomKey(32))

 
// for GET
func LoginPageHandler(response http.ResponseWriter, request *http.Request) {
    var body, _ = helpers.LoadFile("templates/login.html")
    fmt.Fprintf(response, body)
}
 
// for POST
func LoginHandler(response http.ResponseWriter, request *http.Request) {
    name := request.FormValue("name")
    pass := request.FormValue("password")
    redirectTarget := "/"
    if !helpers.IsEmpty(name) && !helpers.IsEmpty(pass) {
        // Database check for user data
        collection, err := repos.GetDBCollection()
        if err != nil {
    		http.Redirect(response, request, "/register", 302)
		}
		
		var user model.User
		// Checking if typed in Username exists, if not redirect to register page
        err = collection.FindOne(context.TODO(), bson.D{{"username", name}}).Decode(&user)
		if err != nil {
    		http.Redirect(response, request, "/register", 302)
		}
		// Checking if typed in password is equivalent to the password typed in registry process, if not redirect to register page
        err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))
		if err != nil {
    		http.Redirect(response, request, "/register", 302)
		}
		userCredentials := name + "?" + user.Group
        SetCookie(userCredentials, response)
        
        if name == "steveJobs" {
        	redirectTarget = "/appleHeadquarter"
        } else {
        	redirectTarget = "/index"
        }
    }
    http.Redirect(response, request, redirectTarget, 302)
}
 
// for GET
func RegisterPageHandler(response http.ResponseWriter, request *http.Request) {
    var body, _ = helpers.LoadFile("templates/register.html")
    fmt.Fprintf(response, body)
}
 

// for POST
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	collection, err := repos.GetDBCollection()
	r.ParseForm()
 
    uName := r.FormValue("username")
    group := r.FormValue("group")
    pwd := r.FormValue("password")

 	// initializing as false for not filled
    _uName, _group, _pwd := false, false, false
    _uName = !helpers.IsEmpty(uName)
    _group = !helpers.IsEmpty(group)
    _pwd = !helpers.IsEmpty(pwd)
 
    if _uName && _group && _pwd{

		var user model.User
		err = collection.FindOne(context.TODO(), bson.D{{"username", uName}}).Decode(&user)

		if err != nil{
			hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 14)

			if err != nil {
				return
			}
			user.Username = string(uName)
			user.Password = string(hash)
			user.Group = string(group)
			usr := model.User{uName, group, string(hash)}
			collection.InsertOne(context.TODO(), usr)
   			http.Redirect(w, r, "/", 302)
        } else { 
        	fmt.Fprintln(w, "User already exists")
        }
    } else {
        fmt.Fprintln(w, "This fields can not be blank!")
    }
}
 
// for GET
func IndexPageHandler(response http.ResponseWriter, request *http.Request) {
	user := GetUser(request)
	userName := strings.Split(user, "?")[0]
	group := strings.Split(user, "?")[1]
    if !helpers.IsEmpty(userName) {
    	type page struct {
    		Points int
    		SubNetwork string
    	}
    	pageSet := page{Points: groupPoints[group], SubNetwork: group}
    	
    	// Changes the {{.$subNetwork}} and {{.$points}} tags in the htmtl file to the according groupID and Points of Group from User.
		//tpl := template.Must(template.ParseFiles("templates/index.html"))
		tpl, _:= template.ParseFiles("templates/index.html")
		tpl.Execute(response, pageSet)
    } else {
        http.Redirect(response, request, "/", 302)
    }
}
 
// for POST
func LogoutHandler(response http.ResponseWriter, request *http.Request) {
    ClearCookie(response)
    http.Redirect(response, request, "/", 302)
}

// Handling the Reset of a machine
func ResetHandler(response http.ResponseWriter, request *http.Request) {
	resp := GetUser(request)
	group := strings.Split(resp, "?")[1]
	machine := request.URL.Query().Get("machine")
	
	//For testing using fedora, but use the one below, currently not working though
	/*
	command := "qemu-system-i386 -machine fedora loadvm"
	cmd := exec.Command(command)
	cmd.Run()
	*/
	command := "qemu-system-i386 -machine 10.0." + group + "." + machine + " loadvm"
	fmt.Fprintln(response, command)
}
 
// Cookie
func SetCookie(user string, response http.ResponseWriter) {
    cookie := &http.Cookie{
    	Name: "cookie",
    	Value: user,
    	Path: "/",
    }
    http.SetCookie(response, cookie)
}

func SubmitHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
    userFlag := request.FormValue("user")
    rootFlag := request.FormValue("root")
    
    user := GetUser(request)
    //userName := strings.Split(user, "?")[0]
	group := strings.Split(user, "?")[1]
	machine := request.URL.Query().Get("machine")
	
	// TODO prevent multiple Inputs, maybe boolean, but where?
	if userFlag == userFlags[machine] {
		groupPoints[group] += 500
	} 
	if rootFlag == rootFlags[machine] {
		groupPoints[group] += 1500
	}
	http.Redirect(response, request, "/index", 302)
}

func SteveJobsHandler(response http.ResponseWriter, request *http.Request) {
		type points struct{
			Points1 int
			Points2 int
			Points3 int
			Points4 int
			Points5 int
			Points6 int
		}
    	point := points{
    		Points1: groupPoints["1"], 
    		Points2: groupPoints["2"], 
    		Points3: groupPoints["3"], 
    		Points4: groupPoints["4"], 
    		Points5: groupPoints["5"], 
    		Points6: groupPoints["6"]}
		tpl, _:= template.ParseFiles("templates/appleHeadquarter.html")
		tpl.Execute(response, point)
}
 
func ClearCookie(response http.ResponseWriter) {
    cookie := &http.Cookie{
        Name:   "cookie",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    }
    http.SetCookie(response, cookie)
}
 
func GetUser(request *http.Request) (user string) {

    if cookie, err := request.Cookie("cookie"); err == nil {
        cookieValue := cookie.Value
            user = cookieValue
    }
    return user
}
