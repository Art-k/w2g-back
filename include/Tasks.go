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

type PostTask struct {
	Subject string
}

type Task struct {
	Model
	Subject     string
	Description string

	Priority string // Nut Urgent & Not Important |

	IsCompleted bool

	CompletedAt *time.Time
	CompletedBy string

	Calendar *time.Time
	Deadline *time.Time

	Estimate   int
	EstimateIn string

	WhoIsDoing string

	//PreviousLinkedTask string
	//NextLinkedTask     string
	//LinkedTask         string
}

type PostChecklist struct {
	Name   string
	TaskId string
}

type CheckList struct {
	Model
	Name     string
	Status   string
	TaskId   string
	Position int
}

type PostChecklistItem struct {
	Name string
}

type CheckListItem struct {
	Model
	Name        string
	Checked     bool
	CheckListId string
	Position    int
}

type CheckListState struct {
	ID      string `json:"id"`
	Value   int    `json:"value"`
	Variant string `json:"variant"`
}

type CheckListsStateResponse struct {
	MaxValue int              `json:"max_value"`
	Data     []CheckListState `json:"data"`
}

func GetActiveTaskByID(taskId string) (task Task, err error) {

	Db.
		Where("id = ?", taskId).
		Where("is_completed = ?", false).
		Where("completed_at = ?", nil).
		Find(&task)

	if task.ID == "" {
		return task, errors.New("[GetActiveTaskByID] Task not found")
	} else {
		return task, nil
	}
}

func APIAddChecklist(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var rec PostChecklist
		err = json.Unmarshal(body, &rec)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var checkList CheckList
		Db.Where("task_id = ?", rec.TaskId).Where("name = ?", rec.Name).Find(&checkList)
		if checkList.ID != "" {
			ResponseConflict(w, "Already Exists!")
			return
		}

		currentUser, err := GetUserByID(r.Context().Value("user").(string))
		if err != nil {
			ResponseInternalServerError(w)
			return
		}

		var recs []CheckList
		Db.Where("task_id = ?", rec.TaskId).Order("position asc").Find(&recs)
		for ind, el := range recs {
			el.Position = ind + 1
			Db.Save(&el)
		}

		checkList.CreatedBy = currentUser.ID
		checkList.Name = rec.Name
		checkList.TaskId = rec.TaskId
		checkList.Position = 0

		Db.Create(&checkList)

		response, err := json.Marshal(&checkList)
		if err != nil {
			ResponseInternalServerError(w)
			return
		}
		ResponseOK(w, response)

		break

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APIGetTaskChecklistsStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "GET":

		var resp CheckListsStateResponse

		var recs []CheckList
		Db.Where("task_id = ?", params["task_id"]).Order("position asc").Find(&recs)
		for _, rec := range recs {

			var respRec CheckListState
			respRec.ID = rec.ID

			var items []CheckListItem
			Db.Where("check_list_id = ?", rec.ID).Find(&items)
			for _, item := range items {
				if item.Checked {
					respRec.Value++
				}
				resp.MaxValue++
			}
			if len(items) > 0 {
				percent := float32(respRec.Value) / float32(len(items))
				if percent <= 0.25 {
					respRec.Variant = "danger"
				}
				if percent > 0.25 && percent <= 0.75 {
					respRec.Variant = "warning"
				}
				if percent > 0.75 {
					respRec.Variant = "info"
				}
				if respRec.Value == len(items) {
					respRec.Variant = "success"
				}
				if respRec.Value == 0 {
					respRec.Variant = "danger"
					respRec.Value = len(items)
				}
			}

			resp.Data = append(resp.Data, respRec)

		}

		response, err := json.Marshal(&resp)
		if err != nil {
			ResponseInternalServerError(w)
			return
		}

		ResponseOK(w, response)
		return

	default:
		ResponseUnknown(w, "Method is not allowed")
	}

}

