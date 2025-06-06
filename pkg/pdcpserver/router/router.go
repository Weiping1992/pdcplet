package router

import (
	"pdcplet/pkg/pdcpserver/controller"

	"github.com/gin-gonic/gin"
)

func RegisterVMRoutes(r *gin.Engine) {

	ctl := controller.NewController()

	vmGroup := r.Group("/pdcpserver/api/workload/vm")
	{
		vmGroup.GET("", ctl.GetVMSHandler)
		vmGroup.POST("/create", ctl.CreateVMHandler)
		vmGroup.POST("/delete", ctl.DeleteVMHandler)
	}
}
