package loader

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/BGrewell/go-execute"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strconv"
	"strings"
)

// PluginLoader is the interface that any plugin loaders must satisfy. The default concrete implementation is
// the DefaultPluginLoader which at this point in time is the only planned loader.
type PluginLoader interface {
	Initialize() (loadedPlugins []*PluginInfo, err error)
	ListPlugins() (plugins []string, err error)
	LaunchPlugin(config *PluginConfig) (info *PluginInfo, err error)
	RegisterPlugin(pluginName string) (err error)
	UnregisterPlugin(pluginName string) (err error)
	ClosePlugin(pluginName string) (err error)
}

// NewPluginLoader takes in the plugin directory, the sanity cookie and the routeGroup.
//
//	pluginDirectory: the directory that contains all the plugins
//	cookie: the sanity cookie that is used to verify that what is being executed is the expected plugin
//	routeGroup: the routeGroup that all the plugin routes will be placed inside
func NewPluginLoader(pluginDirectory string, plugConfigs map[string]*PluginConfig, routeGroup *gin.RouterGroup, loadUnconfiguredPlugins bool) PluginLoader {
	l := &DefaultPluginLoader{
		PluginDirectory:         pluginDirectory,
		PluginConfigs:           plugConfigs,
		RouteGroup:              routeGroup,
		loadUnconfiguredPlugins: loadUnconfiguredPlugins,
		plugins:                 make(map[string]*PluginInfo),
		routeMap:                make(map[string]*HandlerEntry),
	}

	return l
}

// executePlugin is called to launch a plugin
func executePlugin(config *PluginConfig) (info *PluginInfo, err error) {
	// If the hash was configured with a sha1 hash verify that the loaded image matches the expected value
	if config.Hash != "" {
		if err := hashValid(config.PluginPath, config.Hash); err == nil {
			return nil, err
		}
	}

	stdout, _, exitChan, cancel, err := execute.ExecuteAsyncWithCancel(config.PluginPath, nil)
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
		Path:         config.PluginPath,
		Name:         fields[0],
		RouteRoot:    fields[1],
		Pid:          0,
		Proto:        fields[2],
		Ip:           fields[3],
		Port:         port,
		Cookie:       fields[5],
		Rpc:          nil,
		CancelToken:  &cancel,
		ExitChan:     exitChan,
		ExitCode:     0,
		PluginConfig: config,
	}

	// Ensure the cookies match
	if info.Cookie != config.Cookie {
		return nil, errors.New(fmt.Sprintf("cookie: %s did not match expected value %s", info.Cookie, config.Cookie))
	}

	// Create a go routine to watch for the plugin to exit
	go func() {
		ec := <-info.ExitChan
		info.ExitCode = ec
		info.HasExited = true
	}()
	return info, nil
}

func hashValid(pluginPath string, expectedHash string) (err error) {
	// Ensure expected hash is lowercase
	expectedHash = strings.ToLower(expectedHash)

	// Read the binary file from disk
	binaryData, err := ioutil.ReadFile(pluginPath)
	if err != nil {
		return
	}

	// Compute the SHA1 hash of the binary data
	sha1Hash := sha1.Sum(binaryData)

	// Convert the hash to a hex-encoded string
	sha1Hex := hex.EncodeToString(sha1Hash[:])

	if sha1Hex != expectedHash {
		return fmt.Errorf("SHA1 hash: %s did not match expected value %s\"", sha1Hex, expectedHash)
	}

	return nil
}
