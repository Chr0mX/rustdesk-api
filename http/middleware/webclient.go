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

// webclientSessionData is what's actually stored server-side for a wc_sess
// cookie. Keeping ShareToken (not just the resolved fact that it was once
// valid) lets the cookie fast-path in WebclientAuth re-check the share
// record on every request, the same way it re-checks CheckUserEnable for
// token-based sessions - otherwise a session minted before a user got
// disabled, or before a share record got revoked/expired, would keep
// working for the rest of its TTL regardless.
type webclientSessionData struct {
	Token      string `json:"token"`
	ShareToken string `json:"share_token,omitempty"`
}

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
// query param again. When the ?token= path is what authed the visitor, that
// same token rides along in the cached session (see LookupWebclientSessionToken)
// so admin.Config.WebclientBridge can later hand it back to _admin - the
// share_token path carries no such token, since it isn't tied to any admin
// account.
func WebclientAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authed := false
		adminToken := ""
		shareToken := ""

		if sid, err := c.Cookie(webclientSessionCookie); err == nil && sid != "" {
			var sess webclientSessionData
			if err := global.Cache.Get(webclientSessionCachePrefix+sid, &sess); err == nil {
				// Re-validate on every request rather than just trusting the
				// cache entry exists: a token can be revoked, a user
				// disabled, or a share record deleted/expired at any point
				// during this session's TTL.
				switch {
				case sess.Token != "":
					user, _ := service.AllService.UserService.InfoByAccessToken(sess.Token)
					if user.Id != 0 && service.AllService.UserService.CheckUserEnable(user) {
						authed = true
						adminToken = sess.Token
					}
				case sess.ShareToken != "":
					sr := service.AllService.ShareRecordService.InfoByShareToken(sess.ShareToken)
					if sr.Id != 0 {
						authed = true
						shareToken = sess.ShareToken
					}
				}
				if !authed {
					// The cached session no longer checks out - drop it so
					// it doesn't keep getting silently re-tried.
					_ = global.Cache.Delete(webclientSessionCachePrefix + sid)
				}
			}
		}

		if !authed {
			if token := c.Query("token"); token != "" {
				user, _ := service.AllService.UserService.InfoByAccessToken(token)
				if user.Id != 0 && service.AllService.UserService.CheckUserEnable(user) {
					authed = true
					adminToken = token
				}
			}
		}

		if !authed {
			if st := c.Query("share_token"); st != "" {
				sr := service.AllService.ShareRecordService.InfoByShareToken(st)
				if sr.Id != 0 {
					authed = true
					shareToken = st
				}
			}
		}

		if authed {
			EstablishWebclientSession(c, adminToken, shareToken)
		}

		c.Set(WebclientAuthedKey, authed)
		c.Next()
	}
}

// EstablishWebclientSession mints a short-lived opaque session id, stores it
// server side (global.Cache, alongside adminToken/shareToken - pass "" for
// whichever doesn't apply) and drops it in an httpOnly cookie, same as
// WebclientAuth does on a successful ?token=/?share_token= check. Keeping
// both on the session lets the cookie fast-path re-validate the underlying
// user/share record on every request instead of just trusting the cache
// entry's existence. Exported so an already-authenticated admin request
// (see admin.Config.WebclientSession) can proactively establish the same
// session - useful when the admin console and webclient are reverse-proxied
// under different subdomains (see App.WebclientCookieDomain): the admin
// console can call this right after login so the webclient recognizes the
// visitor without needing a ?token= in the URL.
func EstablishWebclientSession(c *gin.Context, adminToken string, shareToken string) {
	sid := utils.RandomString(32)
	sess := webclientSessionData{Token: adminToken, ShareToken: shareToken}
	_ = global.Cache.Set(webclientSessionCachePrefix+sid, sess, webclientSessionTTL)
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(webclientSessionCookie, sid, webclientSessionTTL, "/", global.Config.App.WebclientCookieDomain, secure, true)
}

// RevokeWebclientSession invalidates the visitor's webclient session, if
// any: deletes it from global.Cache (so a still-held cookie or a copy of
// it can't be replayed) and clears the cookie in the response. Called from
// admin.Login.Logout so logging out of _admin also signs you out of the
// webclient - otherwise a previously-authed webclient tab (or the cookie
// alone, replayed) would keep working after logout. The next load of
// /webclient-config/index.js for this visitor now hits the unauthed path
// in web.Index.ConfigJs, which actively clears any id-server/relay-server/
// key it had previously set, rather than just leaving them stale.
func RevokeWebclientSession(c *gin.Context) {
	if sid, err := c.Cookie(webclientSessionCookie); err == nil && sid != "" {
		_ = global.Cache.Delete(webclientSessionCachePrefix + sid)
	}
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(webclientSessionCookie, "", -1, "/", global.Config.App.WebclientCookieDomain, secure, true)
}

// LookupWebclientSessionToken returns the admin api-token tied to the
// visitor's webclient session cookie, for admin.Config.WebclientBridge to
// bounce them straight into _admin. ok is false when there's no valid
// session at all; token is "" (with ok true) when the session exists but
// isn't tied to any admin account (established via share_token).
func LookupWebclientSessionToken(c *gin.Context) (token string, ok bool) {
	sid, err := c.Cookie(webclientSessionCookie)
	if err != nil || sid == "" {
		return "", false
	}
	var sess webclientSessionData
	if err := global.Cache.Get(webclientSessionCachePrefix+sid, &sess); err != nil {
		return "", false
	}
	return sess.Token, true
}
