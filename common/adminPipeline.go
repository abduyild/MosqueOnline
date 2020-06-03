package common

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"pi-software/model"
	"pi-software/repos"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type AdminPipeline struct {
	AddMosque    bool
	EditMosque   bool
	DeleteMosque bool
	ShowMosque   bool
	Mosques      []model.Mosque
}

type edit struct {
	MosqueChoosen bool
	Mosques       []model.Mosque
}

var ed edit

var mosque model.Mosque

func AdminHandler(response http.ResponseWriter, request *http.Request) {
	adminPipe := AdminPipeline{false, true, false, false, getMosques(response, request)}
	t, _ := template.ParseFiles("templates/admin.html")
	/*if adminPipe.EditMosque {

	}*/
	t.Execute(response, adminPipe)
}

// Function for adding a VM to the Table
func AddMosque(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	if len(request.PostForm) > 0 {
		name := request.FormValue("name")
		plz, _ := strconv.Atoi(request.FormValue("plz"))
		street := request.FormValue("street")
		city := request.FormValue("city")
		cap_m, _ := strconv.Atoi(request.FormValue("cap-m"))
		cap_w, _ := strconv.Atoi(request.FormValue("cap-w"))

		collection, err := repos.GetDBCollection(1)
		t := check(response, request, err)
		if t != nil {
			t.Execute(response, errors.New(dbConnectionError))
			return
		}
		var date model.Date
		var dates []model.Date
		var prayer model.Prayer
		var prayers []model.Prayer
		prayer.CapacityMen = cap_m
		prayer.CapacityWomen = cap_w
		prayer.Users = []model.User{}
		for i := 1; i < 6; i++ {
			prayer.Name = model.PrayerName(i)
			prayers = append(prayers, prayer)
		}
		// TODO: statt 10 einfach 100 nemhen oder so
		for i := 0; i < 10; i++ {
			currentDate := time.Now().AddDate(0, 0, i).Format(time.RFC3339)
			date.Date, _ = time.Parse(time.RFC3339, currentDate)
			date.Prayer = prayers
			dates = append(dates, date)
		}
		mosque := *new(Mosque)
		mosque.Name = name
		mosque.PLZ = plz
		mosque.Street = street
		mosque.City = city
		mosque.Date = dates
		mosque.MaxCapM = cap_m
		mosque.MaxCapW = cap_w

		collection.InsertOne(context.TODO(), mosque)

		http.Redirect(response, request, "/admin", 302) // redirect back to Adminpage
	} else {
		t, _ := template.ParseFiles("templates/addMosque.html")
		t.Execute(response, nil)
	}
}

// tested getting old capacity and writing to it, but somehow didnt work
func editMosque(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	mosqueName := request.URL.Query().Get("mosque")
	fmt.Println(mosqueName, ed.MosqueChoosen)
	if !ed.MosqueChoosen && mosqueName != "" {
		fmt.Println(1)
		collection, _ := repos.GetDBCollection(1)
		collection.FindOne(context.TODO(),
			bson.D{
				{"Name", mosqueName},
			}).Decode(&mosque)
		ed.MosqueChoosen = true
		response.Write([]byte("<script>window.location.reload();</script>"))
	} else if ed.MosqueChoosen && len(request.PostForm) > 0 {
		fmt.Println(2)
		type tempMosque struct {
			CapacityMen   int
			CapacityWomen int
			UsedMen       int
			UsedWomen     int
		}
		var temp tempMosque
		usem := 0 // used for temporary result as we want to get macx. usage
		usew := 0
		usem1 := mosque.Date[0].Prayer[0].CapacityMen //used for finished result
		usew1 := mosque.Date[0].Prayer[0].CapacityWomen
		temp.CapacityMen = mosque.MaxCapM
		temp.CapacityWomen = mosque.MaxCapW
		for _, date := range mosque.Date {
			for _, prayer := range date.Prayer {
				usem = prayer.CapacityMen
				usew = prayer.CapacityWomen
				if usem <= usem1 {
					usem1 = mosque.MaxCapM - prayer.CapacityMen
				}
				if usew <= usew1 {
					usew1 = mosque.MaxCapW - prayer.CapacityWomen
				}

			}
		}
		temp.UsedMen = usem1
		temp.UsedWomen = usew1

		capm, err := strconv.Atoi(request.FormValue("cap-m"))
		if err != nil {
			capm = mosque.MaxCapM
		}
		capw, err := strconv.Atoi(request.FormValue("cap-w"))
		if err != nil {
			capw = mosque.MaxCapW
		}

		collection, _ := repos.GetDBCollection(1)
		collection.UpdateOne(context.TODO(), bson.D{{"Name", mosque.Name}}, bson.D{{"$set", bson.D{
			{"MaxCapM", capm},
			{"MaxCapW", capw},
		}}})
		mosque = *new(model.Mosque)
		ed.MosqueChoosen = false
	} else {
		fmt.Println(3)
		t, _ := template.ParseFiles("templates/editMosque.html")
		ed.Mosques = getMosques(response, request)
		t.Execute(response, ed)
	}
}

