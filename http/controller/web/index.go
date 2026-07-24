package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"strconv"
)

type Index struct {
}

func (i *Index) Index(c *gin.Context) {
	c.Redirect(302, "/_admin/")
}

// ConfigJs seeds the values the bundled webclient (resources/web, the
// flutter_hbb web build) reads out of localStorage on load. Without
// custom-rendezvous-server/key set here, connection.ts falls back to
// its own hardcoded HOSTS list (rs-sg/rs-cn/rs-us.rustdesk.com).
func (i *Index) ConfigJs(c *gin.Context) {
	apiServer := global.Config.Rustdesk.ApiServer
	idServer := global.Config.Rustdesk.IdServer
	key := global.Config.Rustdesk.Key
	magicQueryonline := global.Config.Rustdesk.WebclientMagicQueryonline
	tmp := fmt.Sprintf(`localStorage.setItem('api-server', '%v');
localStorage.setItem('custom-rendezvous-server', %v);
localStorage.setItem('key', %v);
const ws2_prefix = 'wc-';
localStorage.setItem(ws2_prefix+'api-server', '%v');
localStorage.setItem(ws2_prefix+'custom-rendezvous-server', %v);
localStorage.setItem(ws2_prefix+'key', %v);

window.webclient_magic_queryonline = %d;
window.ws_host = '%v';
`, apiServer, strconv.Quote(idServer), strconv.Quote(key), apiServer, strconv.Quote(idServer), strconv.Quote(key), magicQueryonline, global.Config.Rustdesk.WsHost)

	c.Header("Content-Type", "application/javascript")
	c.String(200, tmp)
}
