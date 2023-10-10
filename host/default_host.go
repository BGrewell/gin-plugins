package host

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"

	plugins "github.com/bgrewell/gin-plugins"
	"github.com/bgrewell/gin-plugins/helpers"
)

type DefaultPluginHost struct {
	Plugin interface{}
	Proto  string
	Ip     string
	Cookie string
	port   int
}

func (ph *DefaultPluginHost) Serve() error {
	// Hacky way to keep the net.rpc package from complaining about some method signatures
	logger := log.Default()
	logger.SetOutput(ioutil.Discard)

	// Register plugin
	err := rpc.Register(ph.Plugin)
	logger.SetOutput(os.Stderr)
	if err != nil {
		return err
	}

	// Find a TCP port to use
	ph.port, err = helpers.GetUnusedTcpPort()
	if err != nil {
		return err
	}

	// Output connection information ( format: CONNECT{{NAME:ROUTE_ROOT:PROTO:IP:PORT:COOKIE}} )
	fmt.Printf("CONNECT{{%s:%s:%s:%s:%d:%s}}\n", ph.Plugin.(plugins.Plugin).Name(), ph.Plugin.(plugins.Plugin).RouteRoot(), ph.Proto, ph.Ip, ph.port, ph.Cookie)

	// Listen for connections
	l, e := net.Listen(ph.Proto, fmt.Sprintf("%s:%d", ph.Ip, ph.port))
	if e != nil {
		return e
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}
		go jsonrpc.ServeConn(conn) // Serve connection with JSON-RPC
	}
}

func (ph *DefaultPluginHost) GetPort() int {
	return ph.port
}
