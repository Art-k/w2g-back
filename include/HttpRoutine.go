package include

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

const Port = "56565"

const Version = "0.2.2"

type anyType struct{ f1 string }

type HTTPResponse struct {
	Total int
	Found int
	Data  interface{}
}

type HTTPEmptyResponse struct {
	Total int
	Found int
	Data  []int
}

func HandleHTTP() {

	r := mux.NewRouter()
	r.Use(authMiddleware)
	r.Use(headerMiddleware)

	r.HandleFunc("/token", APIToken)
	r.HandleFunc("/me", APIMe)

	r.HandleFunc("/logout", APIlogOut)

	r.HandleFunc("/password/{id}", SetPassword)
	r.HandleFunc("/invite/{id}", Invite)

	r.HandleFunc("/users", APIUsers)
	r.HandleFunc("/user", APIAddUser)
	r.HandleFunc("/user/{id}", APIUser)
	r.HandleFunc("/objects", APIUserObjects)
	r.HandleFunc("/available-objects", APIUserAvailableObjects)

	r.HandleFunc("/circles", APITeams)
	r.HandleFunc("/circle", APIAddTeam)
	r.HandleFunc("/circle/{id}/join", APIJoinTeam)
	//r.HandleFunc("/circle/{id}", APICircle)

	r.HandleFunc("/tasks", APITasks)
	r.HandleFunc("/task", APIAddTask)
	r.HandleFunc("/task/{task_id}", APITask)
	r.HandleFunc("/task/{task_id}/history", APITaskHistory)
	r.HandleFunc("/task/{task_id}/objects", APITaskObjects)

	r.HandleFunc("/checklist", APIAddChecklist)                              // Add check list to task
	r.HandleFunc("/checklists/{task_id}", APITaskChecklists)                 // Update Checklists Order (within a task) & get checklists from task
	r.HandleFunc("/checklists/{task_id}/status", APIGetTaskChecklistsStatus) // Update Checklists Order (within a task) & get checklists from task
	r.HandleFunc("/checklist/{checklist_id}", APICrudCheckList)              // Update Checklist Order (internally)
	r.HandleFunc("/checklist/{checklist_id}/item", APIAddCheckListItem)
	r.HandleFunc("/checklist/{checklist_id}/item/{item_id}", APICheckListItem)

	r.HandleFunc("/estimates", APIEstimate) // Get my types of estimation (Hours, Days, Minutes, etc....

	r.HandleFunc("/telegram-settings", APITelegramSettings)
	r.HandleFunc("/telegram-setting", APIAddTelegramSetting)
	r.HandleFunc("/telegram-setting/{id}", APITelegramSetting)

	fmt.Printf("Starting Server to HANDLE w2g.tech back end\nPort : " + Port + "\nAPI revision :" + Version + "\n\n")
	if err := http.ListenAndServe(":"+Port, r); err != nil {
		log.Fatal(err)
	}

}

func PrepareHTTPResponse(data interface{}, total, found int) []byte {

	if found != 0 {

		var response HTTPResponse
		response.Data = data
		response.Found = found
		response.Total = 0
		responseByte, err := json.Marshal(response)
		if err != nil {
			log.Println(err)
			return nil
		}
		return responseByte

	} else {

		var response HTTPEmptyResponse
		response.Data = make([]int, 0)
		response.Found = 0
		response.Total = 0
		responseByte, err := json.Marshal(response)
		if err != nil {
			log.Println(err)
			return nil
		}
		return responseByte

	}

}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		setupResponse(&w, r)

		if r.Method != "OPTIONS" {

			if r.RequestURI == "/token" || strings.Contains(r.RequestURI, "/password/") || r.RequestURI == "/me" {

				msg := r.RequestURI + " the received request is /token we don't need to auth here, do next handler"
				log.Println(msg)
				next.ServeHTTP(w, r)

			} else {

				msg := r.RequestURI + " end point is not /token we need to auth here"
				log.Println(msg)

				Authorization := r.Header.Get("Authorization")
				isUser, cUser := IsLegalUser(Authorization)

				msg = cUser.Name + " if found, so we will do all things under that user"
				log.Println(msg)

				if isUser && cUser.Enabled {

					//route := r.URL.Path
					//method := r.Method

					msg = "'" + r.Method + "' user wants to use that method, we need to check permissions for him"
					log.Println(msg)

					//if !IfUserHasPermission(cUser, GetRoute(route), method) {
					//
					//	msg = cUser.Name + " doesn't have a permission to do that"
					//	log.Println(msg)
					//	fmt.Println(msg)
					//
					//	w.WriteHeader(http.StatusForbidden)
					//	n, _ := fmt.Fprintf(w, "{\"message\":\"Access Denided\"}")
					//	log.Println(n)
					//	return
					//}

					msg = cUser.Name + " has permission to do that, do a next handle"
					log.Println(msg)

					ctx := context.WithValue(r.Context(), "user", cUser.ID)
					r = r.WithContext(ctx)

					if len(r.RequestURI) >= 7 {
						if cUser.Role != "ADMIN" && r.RequestURI[0:7] == "/invite/" {
							log.Println(msg)
							ResponseForbidden(w, msg, "log_out")
							return
						}
					}

					next.ServeHTTP(w, r)

				} else {
					msg = Authorization + " not connected to any user"
					log.Println(msg)
					ResponseForbidden(w, msg, "log_out")
				}

			}
		} else {
			return
		}
	})
}

func headerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		FillAnswerHeader(w)
		OptionsAnswer(w)
		next.ServeHTTP(w, r)
	})
}

func FillAnswerHeader(w http.ResponseWriter) {
	w.Header().Set("content-type", "application/json")
}

func OptionsAnswer(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
}

func ResponseOK(w http.ResponseWriter, addedRecordString []byte) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	n, _ := fmt.Fprintf(w, string(addedRecordString))
	fmt.Println("Response was sent ", n, " bytes")
	return
}

func ResponseNoContent(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	n, _ := fmt.Fprintf(w, "")
	fmt.Println("Response was sent ", n, " bytes")
	return
}

func ResponseConflict(w http.ResponseWriter, msg string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusConflict)
	n, _ := fmt.Fprintf(w, "{\"message\":\""+msg+"\"}")
	fmt.Println("Response was sent ", n, " bytes")
	return
}

func ResponseBadRequest(w http.ResponseWriter, err error, message string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")
	errorString := "{\"error_message\":\"" + err.Error() + "\",\"message\":\"" + message + "\"}"
	http.Error(w, errorString, http.StatusBadRequest)
	return
}

func ResponseNotFound(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	n, _ := fmt.Fprintf(w, "")
	fmt.Println("Response was sent ", n, " bytes")
	return
}

func ResponseInternalServerError(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	n, _ := fmt.Fprintf(w, "")
	fmt.Println("Response was sent ", n, " bytes")
	return
}

func ResponseUnknown(w http.ResponseWriter, message string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")
	errorString := "{\"message\":\"" + message + "\"}"
	http.Error(w, errorString, http.StatusInternalServerError)
	return
}

func ResponseForbidden(w http.ResponseWriter, message string, frontEndAction string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	n, _ := fmt.Fprintf(w, "{\"message\":\""+message+"\", \"action\":\""+frontEndAction+"\"}")
	fmt.Println("Response was sent ", n, " bytes")
	return
}
