package config

import (
	"os"
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
// webclient should use: the forced override if one is set, else the
// regular relay-server.
func (rd *Rustdesk) EffectiveWebclientRelayServer() string {
	if rd.WebclientRelayServer != "" {
		return rd.WebclientRelayServer
	}
	return rd.RelayServer
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
