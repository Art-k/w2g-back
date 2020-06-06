package include

type History struct {
	Model
	TaskId         string
	Object         string
	ObjectId       string
	BeforeUpdate   string
	CommandType    string
	CommandBody    string
	AfterUpdate    string
	MadeByUser     string
	MadeByFullName string
}
