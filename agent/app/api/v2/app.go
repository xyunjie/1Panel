package v2

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/appicon"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) SearchApp(c *gin.Context) {
	var req request.AppSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	list, err := appService.PageApp(c, req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithDataGzipped(c, list)
}

func (b *BaseApi) SyncApp(c *gin.Context) {
	var req dto.OperateWithTask
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := appService.GetAppUpdate()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	if !res.CanUpdate {
		if res.IsSyncing {
			helper.SuccessWithMsg(c, i18n.GetMsgByKey("AppStoreIsSyncing"))
		} else {
			helper.SuccessWithMsg(c, i18n.GetMsgByKey("AppStoreIsUpToDate"))
		}
		return
	}
	if err = appService.SyncAppListFromRemote(req.TaskID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) SyncLocalApp(c *gin.Context) {
	var req dto.OperateWithTask
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	go appService.SyncAppListFromLocal(req.TaskID)
	helper.Success(c)
}

func (b *BaseApi) GetApp(c *gin.Context) {
	appKey, err := helper.GetStrParamByKey(c, "key")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	appDTO, err := appService.GetApp(c, appKey)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, appDTO)
}

func (b *BaseApi) GetAppDetail(c *gin.Context) {
	appID, err := helper.GetIntParamByKey(c, "appId")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	version := c.Param("version")
	appType := c.Param("type")
	appDetailDTO, err := appService.GetAppDetail(appID, version, appType)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, appDetailDTO)
}

func (b *BaseApi) GetAppDetailByID(c *gin.Context) {
	appDetailID, err := helper.GetIntParamByKey(c, "id")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	appDetailDTO, err := appService.GetAppDetailByID(appDetailID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, appDetailDTO)
}

func (b *BaseApi) InstallApp(c *gin.Context) {
	var req request.AppInstallCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	install, err := appService.Install(req, true)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, install)
}

func (b *BaseApi) GetAppTags(c *gin.Context) {
	tags, err := appService.GetAppTags(c)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, tags)
}

func (b *BaseApi) GetAppListUpdate(c *gin.Context) {
	res, err := appService.GetAppUpdate()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

func (b *BaseApi) GetAppIcon(c *gin.Context) {
	appKey, err := helper.GetStrParamByKey(c, "key")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	iconBytes, _, etag, err := appService.GetAppIcon(appKey)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	if len(iconBytes) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.Header("Cache-Control", "public, max-age=2592000")
	if etag != "" {
		c.Header("ETag", etag)
		if c.GetHeader("If-None-Match") == etag {
			c.Status(http.StatusNotModified)
			return
		}
	}

	c.Data(http.StatusOK, appicon.ContentTypePNG, iconBytes)
}

func (b *BaseApi) GetAppDetailForNode(c *gin.Context) {
	appKey := c.Param("appKey")
	version := c.Param("version")
	appDetailDTO, err := appService.GetAppDetailByKey(appKey, version)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, appDetailDTO)
}
