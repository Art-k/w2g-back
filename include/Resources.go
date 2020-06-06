package include

type Resource struct {
	Model
	ResourceName string
}

type ResourceNeeded struct {
	Model
	TaskId     string
	ResourceId string
}

type ResourceProvided struct {
	Model
	UserId     string
	ResourceId string
}
