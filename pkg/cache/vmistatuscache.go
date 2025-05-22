package cache

import "fmt"

type VmiStatus int

// TODO: 考虑使用VmiStatus记录Vmi是否有CreateTask
const (
	VmiStatusNotReady VmiStatus = iota
	VmiStatusReady
)

type vmiStatusInfo struct {
	isReady       VmiStatus
	taskId        int
	isTaskCreated bool
	isTaskClosed  bool
}

func newVmiStatusInfo() *vmiStatusInfo {
	return &vmiStatusInfo{
		isReady:       VmiStatusNotReady,
		taskId:        -1,
		isTaskCreated: false,
		isTaskClosed:  false,
	}
}

type vmiStatusCache struct {
	cacheMap map[string]*vmiStatusInfo
}

func NewVmiStatusCache() *vmiStatusCache {
	return &vmiStatusCache{
		cacheMap: make(map[string]*vmiStatusInfo),
	}
}

func (c *vmiStatusCache) Update(vmiName string, status VmiStatus) (isStatusChanged bool) {
	lastStatus, exist := c.cacheMap[vmiName]
	if !exist {
		s := newVmiStatusInfo()
		s.isReady = status
		c.cacheMap[vmiName] = s
		//fmt.Printf("no exsits before, status: %d, changed: %t\n", status, status == VmiStatusReady)
		return status == VmiStatusReady
	}
	changed := lastStatus.isReady != status
	c.cacheMap[vmiName].isReady = status
	//fmt.Printf("exsits, status: %d, lastStatus: %d, changed: %t\n", status, lastStatus, changed)
	return changed
}

func (c *vmiStatusCache) MarkTaskCreated(vmiName string) error {
	vmiStatus, exist := c.cacheMap[vmiName]
	if !exist {
		return fmt.Errorf("no VmiName(%s) in cache", vmiName)
	}
	vmiStatus.isTaskCreated = true
	return nil
}

func (c *vmiStatusCache) MarkTaskClosed(vmiName string) error {
	vmiStatus, exist := c.cacheMap[vmiName]
	if !exist {
		return fmt.Errorf("no VmiName(%s) in cache", vmiName)
	}
	vmiStatus.isTaskClosed = true
	return nil
}

func (c *vmiStatusCache) IsTaskCreated(vmiName string) (bool, error) {
	vmiStatus, exist := c.cacheMap[vmiName]
	if !exist {
		return false, fmt.Errorf("no VmiName(%s) in cache", vmiName)
	}
	return vmiStatus.isTaskCreated, nil
}

func (c *vmiStatusCache) IsTaskClosed(vmiName string) (bool, error) {
	vmiStatus, exist := c.cacheMap[vmiName]
	if !exist {
		return false, fmt.Errorf("no VmiName(%s) in cache", vmiName)
	}
	return vmiStatus.isTaskClosed, nil
}

func (c *vmiStatusCache) IsReady(vmiName string) (bool, error) {
	vmiStatus, exist := c.cacheMap[vmiName]
	if !exist {
		return false, fmt.Errorf("no VmiName(%s) in cache", vmiName)
	}
	if vmiStatus.isReady == VmiStatusReady {
		return true, nil
	}
	return false, nil
}

func (c *vmiStatusCache) SetTaskId(vmiName string, taskId int) error {
	vmiStatus, exist := c.cacheMap[vmiName]
	if !exist {
		return fmt.Errorf("no VmiName(%s) in cache", vmiName)
	}
	vmiStatus.taskId = taskId
	return nil
}

func (c *vmiStatusCache) GetTaskId(vmiName string) (taskId int, err error) {
	vmiStatus, exist := c.cacheMap[vmiName]
	if !exist {
		return -1, fmt.Errorf("no VmiName(%s) in cache", vmiName)
	}
	if vmiStatus.taskId == -1 {
		return -1, fmt.Errorf("no taskId for VmiName(%s)", vmiName)
	}
	return vmiStatus.taskId, nil
}

func (c *vmiStatusCache) Delete(vmiName string) {
	delete(c.cacheMap, vmiName)
}
