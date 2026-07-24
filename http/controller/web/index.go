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

// clearConfigScript wipes whatever ConfigJs previously set, for a visitor
// who's no longer authed (see ConfigJs). Mirrors every key the authed
// branch below sets - both the unprefixed and "wc-" prefixed forms.
const clearConfigScript = `localStorage.removeItem('api-server');
localStorage.removeItem('custom-rendezvous-server');
localStorage.removeItem('relay-server');
localStorage.removeItem('key');
const ws2_prefix = 'wc-';
localStorage.removeItem(ws2_prefix+'api-server');
localStorage.removeItem(ws2_prefix+'custom-rendezvous-server');
localStorage.removeItem(ws2_prefix+'relay-server');
localStorage.removeItem(ws2_prefix+'key');
`

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
	// This has to be revalidated on every load: it reflects whatever the
	// admin most recently set (id-server/relay-server, or the webclient
	// override), and it's also the one place that decides whether *this*
	// visitor is authed right now. A cached response would keep serving
	// stale values/an old auth decision after either changes.
	c.Header("Cache-Control", "no-store, must-revalidate")

	authed, _ := c.Get(middleware.WebclientAuthedKey)
	if authed != true {
		// Actively wipe any id-server/relay-server/api-server/key this
		// visitor had from before, rather than just leaving them be. Without
		// this, logging out of _admin (middleware.RevokeWebclientSession)
		// would revoke the session server-side, but a webclient tab that
		// still had the old values cached in localStorage would keep
		// showing them until something else overwrote them - so
		// deauthenticating wouldn't visibly do anything on that side.
		c.Header("Content-Type", "application/javascript")
		c.String(200, clearConfigScript)
		return
	}

	apiServer := global.Config.Rustdesk.EffectiveWebclientApiServer()
	idServer := global.Config.Rustdesk.EffectiveWebclientIdServer()
	relayServer := global.Config.Rustdesk.EffectiveWebclientRelayServer()
	key := global.Config.Rustdesk.Key
	magicQueryonline := global.Config.Rustdesk.WebclientMagicQueryonline
	tmp := fmt.Sprintf(`localStorage.setItem('api-server', %v);
localStorage.setItem('custom-rendezvous-server', %v);
localStorage.setItem('relay-server', %v);
localStorage.setItem('key', %v);
const ws2_prefix = 'wc-';
localStorage.setItem(ws2_prefix+'api-server', %v);
localStorage.setItem(ws2_prefix+'custom-rendezvous-server', %v);
localStorage.setItem(ws2_prefix+'relay-server', %v);
localStorage.setItem(ws2_prefix+'key', %v);

window.webclient_magic_queryonline = %d;
window.ws_host = '%v';
`, strconv.Quote(apiServer), strconv.Quote(idServer), strconv.Quote(relayServer), strconv.Quote(key),
		strconv.Quote(apiServer), strconv.Quote(idServer), strconv.Quote(relayServer), strconv.Quote(key),
		magicQueryonline, global.Config.Rustdesk.WsHost)

	c.Header("Content-Type", "application/javascript")
	c.String(200, tmp)
}
