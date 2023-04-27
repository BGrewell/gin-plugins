package loader

import (
	"errors"
	"fmt"
	plugins "github.com/BGrewell/gin-plugins"
	"github.com/BGrewell/gin-plugins/helpers"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"strings"
)

type DefaultPluginLoader struct {
	PluginDirectory string
	Cookie          string
	RouteGroup      *gin.RouterGroup
	plugins         map[string]*PluginInfo
	routeMap        map[string]*HandlerEntry
}

func (pl *DefaultPluginLoader) Initialize() (loadedPlugins []*PluginInfo, err error) {

	// List plugins
	loadedPlugins = make([]*PluginInfo, 0)
	plugs, err := pl.ListPlugins()
	if err != nil {
		return nil, err
	}

	for _, plug := range plugs {
		// Launch plugins
		info, err := pl.LaunchPlugin(plug)
		if err != nil {
			return nil, err
		}

		// Register plugins
		err = pl.RegisterPlugin(info.Name)
		if err != nil {
			return nil, err
		}
		loadedPlugins = append(loadedPlugins, info)
	}

	// Register control routes GET methods are just there for ease of use
	pl.RouteGroup.GET("load", pl.Load)
	pl.RouteGroup.POST("load", pl.Load)
	pl.RouteGroup.GET("unload", pl.Unload)
	pl.RouteGroup.DELETE("unload", pl.Unload)

	return loadedPlugins, nil
}

func (pl *DefaultPluginLoader) ListPlugins() (plugins []string, err error) {
	return helpers.FindPlugins(pl.PluginDirectory, "*.plugin")
}

func (pl *DefaultPluginLoader) LaunchPlugin(pluginPath string) (info *PluginInfo, err error) {
	info, err = executePlugin(pluginPath, pl.Cookie)
	if err != nil {
		return nil, err
	}
	pl.plugins[info.Name] = info
	return info, err
}

func (pl *DefaultPluginLoader) RegisterPlugin(pluginName string) (err error) {

	if plug, ok := pl.plugins[pluginName]; !ok {
		return errors.New(fmt.Sprintf("no plugin was found with the name: %s", pluginName))
	} else {
		// Connect the rpc client
		plug.Rpc, err = rpc.DialHTTP(plug.Proto, fmt.Sprintf("%s:%d", plug.Ip, plug.Port))
		if err != nil {
			return err
		}
		// Register the plugin
		ra := plugins.RegisterArgs{}
		rr := &plugins.RegisterReply{}
		err = plug.Rpc.Call(fmt.Sprintf("%s.Register", plug.Name), ra, rr)
		if err != nil {
			return err
		}

		// Setup routes
		plug.Routes = rr.Routes
		for _, route := range plug.Routes {
			// Build path
			root := ""
			if plug.RouteRoot != "" {
				root = fmt.Sprintf("%s/", plug.RouteRoot)
			}
			path := fmt.Sprintf("%s%s", root, route.Path)

			// Create a map to direct api calls to the correct plugin and function
			routeKey := fmt.Sprintf("%s:%s", route.Method, path)

			// Only add if this the first load. This is sloppy but since Gin doesn't provide a way to
			// remove routes and will panic if we try to add the same route twice we just check our
			// existing map to see if it has the entry already, if it does we skip the register as it
			// should already exist.
			if _, ok := pl.routeMap[routeKey]; !ok {
				log.Printf("Setting up route: %s -> %s::%s\n", routeKey, plug.Name, route.HandleFunc)
				handlerEntry := &HandlerEntry{
					PluginName: plug.Name,
					HandleFunc: route.HandleFunc,
				}
				pl.routeMap[routeKey] = handlerEntry

				// Create the entry in the RouterGroup
				pl.RouteGroup.Handle(route.Method, path, pl.callShim)
			}

		}

	}

	return nil
}

func (pl *DefaultPluginLoader) UnregisterPlugin(pluginName string) (err error) {
	if _, ok := pl.plugins[pluginName]; !ok {
		return errors.New("plugin not found")
	} else {
		return nil
	}
}

func (pl *DefaultPluginLoader) ClosePlugin(pluginName string) (err error) {
	if plug, ok := pl.plugins[pluginName]; !ok {
		return errors.New("plugin not found")
	} else {
		token := *plug.CancelToken
		token()
	}
	return nil
}

func (pl *DefaultPluginLoader) Load(c *gin.Context) {
	if name, ok := c.GetQuery("name"); ok {
		if plug, ok := pl.plugins[name]; ok {
			if !plug.HasExited {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "plugin is already loaded"})
				return
			}
			_, err := pl.LaunchPlugin(plug.Path)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
			err = pl.RegisterPlugin(plug.Name)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("no plugin with the name %s found", name)})
			return
		}
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing 'name' parameter specifying the plugin name"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin loaded"})
}

func (pl *DefaultPluginLoader) Unload(c *gin.Context) {
	if name, ok := c.GetQuery("name"); ok {
		if plug, ok := pl.plugins[name]; ok {
			if plug.HasExited {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "plugin is already unloaded"})
				return
			}
			err := pl.UnregisterPlugin(plug.Name)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
			err = pl.ClosePlugin(plug.Name)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("no plugin with the name %s found", name)})
			return
		}
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing 'name' parameter specifying the plugin name"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "plugin unloaded"})
}

func (pl *DefaultPluginLoader) callShim(c *gin.Context) {

	// Extract the RouteKey
	routeKey := fmt.Sprintf("%s:%s",
		c.Request.Method,
		strings.Replace(c.FullPath(), "/plugins/", "", 1))

	// Get the HandlerEntry
	if handler, ok := pl.routeMap[routeKey]; !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, "unknown function")
		return
	} else {
		// Make sure plugin isn't canceled
		if pl.plugins[handler.PluginName].HasExited {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("plugin has exited with code: %d", pl.plugins[handler.PluginName].ExitCode)})
			return
		}
		var err error

		// Get the plugin
		plug := pl.plugins[handler.PluginName]

		// Get the body if there is one
		data := make([]byte, 0)
		contentLength := c.Request.ContentLength
		if contentLength > 0 {
			data, err = ioutil.ReadAll(c.Request.Body)
			if err != nil {
				// handle error
			}
		}

		// Populate the args
		args := plugins.Args{
			QueryParams: c.Request.URL.Query(),
			Headers:     c.Request.Header,
			Data:        data,
		}
		var reply string

		// Make the rpc call
		err = plug.Rpc.Call(fmt.Sprintf("%s.%s", plug.Name, handler.HandleFunc), args, &reply)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		c.String(http.StatusOK, reply)
	}
}
