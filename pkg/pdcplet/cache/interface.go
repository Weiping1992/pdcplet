package cache

type Cache interface {
	Update(vmiName string, status VmiStatus) (isStatusChanged bool)
	MarkTaskCreated(vmiName string) error
	MarkTaskClosed(vmiName string) error
	IsTaskCreated(vmiName string) (bool, error)
	IsTaskClosed(vmiName string) (bool, error)
	IsReady(vmiName string) (bool, error)
	Delete(vmiName string)
	SetTaskId(vmiName string, taskId int) error
	GetTaskId(vmiName string) (int, error)
}
