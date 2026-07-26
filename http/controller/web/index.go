package web

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/middleware"
	apiResp "github.com/lejianwen/rustdesk-api/v2/http/response/api"
	"github.com/lejianwen/rustdesk-api/v2/service"
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
localStorage.removeItem('access_token');
localStorage.removeItem('user_info');
const ws2_prefix = 'wc-';
localStorage.removeItem(ws2_prefix+'api-server');
localStorage.removeItem(ws2_prefix+'custom-rendezvous-server');
localStorage.removeItem(ws2_prefix+'relay-server');
localStorage.removeItem(ws2_prefix+'key');
`

func (i *Index) Index(c *gin.Context) {
	c.Redirect(302, "/_admin/")
}

// WebclientLogin is what unauthenticated visitors to /webclient/* get
// instead of the bundled webclient (see router.go) - a minimal, dependency-
// free login form. It's deliberately separate from _admin's login: any
// enabled account can sign in here, not just admins, since there's no
// reason a non-admin user shouldn't be able to reach the webclient with
// their own account. It authenticates against the same POST /api/login
// every native RustDesk client already uses (not /admin/login), then
// redirects to /webclient/?token=<access_token>, which
// middleware.WebclientAuth already accepts from any enabled user - no new
// server-side auth path needed, just a UI for the one that already existed
// but only the "Web Client" button in _admin could reach.
func (i *Index) WebclientLogin(c *gin.Context) {
	c.Header("Cache-Control", "no-store, must-revalidate")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, webclientLoginPage)
}

const webclientLoginPage = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>RustDesk</title>
<style>
  :root{
    --bg-page:#f2f3f5; --bg-card:#fff; --text-primary:#1f2329; --text-secondary:#8a8f99;
    --border:#d9d9d9; --primary:#1677ff; --primary-hover:#4b96ff; --primary-disabled:#9cc4ff;
    --icon:#8a8f99; --err:#f56c6c;
  }
  html.dark{
    --bg-page:#18222c; --bg-card:#20293a; --text-primary:#e5eaf3; --text-secondary:#a3a6ad;
    --border:#3a4659; --icon:#a3a6ad;
  }
  *{box-sizing:border-box;}
  body{font-family:system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:var(--bg-page);display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;transition:background-color .2s;}
  .top-bar{position:fixed;top:20px;right:24px;}
  .theme-toggle{cursor:pointer;background:none;border:none;padding:6px;border-radius:50%;display:flex;color:var(--icon);}
  .theme-toggle:hover{background:rgba(128,128,128,.15);}
  .theme-toggle svg{width:20px;height:20px;}
  .card{background:var(--bg-card);padding:32px;border-radius:8px;box-shadow:0 2px 12px rgba(0,0,0,.08);width:340px;box-sizing:border-box;text-align:center;}
  .brand{display:flex;align-items:center;justify-content:center;gap:10px;margin-bottom:6px;}
  .brand svg{width:36px;height:36px;flex-shrink:0;}
  .brand-name{font-size:26px;font-weight:700;color:var(--text-primary);}
  .subtitle{font-size:14px;color:var(--text-secondary);margin-bottom:26px;}
  .field{position:relative;margin-bottom:14px;}
  .field svg{position:absolute;left:11px;top:50%;transform:translateY(-50%);width:16px;height:16px;color:var(--icon);}
  input{width:100%;box-sizing:border-box;padding:10px 12px 10px 36px;border:1px solid var(--border);border-radius:4px;font-size:14px;background:var(--bg-card);color:var(--text-primary);}
  input:focus{outline:none;border-color:var(--primary);}
  button[type=submit]{width:100%;padding:10px;background:var(--primary);color:#fff;border:none;border-radius:4px;font-size:14px;cursor:pointer;margin-top:6px;}
  button[type=submit]:hover{background:var(--primary-hover);}
  button[type=submit]:disabled{background:var(--primary-disabled);cursor:default;}
  .err{color:var(--err);font-size:13px;margin-bottom:12px;display:none;text-align:left;}
</style>
</head>
<body>
<div class="top-bar">
  <button type="button" class="theme-toggle" id="themeToggle" aria-label="Toggle dark mode"></button>
</div>
<div class="card">
  <div class="brand">
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><rect width="24" height="24" rx="6" fill="#1677ff"/><path d="M6 8.5C6 7.67 6.67 7 7.5 7h9c.83 0 1.5.67 1.5 1.5v6c0 .83-.67 1.5-1.5 1.5h-9A1.5 1.5 0 0 1 6 14.5v-6Z" stroke="#fff" stroke-width="1.4"/><path d="M9.5 17h5" stroke="#fff" stroke-width="1.4" stroke-linecap="round"/></svg>
    <span class="brand-name">RustDesk</span>
  </div>
  <div class="subtitle">Web Client</div>
  <div class="err" id="err"></div>
  <form id="f">
    <div class="field">
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Z" stroke="currentColor" stroke-width="1.6"/><path d="M4.5 20c1.3-3.5 4.2-5.5 7.5-5.5s6.2 2 7.5 5.5" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>
      <input type="text" id="username" placeholder="Username" autocomplete="username" required>
    </div>
    <div class="field">
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><rect x="5" y="10.5" width="14" height="9" rx="1.6" stroke="currentColor" stroke-width="1.6"/><path d="M8 10.5V8a4 4 0 1 1 8 0v2.5" stroke="currentColor" stroke-width="1.6"/></svg>
      <input type="password" id="password" placeholder="Password" autocomplete="current-password" required>
    </div>
    <button type="submit" id="submit">Login</button>
  </form>
</div>
<script>
(function () {
  var root = document.documentElement;
  var storeKey = 'webclient-login-theme';
  var sun = '<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><circle cx="12" cy="12" r="4" stroke="currentColor" stroke-width="1.6"/><path d="M12 2v2M12 20v2M4.2 4.2l1.4 1.4M18.4 18.4l1.4 1.4M2 12h2M20 12h2M4.2 19.8l1.4-1.4M18.4 5.6l1.4-1.4" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>';
  var moon = '<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M20 14.5A8 8 0 1 1 9.5 4a6.5 6.5 0 0 0 10.5 10.5Z" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"/></svg>';
  var btn = document.getElementById('themeToggle');
  function apply(dark) {
    root.classList.toggle('dark', dark);
    btn.innerHTML = dark ? sun : moon;
  }
  var stored = localStorage.getItem(storeKey);
  var dark = stored ? stored === 'dark' : window.matchMedia('(prefers-color-scheme: dark)').matches;
  apply(dark);
  btn.addEventListener('click', function () {
    dark = !dark;
    localStorage.setItem(storeKey, dark ? 'dark' : 'light');
    apply(dark);
  });
})();
document.getElementById('f').addEventListener('submit', async function (e) {
  e.preventDefault();
  var errEl = document.getElementById('err');
  var btn = document.getElementById('submit');
  errEl.style.display = 'none';
  btn.disabled = true;
  try {
    var res = await fetch('/api/login', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        username: document.getElementById('username').value,
        password: document.getElementById('password').value,
        deviceInfo: {name: navigator.userAgent, os: 'web', type: 'webclient'},
        id: '',
        uuid: ''
      })
    });
    var data = await res.json();
    if (res.ok && data.access_token) {
      window.location.href = '/webclient/?token=' + encodeURIComponent(data.access_token) + window.location.hash;
    } else {
      errEl.textContent = (data && data.message) || 'Login failed';
      errEl.style.display = 'block';
      btn.disabled = false;
    }
  } catch (err) {
    errEl.textContent = 'Login failed';
    errEl.style.display = 'block';
    btn.disabled = false;
  }
});
</script>
</body>
</html>`

