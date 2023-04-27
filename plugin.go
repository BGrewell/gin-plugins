package plugins

import (
	"encoding/json"
	"errors"
)

type Args struct {
	QueryParams map[string][]string
	Headers     map[string][]string
	Data        []byte
}

type Reply struct {
	Value interface{}
}

// Plugin is the interface that plugins must implement
type Plugin interface {
	Name() string
	RouteRoot() string
	Register(RegisterArgs, *RegisterReply) error
}

// PluginShared is a struct to hold the shared methods. It must be seperate from
// the PluginBase struct in order to prevent issues with the net.rpc trying to
// serve functions that don't fit the rpc function signatures.
type PluginShared struct {
}

func (p *PluginShared) Name() string {
	return "UnNamedPlugin"
}

func (p *PluginShared) RouteRoot() string {
	return ""
}

func (p *PluginShared) Serialize(v interface{}) (string, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return "", e
	}
	return string(b), nil
}

// PluginBase is a helper struct that implements some common shared methods
type PluginBase struct {
	PluginShared
}

func (p *PluginBase) Register(args RegisterArgs, reply *Reply) error {
	return errors.New("this method must be implemented")
}
