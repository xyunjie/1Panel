package v2

import (
	"fmt"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/psutil"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) SearchAppInstalled(c *gin.Context) {
	var req request.AppInstalledSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if req.All {
		list, err := appInstallService.SearchForWebsite(req)
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		helper.SuccessWithData(c, dto.PageResult{Items: list, Total: int64(len(list))})
	} else {
		total, list, err := appInstallService.Page(req)
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		helper.SuccessWithData(c, dto.PageResult{Items: list, Total: total})
	}
}

func (b *BaseApi) ListAppInstalled(c *gin.Context) {
	list, err := appInstallService.GetInstallList()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

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

func (b *BaseApi) PreflightOfflineInstall(c *gin.Context) {
	var req request.AppOfflineInstallPreflight
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	result := response.AppOfflineInstallPreflight{
		AppKey:      req.AppKey,
		InstallDir:  path.Join(global.Dir.AppInstallDir, req.AppKey),
		WebsiteDir:  path.Join(global.Dir.DataDir, "www"),
		DiskWarning: 90,
	}

	dockerStatus := dockerService.LoadDockerStatus()
	result.DockerReady = dockerStatus != nil && dockerStatus.IsExist && dockerStatus.IsActive
	if !result.DockerReady {
		if dockerStatus == nil || !dockerStatus.IsExist {
			result.DockerMessage = "Docker is not installed"
		} else {
			result.DockerMessage = "Docker service is unavailable"
		}
	} else {
		result.DockerMessage = "Docker is ready"
	}

	if req.AppKey == "openresty" {
		result.RequiredImage = "1panel/openresty:1.27.1.2-3-3-focal"
	} else if req.AppKey == "mysql" {
		result.RequiredImage = "mysql:8.0"
	} else if req.AppKey == "postgresql" {
		result.RequiredImage = "postgres:16"
	} else if req.AppKey == "redis" {
		result.RequiredImage = "redis:7"
	}

	result.ImageReady = true
	result.ImageMessage = "No offline image required"
	if result.RequiredImage != "" {
		result.ImageReady = false
		result.ImageMessage = fmt.Sprintf("Required image missing: %s", result.RequiredImage)
		if images, err := imageService.ListAll(); err == nil {
			for _, item := range images {
				for _, tag := range item.Tags {
					if tag == result.RequiredImage {
						result.ImageReady = true
						result.ImageMessage = fmt.Sprintf("Required image ready: %s", result.RequiredImage)
						break
					}
				}
				if result.ImageReady {
					break
				}
			}
		}
	}

	if err := os.MkdirAll(result.InstallDir, 0755); err != nil {
		result.InstallDirReady = false
		result.InstallDirMessage = err.Error()
	} else if stat, err := os.Stat(result.InstallDir); err != nil {
		result.InstallDirReady = false
		result.InstallDirMessage = err.Error()
	} else {
		result.InstallDirReady = stat.IsDir()
		if result.InstallDirReady {
			result.InstallDirMessage = "Install directory is writable"
		} else {
			result.InstallDirMessage = "Install path is not a directory"
		}
	}

	if req.AppKey == "openresty" {
		if err := os.MkdirAll(result.WebsiteDir, 0755); err != nil {
			result.WebsiteDirReady = false
			result.WebsiteDirMessage = err.Error()
		} else if stat, err := os.Stat(result.WebsiteDir); err != nil {
			result.WebsiteDirReady = false
			result.WebsiteDirMessage = err.Error()
		} else {
			result.WebsiteDirReady = stat.IsDir()
			if result.WebsiteDirReady {
				result.WebsiteDirMessage = "Website directory is writable"
			} else {
				result.WebsiteDirMessage = "Website path is not a directory"
			}
		}
	}

	usage, err := psutil.DISK.GetUsage(result.InstallDir, false)
	if err != nil {
		usage, err = psutil.DISK.GetUsage(global.Dir.AppInstallDir, false)
	}
	if err == nil && usage != nil {
		result.DiskUsedPercent = int(usage.UsedPercent)
		result.DiskAvailable = usage.Free
		result.DiskReady = result.DiskUsedPercent < result.DiskWarning
		if result.DiskReady {
			result.DiskMessage = fmt.Sprintf("Disk usage %d%%", result.DiskUsedPercent)
		} else {
			result.DiskMessage = fmt.Sprintf("Disk usage is high: %d%%", result.DiskUsedPercent)
		}
	} else {
		result.DiskMessage = "Disk usage unavailable"
	}

	result.Ready = result.DockerReady && result.ImageReady && result.InstallDirReady && result.DiskReady
	if req.AppKey == "openresty" {
		result.Ready = result.Ready && result.WebsiteDirReady
	}

	helper.SuccessWithData(c, result)
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

func (b *BaseApi) LoadPort(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	port, err := appInstallService.LoadPort(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, port)
}

func (b *BaseApi) LoadConnInfo(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	conn, err := appInstallService.LoadConnInfo(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, conn)
}

func (b *BaseApi) DeleteCheck(c *gin.Context) {
	appInstallId, err := helper.GetIntParamByKey(c, "appInstallId")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	checkData, err := appInstallService.DeleteCheck(appInstallId)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, checkData)
}

func (b *BaseApi) SyncInstalled(c *gin.Context) {
	if err := appInstallService.SyncAll(false); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) OperateInstalled(c *gin.Context) {
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

func (b *BaseApi) GetServices(c *gin.Context) {
	key := c.Param("key")
	services, err := appInstallService.GetServices(key)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, services)
}

func (b *BaseApi) GetUpdateVersions(c *gin.Context) {
	var req request.AppUpdateVersion
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	versions, err := appInstallService.GetUpdateVersions(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, versions)
}

func (b *BaseApi) ChangeAppPort(c *gin.Context) {
	var req request.PortUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appInstallService.ChangeAppPort(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) GetDefaultConfig(c *gin.Context) {
	var req dto.OperationWithNameAndType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	content, err := appInstallService.GetDefaultConfigByKey(req.Type, req.Name)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, content)
}

func (b *BaseApi) GetParams(c *gin.Context) {
	appInstallId, err := helper.GetIntParamByKey(c, "appInstallId")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	content, err := appInstallService.GetParams(appInstallId)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, content)
}

func (b *BaseApi) UpdateInstalled(c *gin.Context) {
	var req request.AppInstalledUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appInstallService.Update(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) UpdateAppConfig(c *gin.Context) {
	var req request.AppConfigUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appInstallService.UpdateAppConfig(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) GetAppInstallInfo(c *gin.Context) {
	appInstallId, err := helper.GetIntParamByKey(c, "appInstallId")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	info, err := appInstallService.GetAppInstallInfo(appInstallId)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, info)
}

func (b *BaseApi) UpdateAppInstallSort(c *gin.Context) {
	var req request.AppInstallSort
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appInstallService.UpdateSort(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
