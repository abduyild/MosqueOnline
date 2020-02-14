package handlers
 
import (
	"context"
    "fmt"
    "strings"
    "regexp"
    "net/http"
	"errors"
    "html/template"
    "../helpers"
    "os/exec"
	"../model"
	"../repos"
    "github.com/gorilla/securecookie"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"strconv"   

)
// Machine struct for using it for initializing variables for Database operations
type Machine struct {
	ID_Machine int
	SolvedUser bool
	SolvedRoot bool
	UserFlag string
	RootFlag string
}

// Group struct for using it for initializing variables for Database operations
type Group struct {
	ID int
	Points int
	Machines []Machine
}

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
    name := request.FormValue("name")
    pass := request.FormValue("password")
// Default redirect page is the login page, so if anything goes wrong, the program just redirects to the login page again
    redirectTarget := "/"
    if !helpers.IsEmpty(name) && !helpers.IsEmpty(pass) {
        // Returns Table
        collection, err:= repos.GetDBCollection(0)
	// if there was no error getting the table, te program does these operations
        if err != nil {
    		http.Redirect(response, request, "/register", 302)
		}
		var user model.User
		// Checking if typed in Username exists, if not redirect to register page
        err = collection.FindOne(context.TODO(), bson.D{{"username", name}}).Decode(&user)
		// If there was an error getting an entry with matching username (no user with this username) redirect to faultpage	
		if err != nil {
    		http.Redirect(response, request, "/register", 302)
		}
		// Checking if typed in password is equivalent to the password typed in registry process, if not redirect to faultpage
        err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))
		if err != nil {
    		http.Redirect(response, request, "/register", 302)
		}
		
		userCredentials, err := bcrypt.GenerateFromPassword([]byte(name+user.Group), 14)
		cookie := name + "?" + user.Group + "&" + string(userCredentials)
        SetCookie(cookie, response)
        // If the admin tries to login, change the redirect to the Adminpage
        if name == "steveJobs" {
        	redirectTarget = "/appleHeadquarter"
	// Else redirect to the normal indexpage
	} else {
        	redirectTarget = "/index"
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
    uName := request.FormValue("username")
    group := request.FormValue("group")
    pwd := request.FormValue("password")

 	// initializing as false for not filled
    _uName, _group, _pwd := false, false, false
    _uName = !helpers.IsEmpty(uName)
    _group = !helpers.IsEmpty(group)
    _pwd = !helpers.IsEmpty(pwd)
    // Check if fields are not empty
    if _uName && _group && _pwd {
		// Look if the entered Username is already used
		user := collection.FindOne(context.TODO(), bson.D{{"username", uName}})
		// If not found (throws exception/error) then we can proceed
		if user.Err() != nil{
		    // Generate the hashed password with 14 as salt
			hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 14)
            // If there was an error generating the hash dont proceed
			if err != nil {
				return
			}
			// define a User model with typed username, group and hashed password
			usr := model.User{uName, group, string(hash)}
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
	userCredentials := strings.Split(user, "&")[0]
	userName := strings.Split(userCredentials, "?")[0]
	group := GetGroup(user)
    if !helpers.IsEmpty(userName) {
		var groupItem Group
		collection, _ := repos.GetDBCollection(1)
		// Search for group with given ID as group and decode, if not possibe to decode erro != nil
		erro := collection.FindOne(context.TODO(), bson.D{{"ID", group}},).Decode(&groupItem)
        // If there was an error decoding the item with the Databasequery, throw an error
		if erro != nil {
        	fmt.Fprintf(response, "There was an Error searching your Group!")
        	return
		}
		// Parse the templatefile, changes all Placeholders {{ }} with appropiate Values
		tpl, _:= template.ParseFiles("templates/index.html")
		// Inserts the groups to the Template as we are using the ID and points of the group
		tpl.Execute(response, groupItem)
    } else {
        http.Redirect(response, request, "/", 302)
    }
}
 
// Function for handling the call to logout, simply deletes the cookie associated with the session and redirects to loginpage
func LogoutHandler(response http.ResponseWriter, request *http.Request) {
    ClearCookie(response)
    http.Redirect(response, request, "/", 302)
}

// Handling the Reset of a machine
func ResetHandler(response http.ResponseWriter, request *http.Request) {
	resp, err := GetUser(request)
	if err != nil {
        	fmt.Fprintln(response, err)
        	return
	}
	group := strconv.Itoa(GetGroup(resp))
	machine := request.URL.Query().Get("machine")
	id := group + "." + machine // Forms a string of the form: groupID.MachineID f.ex group 3 and machine 2: 3.2
	// Virsh Command for connecting to Console of Guest and resetting to snapshot
	// Snapshots in form: id (in ex. above: id = 3.2)
	//command := "virsh console " + id + "\n virsh snapshot-revert " + id + " " +id 
	params := "snapshot-revert" + id + " " + id
	
	// this works because first input needs to be command, everything after are for parameters
	cmd := exec.Command("virsh", params)
	err = cmd.Run()
	if err != nil {
	    fmt.Println(err)
	}
	/* If above doesnt work use this and delete the second command beginning with \n from the string
	command2 := "virsh snapshot-revert " + id + " " +id 
	cmd2 := exec.Command(command2)
	cmd2.Run()
	*/
}

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
	group := GetGroup(user)
	machine := request.URL.Query().Get("machine")
	m,_ := strconv.Atoi(machine)
	
	collection, err := repos.GetDBCollection(1)
	var groupItem Group
	var machines []Machine
	
	if userFlag != "" {
	    // Search for the group in the Table with the groupID equivalent to the users groupID and decode it
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

func SetFlag(response http.ResponseWriter, request *http.Request) {
    request.ParseForm()
    userFlag := request.FormValue("user")
    rootFlag := request.FormValue("root")
    machines := request.FormValue("machine")
    group,_ := strconv.Atoi(request.FormValue("group"))
    
	//machines := request.URL.Query().Get("machine")
	//group, _ := strconv.Atoi(request.URL.Query().Get("group"))
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
			},)
		if err != nil {
		    fmt.Fprintf(response, "Error")
		}
	}
	if rootFlag != "" {
		_, err :=collection.UpdateMany(context.TODO(), 
			bson.D{}, 
			bson.D{
				{"$set", bson.D{
					{"Machines.$[].RootFlag", rootFlag},
				}},
			},)
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
			},)
		if err != nil {
		    fmt.Fprintf(response, "Error")
		}
	}
	if rootFlag != "" {
		_, err :=collection.UpdateMany(context.TODO(), 
			bson.D{}, 
			bson.D{
				{"$set", bson.D{
					{"Machines." + machines + ".RootFlag", rootFlag},
				}},
			},)
		if err != nil {
		    fmt.Fprintf(response, "Error")
		}
	}
	http.Redirect(response, request, "/appleHeadquarter", 302)
}

