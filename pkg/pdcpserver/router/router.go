package router

import (
	"pdcplet/pkg/pdcpserver/controller"

	"github.com/gin-gonic/gin"
)

func RegisterVMRoutes(r *gin.Engine) {

	ctl := controller.NewController()

	vmGroup := r.Group("/pdcpserver/workload/vm")
	{
		vmGroup.POST("/create", ctl.CreateVMHandler)
	}
}
