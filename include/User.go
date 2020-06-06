package include

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"time"
)

type PostUser struct {
	Name         string
	FullName     string
	Email        string
	Role         string
	UserTimeZone string
}

type PatchUser struct {
	Name         string
	FullName     string
	Email        string
	Role         string
	UserTimeZone string
}

type User struct {
	Model
	Name         string `gorm:"type:varchar(100);unique_index"`
	LinkToAvatar string
	FullName     string
	Email        string
	Role         string
	Salt         string `json:"-"`
	Hash         string `json:"-"`
	SetPass      string `json:"-"`
	Enabled      bool
	Active       bool `gorm:"default:'false'"`
	PwdChanged   time.Time
	UserTimeZone string
}

func GetUserByID(id string) (user User, err error) {

	var foundUser User

	Db.
		Where("id = ?", id).
		Where("enabled = ?", true).
		Where("active = ?", true).
		Find(&foundUser)

	if foundUser.ID == "" {
		return foundUser, errors.New("[GetUserByID] User not found")
	} else {
		return foundUser, nil
	}

}

func APIUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":

		currentUser, err := GetUserByID(r.Context().Value("user").(string))
		if err != nil {
			ResponseInternalServerError(w)
			return
		}

		var recs []User
		if currentUser.Role == "ADMIN" {

			Db.Find(&recs)
			response := PrepareHTTPResponse(recs, len(recs), len(recs))
			ResponseOK(w, response)
			return

		} else {

			var responceRecs []User
			var tmpRecs []User
			Db.Where("id = ?", currentUser.ID).Find(&tmpRecs)
			// Info About Me
			for _, rec := range tmpRecs {
				responceRecs = append(responceRecs, rec)
			}

			response := PrepareHTTPResponse(responceRecs, len(responceRecs), len(responceRecs))
			ResponseOK(w, response)
			return

		}

	default:

		ResponseUnknown(w, "Method is not allowed")

	}
}

func APIUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "GET":

		var user User
		if result := Db.Where("id = ?", params["id"]).Find(&user); result.Error != nil {
			ResponseNotFound(w)
			return
		}

		response, err := json.Marshal(user)
		if err != nil {
			log.Println(err)
			ResponseInternalServerError(w)
			return
		}

		ResponseOK(w, response)

	case "PATCH":

		ID := params["id"]
		var newData PatchUser
		var DBData User
		var history History
		var ChangesMade bool

		CurrentUser, err := GetUserByID(r.Context().Value("user").(string))
		if err != nil {
			ResponseInternalServerError(w)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ResponseBadRequest(w, err, "")
			return
		}

		err = json.Unmarshal(body, &newData)
		if err != nil {
			ResponseBadRequest(w, err, "")
			return
		}

		err = Db.Where("id = ?", ID).Find(&DBData).Error
		if err != nil {
			ResponseBadRequest(w, err, "")
			return
		}

		historyObj, _ := json.Marshal(&DBData)

		// ========== DATA BLOCK ======================
		if DBData.Name != newData.Name {
			DBData.Name = newData.Name
			ChangesMade = true
		}
		if DBData.Email != newData.Email {
			DBData.Email = newData.Email
			ChangesMade = true
		}
		if DBData.FullName != newData.FullName {
			DBData.FullName = newData.FullName
			ChangesMade = true
		}
		if DBData.Role != newData.Role {
			if CurrentUser.Role == "ADMIN" {
				DBData.Role = newData.Role
				ChangesMade = true
			} else {
				ResponseBadRequest(w, nil, "")
				return
			}
		}
		if DBData.UserTimeZone != newData.UserTimeZone {
			DBData.UserTimeZone = newData.UserTimeZone
			ChangesMade = true
		}

		// ============ RESPONSE BLOCK ================
		if ChangesMade {
			Db.Model(&User{}).Update(&DBData)

			history.BeforeUpdate = string(historyObj)

			response, err := json.Marshal(&DBData)
			if err != nil {
				ResponseInternalServerError(w)
				return
			}

			history.AfterUpdate = string(response)
			history.ObjectId = ID
			history.Object = reflect.TypeOf(&User{}).String()
			history.CommandBody = string(body)
			history.MadeByUser = CurrentUser.ID
			history.MadeByFullName = CurrentUser.FullName

			Db.Create(&history)

			ResponseOK(w, response)

		}
		return

	case "DELETE":

		if !DeleteRecordById(params["id"], &User{}) {
			ResponseNotFound(w)
		}
		ResponseNoContent(w)

		//if result := Db.Where("id = ?", params["id"]).Delete(&User{}); result.Error!=nil{
		//	ResponseNotFound(w)
		//	return
		//}
		//ResponseNoContent(w)

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APIAddUser(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var rec PostUser
		err = json.Unmarshal(body, &rec)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		newRec := &User{
			Name:         rec.Name,
			FullName:     rec.FullName,
			Email:        rec.Email,
			Role:         "USER",
			UserTimeZone: rec.UserTimeZone,
		}
		newRec.CreatedBy = r.Context().Value("user").(string)

		if result := Db.Create(&newRec); result.Error != nil {
			log.Println(result.Error)
			ResponseInternalServerError(w)
			return
		}

		response, err := json.Marshal(&newRec)
		if err != nil {
			log.Println(err)
			ResponseInternalServerError(w)
			return
		}

		ResponseOK(w, response)
		return

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}
