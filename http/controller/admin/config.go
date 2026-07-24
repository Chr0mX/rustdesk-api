package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/request/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"os"
	"strings"
)

type Config struct {
}

// ServerConfig RUSTDESK服务配置
// @Tags ADMIN
// @Summary RUSTDESK服务配置
// @Description 服务配置,给webclient提供api-server
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/config/server [get]
// @Security token
func (co *Config) ServerConfig(c *gin.Context) {
	cf := &response.ServerConfigResponse{
		IdServer:             global.Config.Rustdesk.IdServer,
		Key:                  global.Config.Rustdesk.Key,
		RelayServer:          global.Config.Rustdesk.RelayServer,
		ApiServer:            global.Config.Rustdesk.ApiServer,
		WebclientIdServer:    global.Config.Rustdesk.WebclientIdServer,
		WebclientRelayServer: global.Config.Rustdesk.WebclientRelayServer,
	}
	response.Success(c, cf)
}

// UpdateWebclientConfig forces (or, given blank values, un-forces) the
// id-server/relay-server the bundled webclient is handed via
// web.Index.ConfigJs, independent of what native clients get. Persisted to
// the config file so it survives a restart.
// @Tags ADMIN
// @Summary 强制设置webclient的ID/中转服务器
// @Description 强制设置webclient的ID/中转服务器,留空则不覆盖
// @Accept  json
// @Produce  json
// @Param body body admin.WebclientConfigForm true "webclient配置"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/config/webclient [post]
// @Security token
func (co *Config) UpdateWebclientConfig(c *gin.Context) {
	f := &admin.WebclientConfigForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}

	global.Config.Rustdesk.WebclientIdServer = f.WebclientIdServer
	global.Config.Rustdesk.WebclientRelayServer = f.WebclientRelayServer
	global.Viper.Set("rustdesk.webclient-id-server", f.WebclientIdServer)
	global.Viper.Set("rustdesk.webclient-relay-server", f.WebclientRelayServer)
	if err := global.Viper.WriteConfig(); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}

	response.Success(c, nil)
}

// AppConfig APP服务配置
// @Tags ADMIN
// @Summary APP服务配置
// @Description APP服务配置
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/config/app [get]
// @Security token
func (co *Config) AppConfig(c *gin.Context) {
	response.Success(c, &gin.H{
		"web_client": global.Config.App.WebClient,
	})
}

// AdminConfig ADMIN服务配置
// @Tags ADMIN
// @Summary ADMIN服务配置
// @Description ADMIN服务配置
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/config/admin [get]
// @Security token
func (co *Config) AdminConfig(c *gin.Context) {

	u := &model.User{}
	token := c.GetHeader("api-token")
	if token != "" {
		u, _ = service.AllService.UserService.InfoByAccessToken(token)
		if !service.AllService.UserService.CheckUserEnable(u) {
			u.Id = 0
		}
	}

	if u.Id == 0 {
		response.Success(c, &gin.H{
			"title": global.Config.Admin.Title,
		})
		return
	}

	hello := global.Config.Admin.Hello
	if hello == "" {
		helloFile := global.Config.Admin.HelloFile
		if helloFile != "" {
			b, err := os.ReadFile(helloFile)
			if err == nil && len(b) > 0 {
				hello = string(b)
			}
		}
	}

	//replace {{username}} to username
	hello = strings.Replace(hello, "{{username}}", u.Username, -1)
	response.Success(c, &gin.H{
		"title": global.Config.Admin.Title,
		"hello": hello,
	})
}
