package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/middleware"
	"strconv"
)

type Index struct {
}

func (i *Index) Index(c *gin.Context) {
	c.Redirect(302, "/_admin/")
}

// ConfigJs seeds the values the bundled webclient (resources/web) reads
// out of localStorage on load. It sets both the unprefixed keys (older
// flutter_hbb builds) and the "wc-" prefixed keys (the current build,
// which stores all of its settings under that prefix). Without
// custom-rendezvous-server/key set here, connection.ts falls back to
// its own hardcoded HOSTS list (rs-sg/rs-cn/rs-us.rustdesk.com).
//
// It only hands out the real id-server/relay-server/api-server/key to
// visitors middleware.WebclientAuth has already vetted (see router.go) -
// otherwise those values would be readable by anyone who can reach this
// URL, logged in or not, which is enough to abuse the rendezvous/relay
// server or impersonate it.
func (i *Index) ConfigJs(c *gin.Context) {
	authed, _ := c.Get(middleware.WebclientAuthedKey)
	if authed != true {
		c.Header("Content-Type", "application/javascript")
		c.String(200, "// not authenticated: no server config for you\n")
		return
	}

	apiServer := global.Config.Rustdesk.ApiServer
	idServer := global.Config.Rustdesk.EffectiveWebclientIdServer()
	relayServer := global.Config.Rustdesk.EffectiveWebclientRelayServer()
	key := global.Config.Rustdesk.Key
	magicQueryonline := global.Config.Rustdesk.WebclientMagicQueryonline
	tmp := fmt.Sprintf(`localStorage.setItem('api-server', '%v');
localStorage.setItem('custom-rendezvous-server', %v);
localStorage.setItem('relay-server', %v);
localStorage.setItem('key', %v);
const ws2_prefix = 'wc-';
localStorage.setItem(ws2_prefix+'api-server', '%v');
localStorage.setItem(ws2_prefix+'custom-rendezvous-server', %v);
localStorage.setItem(ws2_prefix+'relay-server', %v);
localStorage.setItem(ws2_prefix+'key', %v);

window.webclient_magic_queryonline = %d;
window.ws_host = '%v';
`, apiServer, strconv.Quote(idServer), strconv.Quote(relayServer), strconv.Quote(key),
		apiServer, strconv.Quote(idServer), strconv.Quote(relayServer), strconv.Quote(key),
		magicQueryonline, global.Config.Rustdesk.WsHost)

	c.Header("Content-Type", "application/javascript")
	c.String(200, tmp)
}
