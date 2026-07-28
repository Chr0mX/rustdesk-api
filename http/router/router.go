package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/controller/web"
	"github.com/lejianwen/rustdesk-api/v2/http/middleware"
	"net/http"
)

func WebInit(g *gin.Engine) {
	i := &web.Index{}
	g.GET("/", i.Index)

	if global.Config.App.WebClient == 1 {
		// Shared by both routes below so a visitor who authenticates via
		// /webclient/?token=... or /webclient/?share_token=... is still
		// recognized on the follow-up GET of /webclient-config/index.js.
		wcAuth := middleware.WebclientAuth()

		// /webclient/?token=<access_token> carries a real, long-lived api
		// token in the URL - without this, that token would land in the
		// Referer header of any outbound request/subresource the page ever
		// makes, in addition to browser history and access logs.
		noReferrer := func(c *gin.Context) {
			c.Header("Referrer-Policy", "no-referrer")
			c.Next()
		}

		g.GET("/webclient-config/index.js", noReferrer, wcAuth, i.ConfigJs)
		// POST, not GET: it has a real side effect (session + token
		// revocation) and needs no prior auth (it must work for an already-
		// expired session too), so a plain GET would let a cross-site
		// <img src="/webclient-logout"> force-logout any visitor. Triggered
		// via fetch() by the "Logout" button ConfigJs injects into the
		// authenticated webclient page - not nested under /webclient/ since
		// that'd conflict with its *filepath wildcard below.
		g.POST("/webclient-logout", noReferrer, i.WebclientLogout)

		// Unauthenticated visitors get web.Index.WebclientLogin instead of
		// the bundled webclient - any enabled account can sign in there
		// (not just admins), which then redirects to /webclient/?token=...
		// and lands back here, now authed.
		requireWebclientAuth := func(c *gin.Context) {
			authed, _ := c.Get(middleware.WebclientAuthedKey)
			if authed != true {
				i.WebclientLogin(c)
				c.Abort()
			}
		}

		wc := g.Group("/webclient")
		wc.Use(noReferrer, wcAuth, requireWebclientAuth)
		wc.StaticFS("/", http.Dir(global.Config.Gin.ResourcesPath+"/web"))
	}
	g.StaticFS("/_admin", http.Dir(global.Config.Gin.ResourcesPath+"/admin"))
}
