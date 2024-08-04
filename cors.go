package fxrouter

import (
	//fxconfig "gotemplate/config"
	"fmt"
	"time"

	fxconfig "github.com/templatedop/fxconfig"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors(cfg fxconfig.Econfig) gin.HandlerFunc {
	return func(c *gin.Context) {

		fmt.Println("Reached inside Cors middleware")

		// TODO: Add the cors configuration
		//config := cors.DefaultConfig()
		config := cors.Config{
			AllowOrigins:           []string{"*"},
			//AllowAllOrigins:        true,
			AllowCredentials:       true,
			AllowHeaders:           []string{"x-request-id", "Content-Type", "Authorization"},
			AllowBrowserExtensions: false,
			AllowMethods:           []string{"GET"},
			MaxAge:                 12 * time.Hour,
		}
		cors.New(config)(c)

		c.Next()
	}
}
