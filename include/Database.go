package include

import "github.com/jinzhu/gorm"

var Db *gorm.DB
var Err error

const DbLogMode = true

func InitializeDatabase() {
	Db.AutoMigrate(
		&User{},
		&Task{},
		&CheckList{},
		&CheckListItem{},
		&History{},
		&Reminder{},
		&TelegramSetting{},
		&ReminderTemplate{},
		&DefaultReminderTemplate{},
		&Token{},
		&RefreshToken{},
		&Team{},
		&TeamMembers{},
		&TeamTasks{},
		&RequestJoinToTeam{},
		&PhysicalObject{},
		&PhysicalObjectsAssigment{},
	)
}
