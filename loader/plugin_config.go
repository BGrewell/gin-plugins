package loader

import "path/filepath"

type PluginConfig struct {
	PluginPath string                 `json:"plugin_path,omitempty" yaml:"plugin_path"`
	Enabled    bool                   `json:"enabled,omitempty" yaml:"enabled"`
	Cookie     string                 `json:"cookie,omitempty" yaml:"cookie"`
	Hash       string                 `json:"hash,omitempty" yaml:"hash"`
	Config     map[string]interface{} `json:"config,omitempty" yaml:"config"`
}

func (pc PluginConfig) Name() string {
	return filepath.Base(pc.PluginPath)
}
