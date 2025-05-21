package cache

type VmiStatus int

const (
	VmiStatusNotReady VmiStatus = iota
	VmiStatusReady
)

type vmiStatusCache struct {
	cacheMap map[string]VmiStatus
}

func NewVmiStatusCache() *vmiStatusCache {
	return &vmiStatusCache{
		cacheMap: make(map[string]VmiStatus),
	}
}

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

func (c *vmiStatusCache) Delete(vmiName string) {
	delete(c.cacheMap, vmiName)
}
