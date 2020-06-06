package include

import "errors"

type ReminderTemplate struct {
	Model
	ReminderForWho   string
	ReminderWay      string
	ReminderTemplate string
}

type DefaultReminderTemplate struct {
	Model
	ReminderWay      string
	ReminderTemplate string
}

func GetReminderTemplate(reminderType Reminder, userType User) (template string, err error) {

	var reminderTemplateType ReminderTemplate
	Db.
		Where("reminder_for_who = ?", userType.ID).
		Where("reminder_way = ?", reminderType.ReminderWay).
		Find(&reminderTemplateType)

	if reminderTemplateType.ID == "" {
		var defaultReminderTemplateType DefaultReminderTemplate
		Db.
			Where("reminder_way = ?", reminderType.ReminderWay).
			Find(&defaultReminderTemplateType)

		if defaultReminderTemplateType.ID == "" {
			return "", errors.New("[GetReminderTemplate] Reminder template not found")
		}
		return defaultReminderTemplateType.ReminderTemplate, nil

	} else {
		return reminderTemplateType.ReminderTemplate, nil
	}

}
