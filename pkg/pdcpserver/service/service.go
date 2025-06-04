package service

import (
	"pdcplet/pkg/kubevirt"
	"pdcplet/pkg/pdcpserver/model"
)

func CreateKubeVirtVM(req model.VMCreateRequest) error {
	vm := model.NewKubeVirtVM(req)
	return kubevirt.ApplyVM(vm)
}
