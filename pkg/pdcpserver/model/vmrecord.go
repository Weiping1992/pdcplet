package model

import (
	"gorm.io/gorm"
)

type VirtualMachineStatus string

const (
	StatusNotSet VirtualMachineStatus = ""
	CreateFailed VirtualMachineStatus = "CreateFailed"
	Created      VirtualMachineStatus = "Created"
	Starting     VirtualMachineStatus = "Starting"
	Running      VirtualMachineStatus = "Running"
	Paused       VirtualMachineStatus = "Paused"
	Stopped      VirtualMachineStatus = "Stopped"
	MarkDeleted  VirtualMachineStatus = "MarkDeleted"
	Ready        VirtualMachineStatus = "Ready"
	NotReady     VirtualMachineStatus = "NotReady"
)

type VirtualMachineRecord struct {
	gorm.Model
	ID     uint                 `gorm:"primaryKey;autoIncrement"`
	Name   string               `gorm:"not null"`
	CPU    int                  `gorm:"not null"`
	Memory string               `gorm:"not null"`
	Status VirtualMachineStatus `gorm:"not null"`
}
