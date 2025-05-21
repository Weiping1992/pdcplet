package cache

import "fmt"

type VmiStatus int

// TODO: 考虑使用VmiStatus记录Vmi是否有CreateTask
const (
	VmiStatusNotReady VmiStatus = iota
	VmiStatusReady
)

type vmiStatusCache struct {
	cacheMap  map[string]VmiStatus
	taskIdMap map[string]int
}

func NewVmiStatusCache() *vmiStatusCache {
	return &vmiStatusCache{
		cacheMap:  make(map[string]VmiStatus),
		taskIdMap: make(map[string]int),
	}
}

// FIXME: 解决多个Update事件中任务会被重复创建的问题。
func (c *vmiStatusCache) Update(vmiName string, status VmiStatus) (isStatusChanged bool) {
	lastStatus, exist := c.cacheMap[vmiName]
	if !exist {
		c.cacheMap[vmiName] = status
		return status == VmiStatusReady
	}
	changed := lastStatus != status
	c.cacheMap[vmiName] = status
	return changed
}

func (c *vmiStatusCache) MarkDelete(vmiName string) {
	delete(c.cacheMap, vmiName)
	//delete(c.taskIdMap, vmiName)
}

func (c *vmiStatusCache) SetTaskId(vmiName string, taskId int) error {
	c.taskIdMap[vmiName] = taskId
	fmt.Printf("%v\n", c.taskIdMap)
	return nil
}

func (c *vmiStatusCache) GetTaskId(vmiName string) (taskId int, err error) {
	fmt.Printf("%v\n", c.taskIdMap)
	taskId, exists := c.taskIdMap[vmiName]
	if !exists {
		return -1, fmt.Errorf("no VmiName(%s) in cache", vmiName)
	}
	return taskId, nil
}

func (c *vmiStatusCache) DeleteDone(vmiName string) {
	delete(c.cacheMap, vmiName)
	delete(c.taskIdMap, vmiName)
}
