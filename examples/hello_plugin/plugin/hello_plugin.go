package plugin

import (
	plugins "github.com/bgrewell/gin-plugins"
	"net/http"
	"reflect"
)

// HelloMessage is just a simple helper struct to encapsulate the hello world message
type HelloMessage struct {
	Message string `json:"message"`
}

// Ensure that our type meets the requirements for being a plugin
var _ plugins.Plugin = &HelloPlugin{}

type HelloPlugin struct {
	// PluginBase provides some helper functions
	plugins.PluginBase
	message HelloMessage
}

func (h HelloPlugin) RouteRoot() string {
	return "hello"
}

func (h HelloPlugin) Name() string {
	t := reflect.TypeOf(h)
	return t.Name()
}

func (h *HelloPlugin) Register(args plugins.RegisterArgs, reply *plugins.RegisterReply) error {
	*reply = plugins.RegisterReply{Routes: make([]*plugins.Route, 1)}

	// Register our one hello world route
	h.message = HelloMessage{
		Message: "hello world!",
	}

	r := &plugins.Route{
		Path:       "hello",
		Method:     http.MethodGet,
		HandleFunc: "Hello",
	}
	reply.Routes[0] = r

	// Return no error
	return nil
}

func (h *HelloPlugin) Hello(args plugins.Args, c *string) error {
	v, e := h.Serialize(h.message)
	if e != nil {
		return e
	}
	*c = v
	return nil
}
