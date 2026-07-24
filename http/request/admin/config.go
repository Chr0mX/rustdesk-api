package admin

// WebclientConfigForm forces the id-server/relay-server the bundled
// webclient is handed, independent of the values native clients get.
// Leaving a field blank clears the override (falls back to the regular
// id-server/relay-server).
type WebclientConfigForm struct {
	WebclientIdServer    string `json:"webclient_id_server"`
	WebclientRelayServer string `json:"webclient_relay_server"`
}
