package loader

import (
	"bufio"
	"github.com/BGrewell/go-execute"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

type PluginLoader interface {
	Initialize() (loadedPlugins []*PluginInfo, err error)
	ListPlugins() (plugins []string, err error)
	LaunchPlugin(pluginPath string) (info *PluginInfo, err error)
	RegisterPlugin(pluginName string) (err error)
	UnregisterPlugin(pluginName string) (err error)
	ClosePlugin(pluginName string) (err error)
}

func NewPluginLoader(pluginDirectory string, routeGroup *gin.RouterGroup) PluginLoader {
	l := &DefaultPluginLoader{
		PluginDirectory: pluginDirectory,
		RouteGroup:      routeGroup,
		plugins:         make(map[string]*PluginInfo),
		routeMap:        make(map[string]*HandlerEntry),
	}

	return l
}

func executePlugin(pluginPath string) (info *PluginInfo, err error) {
	stdout, _, exitChan, cancel, err := execute.ExecuteAsyncWithCancel(pluginPath, nil)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Scan()
	output := scanner.Text()
	output = strings.Replace(output, "CONNECT{{", "", 1)
	output = strings.Replace(output, "}}", "", 1)
	fields := strings.Split(output, ":")
	port, err := strconv.Atoi(fields[4])
	if err != nil {
		return nil, err
	}
	info = &PluginInfo{
		Path:        pluginPath,
		Name:        fields[0],
		RouteRoot:   fields[1],
		Pid:         0,
		Proto:       fields[2],
		Ip:          fields[3],
		Port:        port,
		Cookie:      fields[5],
		Rpc:         nil,
		CancelToken: &cancel,
		ExitChan:    exitChan,
		ExitCode:    0,
	}

	// Create a go routine to watch for the plugin to exit
	go func() {
		ec := <-info.ExitChan
		info.ExitCode = ec
		info.HasExited = true
	}()

	return info, nil
}
