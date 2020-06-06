package include

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type PostReminder struct {
	Remind          string
	ReminderAboutId string //Object ID
	ReminderAbout   string //Object Type
	ReminderAt      int64  //Unix Time
	ReminderWay     string //
}

type Reminder struct {
	Model
	Remind          string
	ReminderAboutId string //Object ID
	ReminderAbout   string //Object Type
	ReminderAt      int64  //Unix Time
	ReminderWay     string //
	ReminderSent    bool   `gorm:"default:'false'"`
	ReminderSentAt  time.Time
}

func DoRemind(user User, reminder Reminder, task Task, template string) {

	var history History
	var Done bool

	switch reminder.ReminderWay {
	case "TELEGRAM":

		var telegramSettings TelegramSetting
		Db.Where("user = ?", user.ID).Where("enabled = ?", true).Find(&telegramSettings)
		if telegramSettings.ID == "" {
			log.Println("Error, Telegram is not enabled for this user")
		}

		Done = SendTelegramMessage("[REMINDER] "+task.Subject, telegramSettings.TeleBotID, telegramSettings.TeleBotChannel)
		if Done {

			beforeUpdate, _ := json.Marshal(reminder)
			reminder.ReminderSent = true
			reminder.ReminderSentAt = time.Now()
			Db.Model(&Reminder{}).Update(&reminder)
			afterUpdate, _ := json.Marshal(reminder)
			history.BeforeUpdate = string(beforeUpdate)
			history.AfterUpdate = string(afterUpdate)

		}
	}

	if Done {
		history.Object = "REMINDER"
		history.ObjectId = reminder.ID
		history.MadeByFullName = "SYSTEM"
		history.MadeByUser = "a"

		Db.Create(&history)
	}

}

func DoScheduledReminds(t time.Time) {
	log.Println("Tick, reminder")

	currentTimeStamp := int64(time.Now().Unix())
	log.Println(currentTimeStamp)
	go func(timestamp int64) {

		var listOfReminders []Reminder
		Db.
			Where("reminder_at >= ?", currentTimeStamp).
			Where("reminder_at <=?", (currentTimeStamp+60)).
			Where("reminder_sent = ?", false).
			Find(&listOfReminders)

		for _, reminder := range listOfReminders {

			user, err := GetUserByID(reminder.Remind)
			if err != nil {
				log.Println(err)
				continue
			}

			template, err := GetReminderTemplate(reminder, user)
			if err != nil {
				log.Println(err)
				continue
			}

			switch reminder.ReminderAbout {
			case "TASK":

				task, err := GetActiveTaskByID(reminder.ReminderAboutId)
				if err != nil {
					log.Println(err)
				}

				DoRemind(user, reminder, task, template)
			default:
				log.Println("[DoScheduledReminds] We dont know why we need to remind, About What !!!")
			}

		}

	}(currentTimeStamp)

}

func DoEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func APIReminders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var recs []Reminder
		Db.Find(&recs)
		response := PrepareHTTPResponse(recs, len(recs), len(recs))
		ResponseOK(w, response)
	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APIReminder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "GET":

		var rec Reminder
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

		if !DeleteRecordById(params["id"], &Reminder{}) {
			ResponseNotFound(w)
		}
		ResponseNoContent(w)

		//if result := Db.Where("id = ?", params["id"]).Delete(&Reminder{}); result.Error != nil {
		//	ResponseNotFound(w)
		//	return
		//}
		//
		//ResponseNoContent(w)

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}

func APIAddReminder(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		var rec PostReminder
		err = json.Unmarshal(body, &rec)
		if err != nil {
			log.Println(err)
			ResponseBadRequest(w, err, "")
			return
		}

		newRec := &Reminder{
			Remind:          r.Context().Value("user").(string),
			ReminderAboutId: rec.ReminderAboutId,
			ReminderAbout:   rec.ReminderAbout,
			ReminderAt:      rec.ReminderAt,
			ReminderWay:     rec.ReminderWay,
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
