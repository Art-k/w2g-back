package include

type LinkedTask struct {
	Model
	MasterSlave bool
	TaskOneId   string
	TaskTwoId   string
	WhichFirst  int
}
