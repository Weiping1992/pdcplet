package router

import (
	"pdcplet/pkg/pdcpserver/controller"

	"github.com/gin-gonic/gin"
)

func RegisterVMRoutes(r *gin.Engine) {
	vmGroup := r.Group("/pdcpserver/workload/vm")
	{
		vmGroup.POST("/create", controller.CreateVMHandler)
	}
}
