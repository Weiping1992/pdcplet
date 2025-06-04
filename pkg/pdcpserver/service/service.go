package service

import (
	"log/slog"
	"pdcplet/pkg/kubevirt"
	"pdcplet/pkg/pdcpserver/database"
	"pdcplet/pkg/pdcpserver/model"

	"gorm.io/gorm"
)

type Service interface {
	CreateVM(req model.VMCreateRequest) error
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
		if err := s.CreateKubeVirtVM(req); err != nil {
			return err
		}
		vmr.Status = model.Created
		if err := tx.Create(&vmr).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *service) CreateKubeVirtVM(req model.VMCreateRequest) error {

	vm := model.NewKubeVirtVM(req)
	err := kubevirt.ApplyVM(vm)
	if err != nil {
		slog.Error("Failed to create Kubevirt VirtualMachine", "error", err, "VmName", vm.Name, "Namespace", vm.Namespace)
	}
	return err
}
