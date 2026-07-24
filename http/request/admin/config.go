package admin

// WebclientConfigForm forces the id-server/relay-server/api-server the
// bundled webclient is handed, independent of the values native clients
// get. Leaving a field blank clears that override (falls back to the
// regular id-server/relay-server/api-server).
type WebclientConfigForm struct {
	WebclientIdServer    string `json:"webclient_id_server"`
	WebclientRelayServer string `json:"webclient_relay_server"`
	WebclientApiServer   string `json:"webclient_api_server"`
	// WebclientRelayFromApiServer: when true and WebclientRelayServer is
	// blank, derive the relay-server host from the effective webclient
	// api-server instead of using the plain relay-server value - useful
	// when hbbs/hbbr are on a LAN-only address but api-server is the
	// WAN-reachable one.
	WebclientRelayFromApiServer bool `json:"webclient_relay_from_api_server"`
}
