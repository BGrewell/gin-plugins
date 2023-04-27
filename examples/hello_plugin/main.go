package main

import (
	"github.com/bgrewell/gin-plugins/examples/hello_plugin/plugin"
	"github.com/bgrewell/gin-plugins/host"
	"log"
)

func main() {

	p := new(plugin.HelloPlugin)

	h, err := host.NewPluginHost(p, "this_is_not_a_security_feature")
	if err != nil {
		log.Fatal(err)
	}

	err = h.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
