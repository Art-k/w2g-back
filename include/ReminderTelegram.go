package include

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type PostTelegramSetting struct {
	Enabled        bool
	User           string
	TeleBotID      string
	TeleBotChannel string
}

type TelegramSetting struct {
	Model
	Enabled        bool
	User           string
	TeleBotID      string
	TeleBotChannel string
}

func SendTelegramMessage(msg, botId, botChannel string) bool {

	var url string
	// fmt.Println(msg)
	url = "https://api.telegram.org/bot" + botId + "/sendMessage?chat_id=" + botChannel + "&parse_mode=HTML&text="

	msg = strings.Replace(msg, " ", "+", -1)
	msg = strings.Replace(msg, "'", "%27", -1)
	msg = strings.Replace(msg, "\n", "%0A", -1)

	url = url + msg

	fmt.Println("\n" + url + "\n")
	response, err := http.Get(url)

	log.Println(response)

	if err != nil {
		return false
	}

	return true
}

func APITelegramSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var recs []TelegramSetting
		Db.Find(&recs)
		response := PrepareHTTPResponse(recs, len(recs), len(recs))
		ResponseOK(w, response)
	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APITelegramSetting(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "GET":

		var rec TelegramSetting
		if result := Db.Where("id = ?", params["id"]).Find(&rec); result.Error != nil {
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

		//log.Println("PATCH /user/" + params["id"])
		//buf, _ := ioutil.ReadAll(r.Body)

	case "DELETE":

		if !DeleteRecordById(params["id"], &TelegramSetting{}) {
			ResponseNotFound(w)
		}
		ResponseNoContent(w)

		//if result := Db.Where("id = ?", params["id"]).Delete(&TelegramSetting{}); result.Error != nil {
		//	ResponseNotFound(w)
		//	return
		//}
		//
		//ResponseNoContent(w)

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APIAddTelegramSetting(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var rec PostTelegramSetting
		err = json.Unmarshal(body, &rec)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		newRec := &TelegramSetting{
			Enabled:        true,
			TeleBotID:      rec.TeleBotID,
			TeleBotChannel: rec.TeleBotChannel,
		}
		newRec.CreatedBy = r.Context().Value("user").(string)
		newRec.User = r.Context().Value("user").(string)

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
