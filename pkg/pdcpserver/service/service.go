package service

import (
	"errors"
	"log/slog"
	"pdcplet/pkg/kubevirt"
	"pdcplet/pkg/pdcpserver/database"
	"pdcplet/pkg/pdcpserver/model"

	"gorm.io/gorm"
)

type Service interface {
	CreateVM(req model.VMCreateRequest) error
	DeleteVM(req model.VMDeleteRequest) error
}

type service struct{}

func New() Service {
	return &service{}
}

func (s *service) CreateVM(req model.VMCreateRequest) error {
	vmr := model.VirtualMachineRecord{
		Name:   req.Name,
		CPU:    req.CPU,
		Memory: req.Memory,
		Status: model.StatusNotSet,
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.createKubeVirtVM(req); err != nil {
			return err
		}
		vmr.Status = model.Created
		if err := tx.Create(&vmr).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *service) createKubeVirtVM(req model.VMCreateRequest) error {
	vm := model.NewKubeVirtVM(req)
	err := kubevirt.CreateVM(vm)
	if err != nil {
		slog.Error("Failed to create Kubevirt VirtualMachine", "error", err, "VmName", vm.Name, "Namespace", vm.Namespace)
	}
	return err
}

func (s *service) DeleteVM(req model.VMDeleteRequest) error {

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.deleteKubeVirtVM(req); err != nil {
			return err
		}
		vmr, err := GetVirtualMachineRecordByName(req.Name)
		if err != nil {
			slog.Error("Failed to get VirtualMachineRecord", "error", err, "VmName", req.Name)
			return err
		}
		if vmr != nil {
			vmr.Status = model.MarkDeleted
			if err := tx.Model(&vmr).Update("Status", model.MarkDeleted).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *service) deleteKubeVirtVM(req model.VMDeleteRequest) error {
	err := kubevirt.DeleteVM(req.Name, req.Namespace)
	if err != nil {
		slog.Error("Failed to create Kubevirt VirtualMachine", "error", err, "VmName", req.Name, "Namespace", req.Namespace)
	}
	return err
}

func GetVirtualMachineRecordByName(name string) (*model.VirtualMachineRecord, error) {
	var vmr model.VirtualMachineRecord
	err := database.DB.First(&vmr, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if err != nil {
		return nil, nil
	}
	return &vmr, nil
}
