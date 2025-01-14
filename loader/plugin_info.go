package loader

import (
	"context"
	plugins "github.com/bgrewell/gin-plugins"
	"net/rpc"
)

// PluginInfo struct that contains information about a running plugin
type PluginInfo struct {
	Path         string
	Name         string
	RouteRoot    string
	Routes       []*plugins.Route
	Pid          int
	Proto        string
	Ip           string
	Port         int
	Cookie       string
	Rpc          *rpc.Client
	CancelToken  *context.CancelFunc
	ExitChan     chan int
	HasExited    bool
	ExitCode     int
	PluginConfig *PluginConfig
}
