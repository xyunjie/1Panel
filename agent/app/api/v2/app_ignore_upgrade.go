package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) ListAppIgnored(c *gin.Context) {
	res, err := appIgnoreUpgradeService.List()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

func (b *BaseApi) IgnoreAppUpgrade(c *gin.Context) {
	var req request.AppIgnoreUpgradeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appIgnoreUpgradeService.CreateAppIgnore(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) CancelIgnoreAppUpgrade(c *gin.Context) {
	var req request.ReqWithID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appIgnoreUpgradeService.Delete(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
