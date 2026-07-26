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

		g.GET("/webclient-config/index.js", wcAuth, i.ConfigJs)
		// Linked from the "Logout" button ConfigJs injects into the
		// authenticated webclient page - not nested under /webclient/ since
		// that'd conflict with its *filepath wildcard below.
		g.GET("/webclient-logout", i.WebclientLogout)

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
		wc.Use(wcAuth, requireWebclientAuth)
		wc.StaticFS("/", http.Dir(global.Config.Gin.ResourcesPath+"/web"))
	}
	g.StaticFS("/_admin", http.Dir(global.Config.Gin.ResourcesPath+"/admin"))
}
