package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type AppInstallRouter struct {
}

func (a *AppInstallRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("apps/installed")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/preflight", baseApi.PreflightOfflineInstall)
		groupRouter.POST("/openresty/upload", baseApi.UploadOpenrestyImage)
		groupRouter.POST("/openresty/install", baseApi.InstallOpenresty)
		groupRouter.GET("/openresty/containers", baseApi.ListOpenrestyContainers)
		groupRouter.POST("/openresty/link", baseApi.LinkOpenrestyContainer)
		groupRouter.POST("/database/upload", baseApi.UploadDatabaseImage)
		groupRouter.POST("/database/install", baseApi.InstallDatabase)
		groupRouter.GET("/database/containers", baseApi.ListDatabaseContainers)
		groupRouter.POST("/database/link", baseApi.LinkDatabaseContainer)
	}
}
