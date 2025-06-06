package model

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

const (
	DEFAULT_NAMESPACE = "default"
)

var runningDefault bool = false

type VMCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	Namespace string `json:"namespace"`
	Memory    string `json:"memory" binding:"required"`
	CPU       int    `json:"cpu" binding:"min=1"`
	Image     string `json:"image" binding:"required"`
}

type VMCreateResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

func NewKubeVirtVM(req VMCreateRequest) *kubevirtv1.VirtualMachine {
	return &kubevirtv1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachine",
			APIVersion: "kubevirt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: kubevirtv1.VirtualMachineSpec{
			Running: &runningDefault,
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"kubevirt.io/vm": req.Name},
				},
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Devices: kubevirtv1.Devices{
							Disks: []kubevirtv1.Disk{{
								Name: "boot-disk",
								DiskDevice: kubevirtv1.DiskDevice{
									Disk: &kubevirtv1.DiskTarget{Bus: "virtio"},
								},
							}},
						},
						Resources: kubevirtv1.ResourceRequirements{
							Requests: map[v1.ResourceName]resource.Quantity{
								"memory": resource.MustParse(req.Memory),
							},
						},
						CPU: &kubevirtv1.CPU{Cores: uint32(req.CPU)},
					},
					Volumes: []kubevirtv1.Volume{{
						Name: "boot-disk",
						VolumeSource: kubevirtv1.VolumeSource{
							ContainerDisk: &kubevirtv1.ContainerDiskSource{
								Image: req.Image,
							},
						},
					}},
				},
			},
		},
	}
}

type VMDeleteRequest struct {
	Name      string `json:"name" binding:"required"`
	Namespace string `json:"namespace"`
}

type VMDeleteResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}
