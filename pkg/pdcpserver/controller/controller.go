package controller

import (
	"net/http"

	"pdcplet/pkg/pdcpserver/model"
	"pdcplet/pkg/pdcpserver/service"

	"github.com/gin-gonic/gin"
)

func CreateVMHandler(c *gin.Context) {
	var req model.VMCreateRequest
	req.Namespace = model.DEFAULT_NAMESPACE // Set default namespace

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.CreateKubeVirtVM(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.VMCreateResponse{
		Message: "Created VM successfully",
		Name:    req.Name,
	})
}
