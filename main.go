package main

import (
	inc "./include"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

func main() {

	err := godotenv.Load("parameters.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//inc.AdmPass = os.Getenv("ADMP")
	//if os.Getenv("STATE") == "dev" {
	//	src.DEV = true
	//}

	inc.Db, inc.Err = gorm.Open("sqlite3", "w2g.db")
	if inc.Err != nil {
		panic("failed to connect database")
	}
	defer inc.Db.Close()
	inc.Db.LogMode(inc.DbLogMode)

	f, err := os.OpenFile("w2g.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	go func() { inc.DoEvery(1*time.Minute, inc.DoScheduledReminds) }()

	log.SetOutput(f)

	inc.InitializeDatabase()

	inc.HandleHTTP()

}
