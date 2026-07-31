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
		// authenticated legacy webclient page - not nested under a
		// webclient path since that'd conflict with its *filepath
		// wildcard.
		g.POST("/webclient-logout", noReferrer, i.WebclientLogout)

		// Unauthenticated visitors get web.Index.WebclientLogin instead of
		// the bundled legacy webclient - any enabled account can sign in
		// there (not just admins), which then redirects back to this same
		// slug's ?token=... and lands back here, now authed. Only the
		// legacy bundle uses this: the from-source Vue webclient below has
		// its own client-side login page/router guard instead, so it isn't
		// gated server-side at all - unauthenticated visitors just get the
		// app shell, which shows its own Login.vue.
		requireWebclientAuth := func(loginPage *web.Index) gin.HandlerFunc {
			return func(c *gin.Context) {
				authed, _ := c.Get(middleware.WebclientAuthedKey)
				if authed != true {
					loginPage.WebclientLogin(c)
					c.Abort()
				}
			}
		}

		// The legacy (pre-Phase-4) bundled webclient, kept reachable at
		// its own admin-toggleable slug (Phase 0 of docs/
		// WEBCLIENT_V2_REBUILD_PLAN.md) now that /webclient/ itself
		// (below) serves the from-source Vue rebuild instead - not the
		// primary route anymore, but still fully functional for admins
		// who haven't cut over yet or want a fallback. The slug itself is
		// fixed at startup (it's a route registration), but whether it's
		// reachable at all is re-checked on every request, so flipping
		// app.webclient-legacy-enabled from admin > settings takes effect
		// immediately, no restart needed.
		legacyPath := strings.Trim(strings.TrimSpace(global.Config.App.WebclientLegacyPath), "/")
		if legacyPath == "" || legacyPath == "webclient" {
			// "webclient" itself would collide with the Vue webclient
			// routes registered below and panic gin at startup - fall
			// back rather than let a misconfigured path take the server
			// down.
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

		// The from-source Vue webclient (Phase 4 of docs/
		// WEBCLIENT_V2_REBUILD_PLAN.md, rustdesk-api-web) - the primary
		// /webclient/ route as of Phase 6's cutover, replacing the legacy
		// bundle above. rustdesk-api-web's vite.config.js builds
		// webclient.html as a second entry alongside _admin's own
		// index.html into the same dist/ output, which Rustdesk-Server-
		// Installer's update.sh already copies wholesale into
		// resources/admin - so webclient.html and its static/ assets sit
		// right next to _admin's own. No server-side auth gate on the
		// page itself (see requireWebclientAuth's comment above) -
		// /webclient-config/index.js above is still the one gated
		// endpoint that actually hands out real connection config.
		g.StaticFS("/webclient/static", http.Dir(global.Config.Gin.ResourcesPath+"/admin/static"))
		g.GET("/webclient", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/webclient/")
		})
		g.GET("/webclient/", func(c *gin.Context) {
			c.Header("Cache-Control", "no-store")
			c.File(global.Config.Gin.ResourcesPath + "/admin/webclient.html")
		})

		// Phase 2's Flutter web engine (Chr0mX/rustdesk, flutter/build/web/
		// output), served for Engine.vue's VITE_ENGINE_BASE_URL=/webclient/engine/.
		// Rustdesk-Server-Installer's update.sh builds and populates this
		// automatically (see its build_flutter_engine).
		g.StaticFS("/webclient/engine", http.Dir(global.Config.Gin.ResourcesPath+"/admin/engine"))
	}
	g.StaticFS("/_admin", http.Dir(global.Config.Gin.ResourcesPath+"/admin"))
}
