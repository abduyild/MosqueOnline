package handlers
 
import (
	"context"
    "fmt"
    "strings"
    "net/http"
	"errors"
    "html/template"
    helpers "../helpers"
    "os/exec"
	"../model"
	"../repos"
    "github.com/gorilla/securecookie"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"strconv"   

)
type Machine struct {
	ID_Machine int
	SolvedUser bool
	SolvedRoot bool
	UserFlag string
	RootFlag string
}
type Group struct {
	ID int
	Points int
	Machines []Machine
}
// Own Group Structure
type groups struct{
    // ID string, //enable for custom group names
    points int
    userSet [6]bool
    rootSet [6]bool
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
        collection, err:= repos.GetDBCollection(0)
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
		cookie, err := bcrypt.GenerateFromPassword([]byte(name+user.Group), 14) //TODO dont forget to check for correct hashed password
		userCredentials := name + "?" + user.Group + "&" + string(cookie)
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
	collection, err := repos.GetDBCollection(0)
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
	user, err := GetUser(request)
	if err != nil {
        	fmt.Fprintln(response, "There was an Error setting the Website up!")
	}
	userCredentials := strings.Split(user, "&")[0]
	userName := strings.Split(userCredentials, "?")[0]
	group := GetGroup(user)
    if !helpers.IsEmpty(userName) {
		var groupItem Group
		collection, _ := repos.GetDBCollection(1) 
		erro := collection.FindOne(context.TODO(), bson.D{{"ID", group}},).Decode(&groupItem)

		if erro != nil {
        	fmt.Fprintf(response, "There was an Error searching your Group!")
		}
		tpl, _:= template.ParseFiles("templates/index.html")
		tpl.Execute(response, groupItem)
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
	resp, err := GetUser(request)
	if err != nil {
        	fmt.Fprintln(response, "There was an Error setting the Website up!")
	}
	group := strings.Split(resp, "?")[1]
	machine := request.URL.Query().Get("machine")
	command := "qm reset 100" + group + "" + machine
	cmd := exec.Command(command)
	cmd.Run()
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
    
    user, err := GetUser(request)
	if err != nil {
        	fmt.Fprintln(response, "There was an Error setting the Website up!")
	}
	group := GetGroup(user)
	machine := request.URL.Query().Get("machine")
	m,_ := strconv.Atoi(machine)
	
	collection, err := repos.GetDBCollection(1)
	var groupItem Group
	var machines []Machine
	
	if userFlag != "" {
		err := collection.FindOne(context.TODO(), 
			bson.D{ 
				{"ID", group},
			},).Decode(&groupItem)
		if err != nil {
	        fmt.Fprintf(response, "Your Group couldn't be found")
	        return
		}
		machines = groupItem.Machines
		if machines[m].SolvedUser {
	        fmt.Fprintf(response, "Flag already submitted")
	        return
		}
		if machines[m].UserFlag != userFlag {
	        fmt.Fprintf(response, "Wrong flag")
	        return
		}
		collection.UpdateOne(context.TODO(), 
			bson.D{ 
				{"ID", group}, 
			}, 
			bson.D{
				{"$set", bson.D{
					{"Machines."+ machine+".SolvedUser", true},
					{"Points", groupItem.Points + 500},
			    }},
		    },)
	}
	
	if rootFlag != "" {
		err := collection.FindOne(context.TODO(), 
			bson.D{ 
				{"ID", group},
			},).Decode(&groupItem)
		if err != nil {
	        fmt.Fprintf(response, "Your Group could'nt be found")
	        return
		}
		machines = groupItem.Machines
		if machines[m].SolvedRoot {
	        fmt.Fprintf(response, "Flag already submitted")
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
					{"Machines."+ machine+".SolvedRoot", true},
					{"Points", groupItem.Points + 1500},
				}},
			},)
	}
	http.Redirect(response, request, "/index", 302)
}

func SteveJobsHandler(response http.ResponseWriter, request *http.Request) {
	
	var groups []Group
    dataBase, err := repos.GetDBCollection(1)
	if err != nil {
		fmt.Println(response, "error")
	}

	cur, _ := dataBase.Find(context.TODO(), bson.D{})
	// Iterate through whole Collection and append the Array consisting of Groups
	for cur.Next(context.TODO()) {
	    var group Group
	    cur.Decode(&group)
	    groups = append(groups, group)
	}

    	//pageSet := page{Points: groupSet[group].points, SubNetwork: group1}
       // var body, _ = helpers.LoadFile("templates/appleHeadquarter.html")
    	//t := template.Must(template.New("appleHeadquarter").ParseFiles("templates/appleHeadquarter.tmpl"))
	t, _ := template.ParseFiles("templates/appleHeadquarter.gohtml")
    	t.Execute(response, groups)
}

func SetFlag(response http.ResponseWriter, request *http.Request) {
    request.ParseForm()
    userFlag := request.FormValue("user")
    rootFlag := request.FormValue("root")
    
	machines := request.URL.Query().Get("machine")
	
	collection, err := repos.GetDBCollection(1)
	if err != nil {
		fmt.Println(response, "error")
	}
	user,_ := GetUser(request)
	group := GetGroup(user)

// Change to modifying database entry
	if userFlag != "" {
		collection.UpdateOne(context.TODO(), 
			bson.D{ 
				{"ID", group}, 
			}, 
			bson.D{
				{"$set", bson.D{
					{"Machines."+ machines+".UserFlag", userFlag},
				}},
			},)
	}
	if rootFlag != "" {
		collection.UpdateOne(context.TODO(), 
			bson.D{ 
				{"ID", group}, 
			}, 
			bson.D{
				{"$set", bson.D{
					{"Machines."+ machines+".RootFlag", rootFlag},
				}},
			},)
	}
	http.Redirect(response, request, "/appleHeadquarter", 302)
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

func GetGroup(user string) int{
	groupA := strings.Split(user, "?")[1]
	groupB := strings.Split(groupA, "&")[0]
	group,_ := strconv.Atoi(groupB)

	return group
}
 
func GetUser(request *http.Request) (string, error) {
	user := ""
    if cookie, err := request.Cookie("cookie"); err == nil {
        cookieValue := cookie.Value
    	cookieVal := strings.Split(cookieValue, "&")[0]
    	values := strings.Split(cookieVal, "?")
    	cookieHash := strings.Split(cookieValue, "&")[1]
        err = bcrypt.CompareHashAndPassword([]byte(cookieHash), []byte(values[0] + values[1]))
	    if err != nil {
	    	return "", errors.New("wrong")
	    }
        user = cookieValue
    }
    return user, nil
}
