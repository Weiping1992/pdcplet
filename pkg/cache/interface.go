package cache

type Cache interface {
	Update(vmiName string, status VmiStatus) (isStatusChanged bool, isReady bool)
	MarkDelete(vmiName string)
	DeleteDone(vmiName string)
	SetTaskId(vmiName string, taskId int) error
	GetTaskId(vmiName string) (int, error)
}