func APITaskChecklists(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "GET":

		var recs []CheckList
		Db.Where("task_id = ?", params["task_id"]).Order("position asc").Find(&recs)
		response := PrepareHTTPResponse(recs, len(recs), len(recs))
		ResponseOK(w, response)

	case "PUT":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var recs []CheckList
		err = json.Unmarshal(body, &recs)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		for ind, el := range recs {
			el.Position = ind
			Db.Save(&el)
		}

		Db.Where("task_id = ?", params["task_id"]).Order("position asc").Find(&recs)
		response := PrepareHTTPResponse(recs, len(recs), len(recs))
		ResponseOK(w, response)

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APICrudCheckList(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "GET":

		var recs []CheckListItem
		Db.Where("check_list_id = ?", params["checklist_id"]).Order("position asc").Find(&recs)
		response := PrepareHTTPResponse(recs, len(recs), len(recs))
		ResponseOK(w, response)

	case "PUT":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var recs []CheckListItem
		err = json.Unmarshal(body, &recs)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		for ind, el := range recs {
			el.Position = ind
			Db.Save(&el)
		}

		Db.Where("check_list_id = ?", params["checklist_id"]).Order("position asc").Find(&recs)
		response := PrepareHTTPResponse(recs, len(recs), len(recs))
		ResponseOK(w, response)

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func GetTaskIdByCheckListItem(itemId string) (taskId string) {

	var recItem CheckListItem
	Db.Where("id = ?", itemId).Find(&recItem)
	var recList CheckList
	Db.Where("id = ?", recItem.CheckListId).Find(&recList)

	return recList.TaskId
}

func APICheckListItem(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	var user User
	user, err := GetUserByID(r.Context().Value("user").(string))
	if err != nil {
		ResponseInternalServerError(w)
	}

	switch r.Method {
	case "GET":

		var item CheckListItem
		Db.Where("id = ?", params["item_id"]).Find(&item)
		response, err := json.Marshal(&item)
		if err != nil {
			log.Println(err)
			ResponseInternalServerError(w)
		}
		ResponseOK(w, response)
		return

	case "DELETE":

		var item CheckListItem

		Db.Where("id = ?", params["item_id"]).Find(&item)
		historyTextBytes, err := json.Marshal(&item)
		if err != nil {
			ResponseInternalServerError(w)
		}

		Db.Where("id = ?", params["item_id"]).Delete(&item)

		var history History
		history.TaskId = GetTaskIdByCheckListItem(params["item_id"])
		history.MadeByFullName = user.FullName
		history.CommandBody = "Delete Item ID: " + params["item_id"]
		history.Object = reflect.TypeOf(CheckListItem{}).String()
		history.ObjectId = params["item_id"]
		history.CommandType = "DELETE"
		history.BeforeUpdate = string(historyTextBytes)
		history.AfterUpdate = ""
		Db.Create(&history)

		ResponseNoContent(w)
		return

	case "PATCH":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var rec CheckListItem
		err = json.Unmarshal(body, &rec)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var exRec CheckListItem
		Db.Where("id = ?", params["item_id"]).Find(&exRec)
		if exRec.ID == "" {
			ResponseBadRequest(w, err, "Item Doesn't exist")
			return
		}

		beforeUpdate, err := json.Marshal(&exRec)
		if err != nil {
			ResponseInternalServerError(w)
		}

		exRec.Name = rec.Name
		exRec.Checked = rec.Checked
		exRec.UpdatedBy = user.ID

		if result := Db.Save(&exRec); result.Error != nil {
			ResponseBadRequest(w, result.Error, "")
			return
		}

		response, err := json.Marshal(&rec)

		var history History
		history.TaskId = GetTaskIdByCheckListItem(params["item_id"])
		history.MadeByFullName = user.FullName
		history.CommandBody = string(body)
		history.Object = reflect.TypeOf(CheckListItem{}).String()
		history.ObjectId = params["item_id"]
		history.CommandType = "PATCH"
		history.BeforeUpdate = string(beforeUpdate)
		history.AfterUpdate = string(response)
		Db.Create(&history)

		ResponseOK(w, response)
		return

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APIAddCheckListItem(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

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

		var rec PostChecklistItem
		err = json.Unmarshal(body, &rec)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var checkListItem CheckListItem
		Db.Where("check_list_id = ?", params["checklist_id"]).Where("name = ?", rec.Name).Find(&checkListItem)
		if checkListItem.ID != "" {
			ResponseConflict(w, "Already Exists!")
			return
		}

		currentUser, err := GetUserByID(r.Context().Value("user").(string))
		if err != nil {
			ResponseInternalServerError(w)
			return
		}

		var recs []CheckListItem
		Db.Where("check_list_id = ?", params["checklist_id"]).Order("position asc").Find(&recs)
		for ind, el := range recs {
			el.Position = ind + 1
			Db.Save(&el)
		}

		checkListItem.CreatedBy = currentUser.ID
		checkListItem.Name = rec.Name
		checkListItem.CheckListId = params["checklist_id"]

		Db.Create(&checkListItem)
		response, err := json.Marshal(&checkListItem)
		if err != nil {
			ResponseInternalServerError(w)
			return
		}

		var history History
		history.TaskId = GetTaskIdByCheckListItem(checkListItem.ID)
		history.MadeByFullName = user.FullName
		history.CommandBody = string(body)
		history.Object = reflect.TypeOf(CheckListItem{}).String()
		history.ObjectId = checkListItem.ID
		history.CommandType = "POST"
		history.BeforeUpdate = ""
		history.AfterUpdate = string(response)
		Db.Create(&history)

		ResponseOK(w, response)

		break

	default:
	}
}

func APITasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var recs []Task
		Db.Find(&recs)
		response := PrepareHTTPResponse(recs, len(recs), len(recs))
		ResponseOK(w, response)
	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APITaskHistory(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	switch r.Method {
	case "GET":

		var recs []History
		Db.Where("task_id = ?", params["task_id"]).Order("created_at desc").Find(&recs)
		response := PrepareHTTPResponse(recs, len(recs), len(recs))
		ResponseOK(w, response)

	default:
		ResponseUnknown(w, "Method is not allowed")
	}

}

// GET PUT DELETE
func APITask(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	var user User
	user, err := GetUserByID(r.Context().Value("user").(string))
	if err != nil {
		ResponseInternalServerError(w)
	}

	switch r.Method {
	case "GET":

		var rec Task
		if result := Db.Where("id = ?", params["task_id"]).Find(&rec); result.Error != nil {
			ResponseNotFound(w)
			return
		}
		response, err := json.Marshal(rec)
		if err != nil {
			log.Println(err)
			ResponseInternalServerError(w)
			return
		}
		ResponseOK(w, response)

	case "PATCH":

		var rec Task
		incomingData, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(incomingData, &rec)
		if err != nil {
			log.Println(err)
			ResponseInternalServerError(w)
		}

		var oldRec Task
		Db.Where("id = ?", rec.ID).Find(&oldRec)

		oldRec = rec
		Db.Save(&oldRec)

		response, err := json.Marshal(&oldRec)
		if err != nil {
			log.Println(err)
			ResponseInternalServerError(w)
		}
		ResponseOK(w, response)
		return

	case "PUT":

		id := params["task_id"]
		incomingData, _ := ioutil.ReadAll(r.Body)
		var incomingChanges interface{}
		var currentRec interface{}
		var rec Task

		err := json.Unmarshal(incomingData, &incomingChanges)
		if err != nil {
			log.Println(err)
			//ResponseInternalServerError(w)
		}
		mapChanges := incomingChanges.(map[string]interface{})

		db := Db.Model(&Task{}).Where("id = ?", id)

		db.Find(&rec)
		var history History
		var changesMade bool
		jsonBytes, _ := json.Marshal(&rec)

		err = json.Unmarshal(jsonBytes, &currentRec)
		mapRec := currentRec.(map[string]interface{})
		history.BeforeUpdate = string(jsonBytes)
		history.CommandBody = string(incomingData)

		for k, v := range mapRec {
			for key, value := range mapChanges {
				if k == key {
					fmt.Println("Found", reflect.TypeOf(v), "type")
					switch reflect.TypeOf(v) {
					case reflect.TypeOf(""):
						if v != value {
							db.Update(key, value)
							changesMade = true
						}
					}
				}
			}
		}

		db.Update("updated_by", user.ID)

		Db.Where("id = ?", id).Find(&rec)

		response, err := json.Marshal(&rec)
		if err != nil {
			log.Println(err)
			ResponseInternalServerError(w)
		}

		history.CreatedBy = r.Context().Value("user").(string)
		history.AfterUpdate = string(response)
		history.Object = "TASK"
		history.ObjectId = id
		history.TaskId = id

		if changesMade {
			Db.Create(&history)
			ResponseOK(w, response)
		} else {
			ResponseNoContent(w)
		}

	case "DELETE":

		if !DeleteRecordById(params["task_id"], &Task{}) {
			ResponseNotFound(w)
		}
		ResponseNoContent(w)

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APIAddTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var rec PostTask
		err = json.Unmarshal(body, &rec)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		newRec := &Task{
			Subject: rec.Subject,
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
