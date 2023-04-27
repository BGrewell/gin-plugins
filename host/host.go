package host

import (
	"errors"
	plugins "github.com/bgrewell/gin-plugins"
)

type PluginHost interface {
	Serve() error
	GetPort() int
}

func NewPluginHost(plugin interface{}, cookie string) (hostPlugin PluginHost, err error) {

	if _, ok := plugin.(plugins.Plugin); !ok {
		return nil, errors.New("structure passed as plugin doesn't implement the Plugin interface")
	}

	plug := &DefaultPluginHost{
		Plugin: plugin,
		Proto:  "tcp",
		Ip:     "127.0.0.1",
		Cookie: cookie,
	}

	return plug, nil
}
