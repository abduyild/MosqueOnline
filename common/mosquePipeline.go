package common

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
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

var prayerB prayerBool

var mosquePipe MosquePipeline

// TODO: make mosqueadmin user in DB and as struct with mosquename / object
func MosqueHandler(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		collection, _ := repos.GetDBCollection(1)
		var mosque model.Mosque
		name, _ := GetPhoneFromCookie(request)
		collection.FindOne(context.TODO(), bson.M{"Name": name}).Decode(&mosque)
		date := time.Now()
		today := strconv.Itoa(date.Day()) + "." + strconv.Itoa(int(date.Month())) + "." + strconv.Itoa(date.Year())
		var prayers []model.Prayer
		var tmpPrayers []model.Prayer
		dates := mosque.Date
		for _, date := range dates {
			if today == strconv.Itoa(date.Date.Day())+"."+strconv.Itoa(int(date.Date.Month()))+"."+strconv.Itoa(date.Date.Year()) {
				tmpPrayers = date.Prayer
				break
			}
		}
		for _, prayer := range tmpPrayers {
			if len(prayer.Users) > 0 {
				prayers = append(prayers, prayer)
			}
		}
		var tmpDate Date
		tmpDate.Date = today
		tmpDate.Prayer = prayers
		mosquePipe = MosquePipeline{mosque, tmpDate, false, false, false}
		t, _ := template.ParseFiles("templates/mosqueIndex.html")
		t.Execute(response, mosquePipe)
	} else {
		accessError(response, request)
	}
}

func MosqueAction(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		action := request.URL.Query().Get("action")
		switch action {
		case "editprayers":
			mosquePipe.EditPrayers = true
		/* Currently not supported case "editcapacity":
		mosquePipe.EditCapacity = true*/
		case "getdate":
			mosquePipe.GetDate = true
		}
		t, _ := template.ParseFiles("templates/mosqueAction.html")
		t.Execute(response, mosquePipe)
	} else {
		accessError(response, request)
	}

}

func GetRegistrations(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
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
			tmpDate[0].Date = dateG
			if choosenDate.Date != "" {
				tmpDate[0] = choosenDate
			}
			t, _ := template.ParseFiles("templates/getRegistrations.html")
			t.Execute(response, tmpDate)
		} else { // For all dates
			var datesMosque []Date
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
	} else {
		accessError(response, request)
	}
}

func ConfirmVisitors(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		request.ParseForm()
		visitors := request.Form["visitor"]
		prayerName := request.URL.Query().Get("prayer")
		today := strings.Split(time.Now().String(), " ")[0]
		if prayerName != "" && len(visitors) == 0 {
			prayerI, _ := strconv.Atoi(prayerName)
			var prayers []model.Prayer
			var tmosque tempMosque
			var choosenDate Date
			mosque := mosquePipe.Mosque
			dates := mosque.Date
			for _, date := range dates {
				if today == strings.Split(date.Date.String(), " ")[0] {
					choosenDate.Date = strconv.Itoa(date.Date.Day()) + "." + strconv.Itoa(int(date.Date.Month())) + "." + strconv.Itoa(date.Date.Year())
					choosenDate.Prayer = date.Prayer
					break
				}
			}
			for _, prayer := range choosenDate.Prayer {
				if len(prayer.Users) > 0 && int(prayer.Name) == prayerI {
					prayers = append(prayers, prayer)
					break
				}
			}
			choosenDate.Prayer = prayers
			t, _ := template.ParseFiles("templates/confirmVisitors.html")
			tmosque.Name, _ = GetPhoneFromCookie(request)
			tmosque.Date = choosenDate
			t.Execute(response, tmosque)
		} else {
			data := strings.Split(request.URL.Query().Get("data"), "!")
			for _, phone := range visitors {
				today := strings.Split(time.Now().String(), " ")[0]
				index := 0
				for i, dateI := range mosquePipe.Mosque.Date {
					if today == strings.Split(dateI.Date.String(), " ")[0] {
						index = i
					}
				}
				collection, _ := repos.GetDBCollection(1)
				in, _ := strconv.Atoi(data[1])
				ind := strconv.Itoa(in - 1)
				result, err := collection.UpdateOne(context.TODO(),
					bson.M{"Name": data[0], "Date." + strconv.Itoa(index) + ".Prayer." + ind + ".Users.Phone": phone},
					bson.M{"$set": bson.M{"Date." + strconv.Itoa(index) + ".Prayer." + ind + ".Users.$.Attended": true}})
				fmt.Println(ind, result, err)
				// TODO: search in DB, but only in the users array of this prayer and set all occurences of phonenumbers to true
			}
		}
	} else {
		accessError(response, request)
	}
}

