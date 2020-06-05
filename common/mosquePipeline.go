package common

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"pi-software/model"
	"pi-software/repos"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type MosquePipeline struct {
	Mosque           model.Mosque
	EditPrayers      bool // add/delete prayer, f.ex. add cuma prayer
	EditCapacity     bool // edit capacity for mosque, for women and men
	GetRegistrations bool // get users which want to attend choosen prayer, confirm visited
	GetDate          bool // get users registered for a choosen date
}

var mosquePipe MosquePipeline

// TODO: make mosqueadmin user in DB and as struct with mosquename / object
func MosqueHandler(response http.ResponseWriter, request *http.Request) {
	collection, _ := repos.GetDBCollection(1)
	var mosque model.Mosque
	// TODO: instead of hardcoded name, get name from user
	collection.FindOne(context.TODO(), bson.M{"Name": "ISV Mühlacker El-Aksa"}).Decode(&mosque)
	mosquePipe = MosquePipeline{mosque, false, false, false, true}
	t, _ := template.ParseFiles("templates/mosqueOverview.html")
	t.Execute(response, mosquePipe)
}

var datesMosque []Date

func GetRegistrations(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	dateG := request.PostFormValue("date")
	// For a specific Date
	if dateG != "" {
		var choosenDate Date
		var prayers []model.Prayer
		mosque := mosquePipe.Mosque
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
				prayers = append(prayers, prayer)
			}
		}
		choosenDate.Prayer = prayers
		tmpDate := make([]Date, 1)
		tmpDate[0] = choosenDate
		t, _ := template.ParseFiles("templates/getRegistrations.html")
		t.Execute(response, tmpDate)
	} else { // For all dates
		var prayers []model.Prayer
		for _, date := range mosquePipe.Mosque.Date {
			for _, prayer := range date.Prayer {
				if len(prayer.Users) > 0 {
					prayers = append(prayers, prayer)
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
		t, _ := template.ParseFiles("templates/getRegistrations.html")
		t.Execute(response, datesMosque)
	}
}

func PrintRegistrations(response http.ResponseWriter, request *http.Request) {
	t, err := template.ParseFiles("templates/activeRegistrations.html")
	fmt.Println(err)
	t.Execute(response, dates)
}
