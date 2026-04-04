package v2

import (
	"fmt"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) CheckAppInstalled(c *gin.Context) {
	var req request.AppInstalledInfo
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	checkData, err := appInstallService.CheckExist(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, checkData)
}

func (b *BaseApi) InstalledOp(c *gin.Context) {
	var req request.AppInstalledOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appInstallService.Operate(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) UploadOpenrestyImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}

	tmpDir := path.Join(global.Dir.TmpDir, "openresty")
	_ = os.MkdirAll(tmpDir, 0755)
	dst := path.Join(tmpDir, file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		helper.InternalServer(c, fmt.Errorf("failed to save uploaded file: %w", err))
		return
	}
	helper.SuccessWithData(c, dst)
}

func (b *BaseApi) InstallOpenresty(c *gin.Context) {
	var req service.OpenrestyInstallReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := openrestyService.Install(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) ListOpenrestyContainers(c *gin.Context) {
	containers, err := openrestyService.ListContainers()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, containers)
}

func (b *BaseApi) LinkOpenrestyContainer(c *gin.Context) {
	var req service.OpenrestyLinkReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := openrestyService.LinkContainer(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) UploadDatabaseImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	appKey := c.Query("appKey")
	if appKey == "" {
		appKey = "database"
	}

	tmpDir := path.Join(global.Dir.TmpDir, appKey)
	_ = os.MkdirAll(tmpDir, 0755)
	dst := path.Join(tmpDir, file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		helper.InternalServer(c, fmt.Errorf("failed to save uploaded file: %w", err))
		return
	}
	helper.SuccessWithData(c, dst)
}

func (b *BaseApi) InstallDatabase(c *gin.Context) {
	var req service.DatabaseInstallReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := databaseInstallService.Install(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) ListDatabaseContainers(c *gin.Context) {
	appKey := c.Query("appKey")
	if appKey == "" {
		helper.BadRequest(c, fmt.Errorf("appKey is required"))
		return
	}
	containers, err := databaseInstallService.ListContainers(appKey)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, containers)
}

func (b *BaseApi) LinkDatabaseContainer(c *gin.Context) {
	var req service.DatabaseLinkReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := databaseInstallService.LinkContainer(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