func GetPrayers(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		today := strings.Split(time.Now().String(), " ")[0]
		mosque := mosquePipe.Mosque
		var prayers []int
		for _, date := range mosque.Date {
			if today == strings.Split(date.Date.String(), " ")[0] {
				for _, prayer := range date.Prayer {
					prayers = append(prayers, int(prayer.Name))
				}
				break
			}
		}
		prayerB.Fajr = contains(prayers, 1)
		prayerB.Dhuhr = contains(prayers, 2)
		prayerB.Maghrib = contains(prayers, 3)
		prayerB.Asr = contains(prayers, 4)
		prayerB.Ishaa = contains(prayers, 5)
		t, _ := template.ParseFiles("templates/activePrayers.html")
		t.Execute(response, prayerB)
		prayerB = *new(prayerBool)
	} else {
		accessError(response, request)
	}
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func EditPrayers(response http.ResponseWriter, request *http.Request) {
	if adminLoggedin(response, request, "mosque") {
		tod := time.Now().Format(time.RFC3339)
		today, _ := time.Parse(time.RFC3339, tod)
		fromDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
		pathUrl, _ := url.ParseQuery(request.URL.RawQuery)
		path := pathUrl.Encode()
		prayer := strings.Split(path, "=")[0]
		value := strings.Split(path, "=")[1]
		collection, _ := repos.GetDBCollection(1)
		if value == "false" {
			var mosq Mosque
			collection.FindOne(context.TODO(), bson.M{"Name": mosquePipe.Mosque.Name}).Decode(&mosq)
			index := mosq.MaxCapM
			today := strings.Split(time.Now().String(), " ")[0]
			for i, dateI := range mosq.Date {
				if today == strings.Split(dateI.Date.String(), " ")[0] {
					index = i
				}
				if i >= index {
					for _, prayerIt := range dateI.Prayer {
						prayerIt.Available = false
					}
				}
			}

			prayerI, _ := strconv.Atoi(prayer)
			result, err := collection.UpdateMany(context.TODO(),
				bson.M{"$and": []bson.M{bson.M{"Name": mosquePipe.Mosque.Name}, bson.M{"Date.Date": bson.M{"$gt": fromDate}}}},
				bson.M{"$pull": bson.M{"Date.$.Prayer": bson.M{"Name": prayerI}}})
			fmt.Println(result, err)
			/*switch prayerI { KEINE AHNUNG WAS DAS HIER IST/WAR
			case 1:
				prayerB.Fajr = true
			case 2:
				prayerModel.Name = 2
			case 3:
				prayerModel.Name = 3
			case 4:
				prayerModel.Name = 4
			case 5:
				prayerModel.Name = 5
			case 6:
				prayerModel.Name = 6
			case 7:
				prayerModel.Name = 7
			}
			var nmosque model.Mosque
			collection.FindOne(context.TODO(), bson.M{"Name": "ISV Mühlacker El-Aksa"}).Decode(&nmosque)
			mosquePipe.Mosque = nmosque*/
		} else {
			var prayerModel model.Prayer
			switch prayer {
			case "1":
				prayerModel.Name = 1
			case "2":
				prayerModel.Name = 2
			case "3":
				prayerModel.Name = 3
			case "4":
				prayerModel.Name = 4
			case "5":
				prayerModel.Name = 5
			case "6":
				prayerModel.Name = 6
			case "7":
				prayerModel.Name = 7
			}
			prayerModel.CapacityMen = mosquePipe.Mosque.MaxCapM
			prayerModel.CapacityWomen = mosquePipe.Mosque.MaxCapW
			collection.UpdateMany(context.TODO(),
				bson.M{"Name": mosquePipe.Mosque.Name, "Date.Date": bson.M{"$gt": fromDate}},
				bson.M{"$push": bson.M{"Date.$.Prayer": prayerModel}})
		}
	} else {
		accessError(response, request)
	}
}
