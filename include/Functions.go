package include

type Function struct {
	Model
	FunctionName string
}

type FunctionProvided struct {
	Model
	UserId     string
	FunctionId string
}

type FunctionNeeded struct {
	Model
	TaskId     string
	FunctionId string
}
