package include

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type PostTeam struct {
	TeamName string
}

type Team struct {
	Model
	TeamName       string
	OwnerFullName  string `gorm:"default:''"`
	TeamVisibility string `gorm:"default:'private'"`
}

type RequestJoinToTeam struct {
	Model
	TeamID               string `gorm:"default:''"`
	RequestOwnerID       string `gorm:"default:''"`
	RequestOwnerFullName string `gorm:"default:''"`
	RequestUserID        string `gorm:"default:''"`
	RequestUserFullName  string `gorm:"default:''"`
	OwnerToUser          bool   `gorm:"default:true"`
}

type TeamMembers struct {
	Model
	TeamId string
	UserId string
}

type TeamTasks struct {
	Model
	TeamId string
	TaskId string
}

func APITeams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":

		currentUser, err := GetUserByID(r.Context().Value("user").(string))
		if err != nil {
			ResponseInternalServerError(w)
			return
		}

		if currentUser.Role == "ADMIN" {

			var recs []Team
			Db.Find(&recs)
			response := PrepareHTTPResponse(recs, len(recs), len(recs))
			ResponseOK(w, response)
			return

		} else {

			var responseRecs []Team
			var tmpRecs []Team
			Db.Where("created_by = ?", currentUser.ID).Find(&tmpRecs)
			// Info About Me
			for _, rec := range tmpRecs {
				responseRecs = append(responseRecs, rec)
			}

			response := PrepareHTTPResponse(responseRecs, len(responseRecs), len(responseRecs))
			ResponseOK(w, response)
			return

		}

	default:

		ResponseUnknown(w, "Method is not allowed")

	}
}

func APIJoinTeam(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APIAddTeam(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":

		currentUser, err := GetUserByID(r.Context().Value("user").(string))
		if err != nil {
			ResponseBadRequest(w, err, "")
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var rec PostTeam
		err = json.Unmarshal(body, &rec)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var existingRec Team
		Db.
			Where("circle_name = ?", rec.TeamName).
			Where("created_by = ?", currentUser.ID).
			Find(&existingRec)

		if existingRec.ID != "" {
			ResponseConflict(w, "You already have circle '"+rec.TeamName+"'")
			return
		}

		newRec := &Team{
			TeamName:      rec.TeamName,
			OwnerFullName: currentUser.FullName,
		}
		newRec.CreatedBy = currentUser.ID

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
