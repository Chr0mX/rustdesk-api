package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"net/http"
	"os"
	"strings"
)

type Index struct {
}

func (i *Index) Index(c *gin.Context) {
	c.Redirect(302, "/_admin/")
}

// WebClient2Index serves web2.1/web/index.html with {{CUSTOM_CONFIG}} substituted.
// The placeholder is replaced with a base64-encoded JSON object that pre-configures
// default settings (api-server) for the Flutter client via the custom-config mechanism.
func (i *Index) WebClient2Index(c *gin.Context) {
	indexPath := global.Config.Gin.ResourcesPath + "/web2.1/web/index.html"
	content, err := os.ReadFile(indexPath)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	apiServer := global.Config.Rustdesk.ApiServer
	cfg := map[string]interface{}{
		"default-settings": map[string]interface{}{
			"api-server": apiServer,
		},
	}
	cfgJSON, _ := json.Marshal(cfg)
	cfgB64 := base64.StdEncoding.EncodeToString(cfgJSON)

	html := strings.ReplaceAll(string(content), "{{CUSTOM_CONFIG}}", cfgB64)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func (i *Index) ConfigJs(c *gin.Context) {
	apiServer := global.Config.Rustdesk.ApiServer
	magicQueryonline := global.Config.Rustdesk.WebclientMagicQueryonline
	tmp := fmt.Sprintf(`localStorage.setItem('api-server', '%v');
const ws2_prefix = 'wc-';
localStorage.setItem(ws2_prefix+'api-server', '%v');

window.webclient_magic_queryonline = %d;
window.ws_host = '%v';
`, apiServer, apiServer, magicQueryonline, global.Config.Rustdesk.WsHost)
	//	tmp := `
	//localStorage.setItem('api-server', "` + apiServer + `")
	//const ws2_prefix = 'wc-'
	//localStorage.setItem(ws2_prefix+'api-server', "` + apiServer + `")
	//
	//window.webclient_magic_queryonline = ` + magicQueryonline + ``

	c.Header("Content-Type", "application/javascript")
	c.String(200, tmp)
}
