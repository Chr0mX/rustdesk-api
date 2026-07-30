package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/controller/web"
	"github.com/lejianwen/rustdesk-api/v2/http/middleware"
	"net/http"
	"strings"
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
		requireWebclientAuth := func(loginPage *web.Index) gin.HandlerFunc {
			return func(c *gin.Context) {
				authed, _ := c.Get(middleware.WebclientAuthedKey)
				if authed != true {
					loginPage.WebclientLogin(c)
					c.Abort()
				}
			}
		}

		wc := g.Group("/webclient")
		wc.Use(noReferrer, wcAuth, requireWebclientAuth(i))
		wc.StaticFS("/", http.Dir(global.Config.Gin.ResourcesPath+"/web"))

		// Serves the exact same compiled bundle at a second, independent
		// slug (Phase 0 of docs/WEBCLIENT_V2_REBUILD_PLAN.md), so it stays
		// reachable - on its own admin-controlled toggle - once /webclient
		// above gets repointed at a from-source replacement. The slug itself
		// is fixed at startup (it's a route registration), but whether it's
		// reachable at all is re-checked on every request, so flipping
		// app.webclient-legacy-enabled from admin > settings takes effect
		// immediately, no restart needed.
		legacyPath := strings.Trim(strings.TrimSpace(global.Config.App.WebclientLegacyPath), "/")
		if legacyPath == "" || legacyPath == "webclient" {
			// "webclient" itself would collide with the group registered
			// just above and panic gin at startup - fall back rather than
			// let a misconfigured path take the server down.
			legacyPath = "webclient-legacy"
		}
		requireLegacyEnabled := func(c *gin.Context) {
			if !global.Config.App.WebclientLegacyEnabled {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			c.Next()
		}
		// Its own Index instance (not the canonical `i`) so its login page
		// redirects back to this slug after a successful login instead of
		// always landing on /webclient/ - see web.Index.RedirectBase.
		iLegacy := &web.Index{RedirectBase: "/" + legacyPath + "/"}
		wcLegacy := g.Group("/" + legacyPath)
		wcLegacy.Use(requireLegacyEnabled, noReferrer, wcAuth, requireWebclientAuth(iLegacy))
		wcLegacy.StaticFS("/", http.Dir(global.Config.Gin.ResourcesPath+"/web"))
	}
	g.StaticFS("/_admin", http.Dir(global.Config.Gin.ResourcesPath+"/admin"))
}
