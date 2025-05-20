package cache

import (
	kubevirtv1 "kubevirt.io/api/core/v1"
)

type Cache interface {
	Update(vmi *kubevirtv1.VirtualMachineInstance) bool
}

type vmiStatusCache struct {
	cacheMap map[string]string
}

func NewVmiStatusCache() Cache {
	return &vmiStatusCache{
		cacheMap: make(map[string]string),
	}
}

func (c *vmiStatusCache) Update(vmi *kubevirtv1.VirtualMachineInstance) bool {
	return false
}
