package admin

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

type Keypair struct {
}

// Get 获取当前密钥对
// @Tags 密钥对
// @Summary 获取当前的id_ed25519密钥对
// @Description 获取当前的id_ed25519密钥对
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response{data=service.Keypair}
// @Failure 500 {object} response.Response
// @Router /admin/keypair [get]
// @Security token
func (k *Keypair) Get(c *gin.Context) {
	kp, err := service.AllService.KeypairService.GetKeypair()
	if err != nil {
		response.Fail(c, 101, err.Error())
		return
	}
	response.Success(c, kp)
}

// Reset 重置密钥对
// @Tags 密钥对
// @Summary 生成一个新的随机id_ed25519密钥对
// @Description 生成一个新的随机id_ed25519密钥对，旧密钥将被覆盖
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response{data=service.Keypair}
// @Failure 500 {object} response.Response
// @Router /admin/keypair [post]
// @Security token
func (k *Keypair) Reset(c *gin.Context) {
	kp, err := service.AllService.KeypairService.ResetKeypair()
	if err != nil {
		response.Fail(c, 101, keypairErrMsg(c, err))
		return
	}
	response.Success(c, kp)
}

// Update 设置自定义密钥对
// @Tags 密钥对
// @Summary 使用提供的私钥设置密钥对
// @Description 使用提供的私钥(base64)设置密钥对，公钥将自动派生
// @Accept  json
// @Produce  json
// @Param body body string true "私钥(base64)"
// @Success 200 {object} response.Response{data=service.Keypair}
// @Failure 500 {object} response.Response
// @Router /admin/keypair [put]
// @Security token
func (k *Keypair) Update(c *gin.Context) {
	var priKey string
	if err := c.ShouldBindJSON(&priKey); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if priKey == "" {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	kp, err := service.AllService.KeypairService.SetKeypair(priKey)
	if err != nil {
		response.Fail(c, 101, keypairErrMsg(c, err))
		return
	}
	response.Success(c, kp)
}

func keypairErrMsg(c *gin.Context, err error) string {
	if errors.Is(err, service.ErrInvalidPrivateKey) {
		return response.TranslateMsg(c, "ParamsError") + ": " + err.Error()
	}
	return err.Error()
}
