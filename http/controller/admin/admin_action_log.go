package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/request/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"gorm.io/gorm"
)

type AdminActionLog struct {
}

// List 列表
// @Tags 管理员操作日志
// @Summary 管理员操作日志列表
// @Description 管理员操作日志列表
// @Accept  json
// @Produce  json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Param operator_name query string false "操作人"
// @Param module query string false "模块"
// @Param success query string false "是否成功"
// @Success 200 {object} response.Response{data=model.AdminActionLogList}
// @Failure 500 {object} response.Response
// @Router /admin/admin_action_log/list [get]
// @Security token
func (a *AdminActionLog) List(c *gin.Context) {
	query := &admin.AdminActionLogQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	res := service.AllService.AdminActionLogService.List(query.Page, query.PageSize, func(tx *gorm.DB) {
		if query.OperatorName != "" {
			tx.Where("operator_name like ?", "%"+query.OperatorName+"%")
		}
		if query.Module != "" {
			tx.Where("module = ?", query.Module)
		}
		if query.Success == "1" {
			tx.Where("success = ?", true)
		} else if query.Success == "0" {
			tx.Where("success = ?", false)
		}
		tx.Order("id desc")
	})
	response.Success(c, res)
}

// Delete 删除
// @Tags 管理员操作日志
// @Summary 管理员操作日志删除
// @Description 管理员操作日志删除
// @Accept  json
// @Produce  json
// @Param body body model.AdminActionLog true "管理员操作日志信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/admin_action_log/delete [post]
// @Security token
func (a *AdminActionLog) Delete(c *gin.Context) {
	f := &model.AdminActionLog{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	id := f.Id
	errList := global.Validator.ValidVar(c, id, "required,gt=0")
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	l := service.AllService.AdminActionLogService.InfoById(f.Id)
	if l.Id > 0 {
		err := service.AllService.AdminActionLogService.Delete(l)
		if err == nil {
			response.Success(c, nil)
			return
		}
		response.Fail(c, 101, err.Error())
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
}

// BatchDelete 批量删除
// @Tags 管理员操作日志
// @Summary 管理员操作日志批量删除
// @Description 管理员操作日志批量删除
// @Accept  json
// @Produce  json
// @Param body body admin.AdminActionLogIds true "管理员操作日志"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/admin_action_log/batchDelete [post]
// @Security token
func (a *AdminActionLog) BatchDelete(c *gin.Context) {
	f := &admin.AdminActionLogIds{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if len(f.Ids) == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}

	err := service.AllService.AdminActionLogService.BatchDelete(f.Ids)
	if err == nil {
		response.Success(c, nil)
		return
	}
	response.Fail(c, 101, err.Error())
	return
}
