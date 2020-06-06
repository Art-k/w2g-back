package include

import (
	"github.com/jinzhu/gorm"
	guuid "github.com/satori/go.uuid"
	"log"
	"time"
)

type Model struct {
	ID        string `gorm:"primary_key"`
	CreatedAt time.Time
	CreatedBy string
	UpdatedAt time.Time
	UpdatedBy string
	DeletedAt *time.Time
	// DeletedBy string
}

func (base *Model) BeforeCreate(scope *gorm.Scope) error {
	// uuID, err := uuid.NewRandom()
	// if err != nil {
	// 	return err
	// }
	return scope.SetColumn("id", GetHash())
}

func GetHash() string {
	id, _ := guuid.NewV4()
	return id.String()
}

func DeleteRecordById(id string, object interface{}) bool {

	if result := Db.Where("id = ?", id).Delete(object); result.Error != nil {
		log.Println(result.Error)
		return false
	}
	return true

}