// Function for adding a VM to the Table
func AddVm(response http.ResponseWriter, request *http.Request) {
    request.ParseForm()
    machines := request.FormValue("machine")
    userFlag := request.FormValue("user")
    rootFlag := request.FormValue("root")
    
	collection, err := repos.GetDBCollection(1)
	if err != nil {
		fmt.Println(response, "error")
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
				{"$push", bson.M{"Machines":  
					    bson.M {
					        "ID_Machine": machineID,
					        "SolvedUser": false,
					        "SolvedRoot": false,
					        "UserFlag": userFlag,
					        "RootFlag": rootFlag,
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
    	Name: "cookie",
    	Value: user,
    	Path: "/",
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
func GetGroup(user string) int{
	groupA := strings.Split(user, "?")[1]
	groupB := strings.Split(groupA, "&")[0]
	group,_ := strconv.Atoi(groupB)

	return group
}
 
func GetUser(request *http.Request) (string, error) {
	user := ""
	// Check if there is an active Cookie
    if cookie, err := request.Cookie("cookie"); err == nil {
        // Check if the Cookie is in a valid format, valid format: userName?groupID&Hash, ex.: user0?1&"Hash"
        // Usernames must be [a-zA-Z0-9:]
        if match,_ :=regexp.MatchString(`\w+\?\d&\S+`, cookie.Value); match == true {
            cookieValue := cookie.Value
    	    cookieVal := strings.Split(cookieValue, "&")[0]
    	    values := strings.Split(cookieVal, "?")
    	    cookieHash := strings.Split(cookieValue, "&")[1]
            err = bcrypt.CompareHashAndPassword([]byte(cookieHash), []byte(values[0] + values[1]))
	        if err != nil {
	            return "", errors.New("Wrong or Modified Cookie")
	        }
            user = cookieValue
        } else {
            return "", errors.New("Wrong or Modified Cookie")
        }
        
    }
    return user, nil
}
