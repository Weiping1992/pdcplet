package vmiwatcher

import (
	kubevirtv1 "kubevirt.io/api/core/v1"
)

type OperateType int

const (
	CreateTaskOp OperateType = iota
	CloseTaskOp
)

type workqueueItem struct {
	vmi *kubevirtv1.VirtualMachineInstance
	op  OperateType
}