func EditMosque(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	if len(request.PostForm) > 0 {
		mosqueName := request.FormValue("edit-mosque")
		collection, _ := repos.GetDBCollection(1)
		collection.FindOne(context.TODO(),
			bson.D{
				{"Name", mosqueName},
			}).Decode(&mosque)
		type tempMosque struct {
			CapacityMen   int
			CapacityWomen int
			UsedMen       int
			UsedWomen     int
		}
		var temp tempMosque
		usem := 0 // used for temporary result as we want to get macx. usage
		usew := 0
		usem1 := mosque.Date[0].Prayer[0].CapacityMen //used for finished result
		usew1 := mosque.Date[0].Prayer[0].CapacityWomen
		temp.CapacityMen = mosque.MaxCapM
		temp.CapacityWomen = mosque.MaxCapW
		for _, date := range mosque.Date {
			for _, prayer := range date.Prayer {
				usem = prayer.CapacityMen
				usew = prayer.CapacityWomen
				if usem <= usem1 {
					usem1 = mosque.MaxCapM - prayer.CapacityMen
				}
				if usew <= usew1 {
					usew1 = mosque.MaxCapW - prayer.CapacityWomen
				}

			}
		}

		capm, err := strconv.Atoi(request.FormValue("cap-m"))
		if err != nil {
			capm = mosque.MaxCapM
		}
		capw, err := strconv.Atoi(request.FormValue("cap-w"))
		if err != nil {
			capw = mosque.MaxCapW
		}
		if capm < usem1 {
			t, _ := template.ParseFiles("templates/errorpage.html")
			http.Redirect(response, request, "/error", 402)
			t.Execute(response, errors.New("Es gibt Gebete die bereits mehr Anmeldungen haben als die angegebene Mindestkapazität! Mindestkapazität für Männer: "+strconv.Itoa(usem1)))
			response.Write([]byte(`<script>window.location.href = "/error"</script>`))
		}
		if capw < usew1 {
			t, _ := template.ParseFiles("templates/errorpage.html")
			http.Redirect(response, request, "/error", 402)
			t.Execute(response, errors.New("Es gibt Gebete die bereits mehr Anmeldungen haben als die angegebene Mindestkapazität! Mindestkapazität für Frauen: "+strconv.Itoa(usew1)))
			response.Write([]byte(`<script>window.location.href = "/error"</script>`))
		}

		temp.UsedMen = usem1
		temp.UsedWomen = usew1
		collection.UpdateOne(context.TODO(), bson.D{{"Name", mosque.Name}}, bson.D{{"$set", bson.D{
			{"MaxCapM", capm},
			{"MaxCapW", capw},
		}}})
		collection.UpdateMany(context.TODO(), bson.D{
			{"Name", mosque.Name},
			{"$or", bson.D{
				{"Date.$[].Prayer", bson.D{
					{"CapacityMen", bson.M{"$lt": capm}},
					{"CapacityWomen", bson.M{"$lt": capw}},
				}}}},
		},
			bson.M{"$set": bson.M{"Date.$[].Prayer": bson.M{"CapacityMen": capm, "CapacityWomen": capw}}})
		for _, date := range mosque.Date {
			for _, prayer := range date.Prayer {
				usem = prayer.CapacityMen
				usew = prayer.CapacityWomen
				if usem <= usem1 {
					usem1 = mosque.MaxCapM - prayer.CapacityMen
				}
				if usew <= usew1 {
					usew1 = mosque.MaxCapW - prayer.CapacityWomen
				}
			}
		}
		mosque = *new(model.Mosque)
		ed.MosqueChoosen = false
	} else {
		fmt.Println(3)
		t, _ := template.ParseFiles("templates/editMosque.html")
		ed.Mosques = getMosques(response, request)
		t.Execute(response, ed)
	}
}

func DeleteMosque(response http.ResponseWriter, request *http.Request) {
	mosque := request.URL.Query().Get("mosque")
	collection, err := repos.GetDBCollection(1)
	t := check(response, request, err)
	if t != nil {
		t.Execute(response, errors.New(dbConnectionError))
		return
	}
	collection.DeleteOne(context.TODO(), bson.M{"Name": mosque})

	collection, err = repos.GetDBCollection(0)
	t = check(response, request, err)
	if t != nil {
		t.Execute(response, errors.New(dbConnectionError))
		return
	}
	update := bson.D{{Key: "$pull", Value: bson.D{{Key: "RegisteredPrayers", Value: bson.D{{Key: "MosqueName", Value: mosque}}}}}}
	collection.UpdateMany(context.TODO(), bson.D{{}}, update)

}
