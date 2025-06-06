package controller

import (
	"net/http"

	"pdcplet/pkg/pdcpserver/model"
	"pdcplet/pkg/pdcpserver/service"

	"github.com/gin-gonic/gin"
)

type Controller interface {
	CreateVMHandler(c *gin.Context)
	DeleteVMHandler(c *gin.Context)
	GetVMSHandler(c *gin.Context)
}

type defaultController struct {
	service service.Service
}

func NewController() Controller {
	return &defaultController{
		service: service.New(),
	}
}

func (controller *defaultController) CreateVMHandler(c *gin.Context) {
	var req model.VMCreateRequest
	req.Namespace = model.DEFAULT_NAMESPACE // Set default namespace

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := controller.service.CreateVM(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.VMCreateResponse{
		Message: "Created VM successfully",
		Name:    req.Name,
	})
}

func (controller *defaultController) DeleteVMHandler(c *gin.Context) {
	var req model.VMDeleteRequest
	req.Namespace = model.DEFAULT_NAMESPACE // Set default namespace

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := controller.service.DeleteVM(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.VMDeleteResponse{
		Message: "Delete VM successfully",
		Name:    req.Name,
	})
}

func (controller *defaultController) GetVMSHandler(c *gin.Context) {

}