// WebclientLogout is what the "Logout" link ConfigJs injects into the
// authenticated webclient page (see below) points at. The bundled webclient
// has no idea our custom wc_sess cookie exists, so nothing inside it could
// ever trigger a real sign-out - without this, the only way to actually
// deauthenticate was to log out of _admin instead. Also revokes the
// underlying access_token itself (service.UserService.Logout), not just
// the wc_sess cookie/cache entry, so replaying the same ?token=... in the
// URL bar afterward doesn't just quietly re-authenticate.
func (i *Index) WebclientLogout(c *gin.Context) {
	if token, ok := middleware.LookupWebclientSessionToken(c); ok && token != "" {
		user, _ := service.AllService.UserService.InfoByAccessToken(token)
		if user.Id != 0 {
			service.AllService.UserService.Logout(user, token)
		}
	}
	middleware.RevokeWebclientSession(c)
	c.Redirect(302, "/webclient/")
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

	// The bundled webclient has its own, entirely separate "Account" login
	// under Settings, unrelated to wc_sess: it reads/writes two local
	// options - "access_token" (used to build its Authorization header, see
	// hm() in main.dart.js) and "user_info" (a JSON-encoded UserPayload,
	// see bLJ() in main.dart.js: it JSON-decodes this and treats an empty
	// string as "not logged in" - access_token alone isn't enough, the
	// Account page reads the display fields straight from user_info without
	// hitting the network). Without both set, a visitor who already
	// authenticated through our outer login page would still hit a second,
	// unrelated login prompt the moment they opened Settings -> Account.
	// Since our login already went through the same POST /api/login and got
	// back the same access_token + user shape, seed both here so the
	// bundled client already considers itself logged in.
	accessTokenScript := "localStorage.removeItem('access_token');\nlocalStorage.removeItem('user_info');"
	if sessionToken, hasSession := middleware.LookupWebclientSessionToken(c); hasSession && sessionToken != "" {
		if user, _ := service.AllService.UserService.InfoByAccessToken(sessionToken); user.Id != 0 {
			if userInfoJson, err := json.Marshal((&apiResp.UserPayload{}).FromUser(user)); err == nil {
				accessTokenScript = fmt.Sprintf("localStorage.setItem('access_token', %v);\nlocalStorage.setItem('user_info', %v);",
					strconv.Quote(sessionToken), strconv.Quote(string(userInfoJson)))
			}
		}
	}

	tmp := fmt.Sprintf(`localStorage.setItem('api-server', %v);
localStorage.setItem('custom-rendezvous-server', %v);
localStorage.setItem('relay-server', %v);
localStorage.setItem('key', %v);
%s
const ws2_prefix = 'wc-';
localStorage.setItem(ws2_prefix+'api-server', %v);
localStorage.setItem(ws2_prefix+'custom-rendezvous-server', %v);
localStorage.setItem(ws2_prefix+'relay-server', %v);
localStorage.setItem(ws2_prefix+'key', %v);

window.webclient_magic_queryonline = %d;
window.ws_host = '%v';

// The bundled webclient has no concept of our custom session cookie, so
// nothing inside it can trigger a real sign-out - inject a visible link
// that does, since otherwise the only way to deauthenticate was via _admin.
(function () {
  var a = document.createElement('a');
  a.href = '/webclient-logout';
  a.textContent = 'Logout';
  a.style.cssText = 'position:fixed;bottom:8px;right:8px;z-index:2147483647;background:#1677ff;color:#fff;padding:4px 10px;border-radius:4px;font:12px system-ui,sans-serif;text-decoration:none;opacity:.85;';
  document.addEventListener('DOMContentLoaded', function () {
    document.body.appendChild(a);
  });
})();
`, strconv.Quote(apiServer), strconv.Quote(idServer), strconv.Quote(relayServer), strconv.Quote(key),
		accessTokenScript,
		strconv.Quote(apiServer), strconv.Quote(idServer), strconv.Quote(relayServer), strconv.Quote(key),
		magicQueryonline, global.Config.Rustdesk.WsHost)

	c.Header("Content-Type", "application/javascript")
	c.String(200, tmp)
}
