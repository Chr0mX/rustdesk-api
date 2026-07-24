package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"github.com/lejianwen/rustdesk-api/v2/utils"
	"net/http"
)

// WebclientAuthedKey is the gin context key WebclientAuth sets to true once
// the visitor has proven they're allowed to see the real id-server/
// relay-server/api-server/key values (see ConfigJs).
const WebclientAuthedKey = "webclientAuthed"

const webclientSessionCookie = "wc_sess"
const webclientSessionCachePrefix = "webclient_session:"
const webclientSessionTTL = 6 * 3600 // 6h, in seconds (matches cache.Handler.Set's exp unit)

// WebclientAuth gates access to the real Rustdesk connection config (id/
// relay/api server + key) that the bundled webclient needs. Without it,
// anyone who can reach /webclient-config/index.js - no login required -
// would get those values in plaintext, which is enough to abuse the
// server's rendezvous/relay for free or impersonate it.
//
// A visitor is considered authed if either:
//   - they hold a valid backend api-token (?token=, the same one the admin
//     console uses), i.e. they're a logged-in user opening the "Web Client"
//     button, or
//   - they hold a valid, non-expired share_token (?share_token=) minted by
//     AddressBook.ShareByWebClient for a specific peer.
//
// On success it mints a short-lived opaque session id, stores it server
// side (global.Cache) and drops it in an httpOnly cookie so the *next*
// request (e.g. the browser's automatic GET of /webclient-config/index.js
// right after loading /webclient/) is recognized too, without needing the
// query param again.
func WebclientAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authed := false

		if sid, err := c.Cookie(webclientSessionCookie); err == nil && sid != "" {
			var ok bool
			if err := global.Cache.Get(webclientSessionCachePrefix+sid, &ok); err == nil && ok {
				authed = true
			}
		}

		if !authed {
			if token := c.Query("token"); token != "" {
				user, _ := service.AllService.UserService.InfoByAccessToken(token)
				if user.Id != 0 && service.AllService.UserService.CheckUserEnable(user) {
					authed = true
				}
			}
		}

		if !authed {
			if shareToken := c.Query("share_token"); shareToken != "" {
				sr := service.AllService.ShareRecordService.InfoByShareToken(shareToken)
				if sr.Id != 0 {
					authed = true
				}
			}
		}

		if authed {
			EstablishWebclientSession(c)
		}

		c.Set(WebclientAuthedKey, authed)
		c.Next()
	}
}

// EstablishWebclientSession mints a short-lived opaque session id, stores it
// server side (global.Cache) and drops it in an httpOnly cookie, same as
// WebclientAuth does on a successful ?token=/?share_token= check. Exported
// so an already-authenticated admin request (see
// admin.Config.WebclientSession) can proactively establish the same
// session - useful when the admin console and webclient are reverse-proxied
// under different subdomains (see App.WebclientCookieDomain): the admin
// console can call this right after login so the webclient recognizes the
// visitor without needing a ?token= in the URL.
func EstablishWebclientSession(c *gin.Context) {
	sid := utils.RandomString(32)
	_ = global.Cache.Set(webclientSessionCachePrefix+sid, true, webclientSessionTTL)
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(webclientSessionCookie, sid, webclientSessionTTL, "/", global.Config.App.WebclientCookieDomain, secure, true)
}
