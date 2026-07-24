package config

import (
	"net"
	"net/url"
	"os"
	"strconv"
)

const (
	DefaultIdServerPort    = 21116
	DefaultRelayServerPort = 21117
)

type Rustdesk struct {
	IdServer        string `mapstructure:"id-server"`
	IdServerPort    int    `mapstructure:"-"`
	RelayServer     string `mapstructure:"relay-server"`
	RelayServerPort int    `mapstructure:"-"`
	ApiServer       string `mapstructure:"api-server"`
	Key             string `mapstructure:"key"`
	KeyFile         string `mapstructure:"key-file"`
	Personal        int    `mapstructure:"personal"`
	//webclient-magic-queryonline
	WebclientMagicQueryonline int    `mapstructure:"webclient-magic-queryonline"`
	WsHost                    string `mapstructure:"ws-host"`
	// WebclientIdServer/WebclientRelayServer force-override the id-server/
	// relay-server handed to the bundled webclient specifically (admin
	// editable at runtime, see Config.UpdateWebclientConfig), independent of
	// what native clients get from IdServer/RelayServer above.
	WebclientIdServer    string `mapstructure:"webclient-id-server"`
	WebclientRelayServer string `mapstructure:"webclient-relay-server"`
	// WebclientRelayFromApiServer: when true and WebclientRelayServer isn't
	// set, derive the webclient's relay-server host from ApiServer instead
	// of using RelayServer's host verbatim (keeping RelayServer's port, or
	// DefaultRelayServerPort if that can't be parsed). Useful when hbbs/hbbr
	// sit on a LAN-only address but api-server is the one WAN-reachable
	// address (reverse-proxied, DNS'd, etc.) - opt-in, since for most
	// deployments RelayServer is already correct as-is and this would just
	// be a wrong guess.
	WebclientRelayFromApiServer bool `mapstructure:"webclient-relay-from-api-server"`
}

// EffectiveWebclientIdServer returns the id-server the bundled webclient
// should use: the forced override if one is set, else the regular id-server.
func (rd *Rustdesk) EffectiveWebclientIdServer() string {
	if rd.WebclientIdServer != "" {
		return rd.WebclientIdServer
	}
	return rd.IdServer
}

// EffectiveWebclientRelayServer returns the relay-server the bundled
// webclient should use: the forced override if one is set; else, if
// WebclientRelayFromApiServer is on, api-server's host paired with
// relay-server's port (falling back to RelayServer verbatim if api-server
// can't be parsed); else the regular relay-server.
func (rd *Rustdesk) EffectiveWebclientRelayServer() string {
	if rd.WebclientRelayServer != "" {
		return rd.WebclientRelayServer
	}
	if rd.WebclientRelayFromApiServer {
		if host := hostFromApiServer(rd.ApiServer); host != "" {
			return host + ":" + portOrDefault(rd.RelayServer, DefaultRelayServerPort)
		}
	}
	return rd.RelayServer
}

// hostFromApiServer extracts just the hostname (no scheme, no port) from
// an api-server URL like "http://121.6.58.12:21114". Returns "" if
// apiServer isn't a parseable URL with a host.
func hostFromApiServer(apiServer string) string {
	u, err := url.Parse(apiServer)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// portOrDefault pulls the port out of a "host:port" string, falling back
// to defaultPort (e.g. DefaultRelayServerPort) if hostPort has no parseable
// port of its own.
func portOrDefault(hostPort string, defaultPort int) string {
	if _, port, err := net.SplitHostPort(hostPort); err == nil && port != "" {
		return port
	}
	return strconv.Itoa(defaultPort)
}

func (rd *Rustdesk) LoadKeyFile() {
	// Load key file
	if rd.Key != "" {
		return
	}
	if rd.KeyFile != "" {
		// Load key from file
		b, err := os.ReadFile(rd.KeyFile)
		if err != nil {
			return
		}
		rd.Key = string(b)
		return
	}
}
