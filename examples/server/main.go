package main

import (
	"github.com/bgrewell/gin-plugins/loader"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Register a basic routes to check the server is working
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "server is running...")
	})

	authorized := r.Group("admin/", gin.BasicAuth(gin.Accounts{
		"admin": "password",
		"user":  "user",
	}))

	authorized.GET("test", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		c.JSON(http.StatusOK, gin.H{"user": user})
	})

	return r
}

type Config struct {
	Message string `json:"message"`
}

func main() {

	r := setupRouter()
	rg := r.Group("plugins/")

	plug := loader.PluginConfig{
		PluginPath: "/home/ben/repos/bengrewell/dtac_tools/main.plugin",
		Enabled:    true,
		Cookie:     "this_is_not_a_security_feature",
		Hash:       "",
		Config: map[string]interface{}{
			"message": "this message comes from the plugin hosts configuration file",
		},
	}

	l := loader.NewPluginLoader("/home/ben/repos/bengrewell/dtac_tools/", map[string]*loader.PluginConfig{"/home/ben/repos/bengrewell/dtac_tools/main.plugin": &plug}, rg, false)
	active, err := l.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded %d plugins\n", len(active))
	for idx, plug := range active {
		log.Printf("  %d: %s [%s]\n", idx+1, plug.Name, plug.Path)
	}
	r.Run(":9999")
}
