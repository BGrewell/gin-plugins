package loader

import "path/filepath"

type PluginConfig struct {
	PluginPath string                 `json:"plugin_path" yaml:"plugin_path"`
	Enabled    bool                   `json:"enabled" yaml:"enabled"`
	Cookie     string                 `json:"cookie" yaml:"cookie"`
	Hash       string                 `json:"hash" yaml:"hash"`
	Config     map[string]interface{} `json:"config" yaml:"config"`
}

func (pc PluginConfig) Name() string {
	return filepath.Base(pc.PluginPath)
}
