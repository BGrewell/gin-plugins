# Gin Plugins

This project is designed to make "pluggable" logic for REST API's built using Gin. 
Due to the lack of support for dynamic modules in Go we are unable to create true 
plugins. There are frameworks like go-plugin by Hashicorp which is a great project
but is a little heavy for this use-case. Go also has built in support for plugins
in newer versions which works well but is limited to Linux. 

## Why does this library exist?

I needed a way to create pluggable logic for a REST API that would work on Linux,
Windows or Mac and not much existed except Hashicorp's go-plugin library which felt
a little heavy to expect other users to use to make plugins for the internal tool
that this library was written for so instead I made a simple plugin system.

## Parts

### Loader

The loader runs in your main process and has functions for discovering plugins,
launching them and connecting to their RPC. Once it connects to the RPC it
registers the plugin after which point its functions are made available. 

### Host

The host runs a process with the plugin logic contained within. It outputs information
on how to connect to the RPC over standard output (yes, this idea is stolen from go-plugin)

## Usage

### Building a plugin

To build a plugin you need to implement the `plugins.Plugin` interface. There is a complete
example in **/examples/hello_plugin/plugin/hello_plugin.go** 

First you create a struct that includes the `plugins.PluginBase` struct which implements some
helper methods.

```go
// Ensure that our struct meets the requirements for being a plugin
var _ plugins.Plugin = &CustomPlugin{}

type CustomPlugin struct {
    plugins.PluginBase
    X int
    Y int
    Z int
}
```

Next you need to create your register method which will setup the routes you can handle. In this
example we will have a single route called `Cube()` and a `Name()` function that returns the name
of our plugin type, this **must** match the name of the struct so you should just copy my method below

```go
func (p *CustomPlugin) Register(args plugins.RegisterArgs, reply *plugins.RegisterReply) error {
    *reply = plugins.RegisterReply{Routes: make([]*plugins.Route, 1)}
    r := &plugins.Route {
        Path: "cube", 
        Method: http.MethodGet,
        HandleFunc: "Cube"
    }
    reply.Routes[0] = r
	
    return nil
}

func (p CustomPlugin) Name() string {
    t := reflect.TypeOf(p)
    return t.Name()
}
```

Then you create any of your functions, in this case we'll make Cube. The parameters must be a concrete
type for the first one which is used to pass serializable arguments to your function. The second is the
return parameter and must be a pointer. We'll use plugins.Args which is an empty struct for the input
since we won't have any although you could pass a struct with an X, Y, and Z value and use those in the
calculation instead of the struct parameters we have above.

```go
type Cuber struct {
    Result int
}


func (p *CustomPlugin) Cube(args plugins.Args, c *string) error {
    cube := p.X * p.Y * p.Z
    cuber := &Cuber{Result: cube}
    value, err := p.Serialize(cuber)
    if err != nil {
        return err
    }
    *c = value
    return nil
}
```

The last method you need is optional but recommended. it is the root of the routes your plugin will
add. For example in my system it would be something like `http://ip:port/plugins/<root>/<route_path>`
so in this example if we set it to **cuber** it would be `http://ip:port/plugins/cuber/cube`

```go
func (p CustomPlugin) RouteRoot() string {
    return "cuber"
}
```

The final step is to make your main application that will host your plugin. You won't need to write
much code to do this as this library provides helpers for most of it.

```go
package main

import (
	"github.com/BGrewell/gin-plugins/host"
	"log"
)

func main() {
    p := new(plugin.CustomPlugin)
    h, err := host.NewPluginHost(p, "some_cookie_value_to_insure_you_dont_load_the_wrong_thing_not_a_security_feature")
    if err != nil {
        log.Fatal(err)
    }
    err = h.Serve()
    if err != nil {
        log.Fatal(err)
    }
}
```

Your plugin should now be complete and ready to compile

```bash
GOOS=linux go build -o plugins/custom.linux.plugin main.go
GOOS=windows go build -o plugins/custom.windows.plugin main.go
GOOS=darwin go build -o plugins/custom.darwin.plugin main.go 
```