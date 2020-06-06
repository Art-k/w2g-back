package include

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

type PostPhysicalObject struct {
	Name        string
	Description string
	//Position    int
}

type PhysicalObject struct {
	Model
	Name        string
	Description string
	Position    int
	//Objects []PhysicalObjectsAssigment
}

type PhysicalObjectsAssigment struct {
	Model
	ObjectId           string
	PhysicalObjectName string
}

func APIUserObjects(w http.ResponseWriter, r *http.Request) {

	//params := mux.Vars(r)
	var user User
	user, err := GetUserByID(r.Context().Value("user").(string))
	if err != nil {
		ResponseInternalServerError(w)
	}

	switch r.Method {

	case "POST":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var rec PostPhysicalObject
		err = json.Unmarshal(body, &rec)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var exRec PhysicalObject
		Db.Where("name = ?", rec.Name).Where("created_by = ?", user.ID).Find(&exRec)
		if exRec.ID != "" {
			ResponseConflict(w, "Already Exists!")
			return
		}

		var newRec PhysicalObject
		newRec.Name = rec.Name
		newRec.Description = rec.Description
		newRec.CreatedBy = user.ID
		Db.Create(&newRec)

		response, err := json.Marshal(&newRec)
		if err != nil {
			log.Println()
			ResponseInternalServerError(w)
		}
		ResponseOK(w, response)
		return

	case "GET":

		var recs []PhysicalObject
		Db.Where("created_by = ?", user.ID).Order("position asc").Find(&recs)

		response := PrepareHTTPResponse(recs, len(recs), len(recs))
		ResponseOK(w, response)

		return

	default:
		ResponseUnknown(w, "Method is not allowed")

	}
}

func APIUserAvailableObjects(w http.ResponseWriter, r *http.Request) {

	//params := mux.Vars(r)
	var user User
	user, err := GetUserByID(r.Context().Value("user").(string))
	if err != nil {
		ResponseInternalServerError(w)
	}

	switch r.Method {

	case "GET":

		var resp []string

		var recs []PhysicalObject
		Db.Where("created_by = ?", user.ID).Order("position asc").Find(&recs)

		for _, rec := range recs {
			resp = append(resp, rec.Name)
		}

		response, err := json.Marshal(resp)
		if err != nil {
			log.Println()
			ResponseInternalServerError(w)
		}

		ResponseOK(w, response)

	default:
		ResponseUnknown(w, "Method is not allowed")

	}
}

func APITaskObjects(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	//var user User
	//user, err := GetUserByID(r.Context().Value("user").(string))
	//if err!=nil{
	//	ResponseInternalServerError(w)
	//}

	switch r.Method {
	case "GET":

		var resp []string

		var recs []PhysicalObjectsAssigment
		Db.Where("object_id = ?", params["task_id"]).Find(&recs)

		for _, rec := range recs {
			resp = append(resp, rec.PhysicalObjectName)
		}

		response, err := json.Marshal(resp)
		if err != nil {
			log.Println()
			ResponseInternalServerError(w)
		}

		ResponseOK(w, response)

	case "PATCH":

	default:
		ResponseUnknown(w, "Method is not allowed")
	}

}
