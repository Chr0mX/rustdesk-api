package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/controller/web"
	"net/http"
)

func WebInit(g *gin.Engine) {
	i := &web.Index{}
	g.GET("/", i.Index)

	if global.Config.App.WebClient == 1 {
		g.GET("/webclient-config/index.js", i.ConfigJs)
		g.StaticFS("/webclient", http.Dir(global.Config.Gin.ResourcesPath+"/web"))

		// web2.1/web: serve index.html dynamically ({{CUSTOM_CONFIG}} substitution)
		// and serve all other assets as static files.
		web21FS := http.Dir(global.Config.Gin.ResourcesPath + "/web2.1/web")
		fileServer := http.StripPrefix("/webclient2", http.FileServer(web21FS))
		webclient2Handler := func(c *gin.Context) {
			fp := c.Param("filepath")
			if fp == "/" || fp == "/index.html" {
				i.WebClient2Index(c)
				return
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
		}
		g.GET("/webclient2/*filepath", webclient2Handler)
		g.HEAD("/webclient2/*filepath", webclient2Handler)
	}

	// Serve web2.1 Umi admin console (replaces the previously missing /resources/admin).
	g.StaticFS("/_admin", http.Dir(global.Config.Gin.ResourcesPath+"/web2.1"))
}
