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

func (c *vmiStatusCache) Update(vmiName string, status VmiStatus) (isStatusChanged bool, isReady bool) {
	lastStatus, exist := c.cacheMap[vmiName]
	if !exist {
		c.cacheMap[vmiName] = status
		fmt.Printf("no exsits before, status: %d, changed: %t\n", status, status == VmiStatusReady)
		return status == VmiStatusReady, status == VmiStatusReady
	}
	changed := lastStatus != status
	c.cacheMap[vmiName] = status
	fmt.Printf("exsits, status: %d, lastStatus: %d, changed: %t\n", status, lastStatus, changed)
	return changed, status == VmiStatusReady
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
