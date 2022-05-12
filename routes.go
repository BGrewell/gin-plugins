package plugins

type Route struct {
	Path       string `json:"path,omitempty"`
	Method     string `json:"method,omitempty"`
	HandleFunc string `json:"handle_func,omitempty"`
}
