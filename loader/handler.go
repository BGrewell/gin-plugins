package loader

// Struct to hold a plugin name and handler function name. This is used to ensure that we don't try to
// register the same route twice with Gin which would cause an error.
type HandlerEntry struct {
	PluginName string
	HandleFunc string
}
