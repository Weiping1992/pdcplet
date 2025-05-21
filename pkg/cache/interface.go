package cache

type Cache interface {
	Update(vmiName string, status VmiStatus) bool
	MarkDelete(vmiName string)
	DeleteDone(vmiName string)
	SetTaskId(vmiName string, taskId int) error
	GetTaskId(vmiName string) (int, error)
}
