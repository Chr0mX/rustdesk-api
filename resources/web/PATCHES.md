# Hand patches applied to this vendored build

`resources/web` is a prebuilt, minified Flutter+JS bundle, not something we
build from source in this repo. When it gets replaced by a newer upstream
build, reapply these:

## 1. `index.html`

- `<base href="...">` set to `/webclient/` (the path it's actually served
  from - upstream builds bake in whatever `--base-href` they used).
- `<script src="/webclient-config/index.js"></script>` added right after
  `<title>RustDesk</title>` so the server-injected config (id-server/
  relay-server/key/api-server, see `http/controller/web/index.go`'s
  `ConfigJs`) loads before the app does.
- Firebase analytics removed entirely: the `libs/firebase-app.js` /
  `libs/firebase-analytics.js` `<script>` tags and the inline
  `firebaseConfig` / `firebase.initializeApp()` / `firebase.analytics()`
  block. That's RustDesk upstream's own telemetry wired to their Firebase
  project - no reason for a self-hosted server to ship it. The two lib
  files were deleted too.

## 2. `js/dist/index.js`

- `function xn(){...}` patched to `function xn(){return false}`.

  Upstream's original `xn()` gates a licensing check: it returns true for
  any hostname that isn't `rustdesk.com`/`localhost` - i.e. every
  self-hosted deployment. When true, the client fetches
  `${api-server}/api/lic/wc` expecting an Ed25519-signed license payload,
  and if that check doesn't pass (which it never will here - we don't
  implement `/api/lic/wc`), it calls `window.closeConnection()` and shows
  a "License Warning: To use the web client, you require a license that
  supports at least 10 users and 300 devices..." dialog *instead of* the
  normal login dialog, blocking every connection. Forcing `xn()` to
  `false` short-circuits the whole check (no network call, no dialog, no
  blocked connections) without touching the surrounding functions
  (`_n`/`n3`/`gn`/`hn`), which stay defined but dead. Unrelated: `_n`/`n3`
  (libsodium `crypto_sign_open`) are also used by `ba()` for a different,
  legitimate per-connection verifier - that path is untouched.
