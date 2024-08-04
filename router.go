package fxgin

import (
	"context"
	//"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	config "github.com/templatedop/fxconfig"
	logger "github.com/templatedop/fxlogger"
	"go.uber.org/fx"
)

type HandlerFunc func(*gin.Context)
type MiddlewareFunc func(*gin.Context)
type HTTPMethod string

const (
	POST   HTTPMethod = "POST"
	GET    HTTPMethod = "GET"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
)

// type Route struct {
// 	Method      HTTPMethod
// 	Path        string
// 	Handler     HandlerFunc
// 	Middlewares []MiddlewareFunc
// }

// type RouteGroup struct {
// 	Prefix      string
// 	Routes      []Route
// 	Groups      []RouteGroup
// 	Middlewares []MiddlewareFunc
// }

type Route struct {
	Method      HTTPMethod
	Path        string
	Handler     HandlerFunc
	Middlewares []gin.HandlerFunc
}

type RouteGroup struct {
	Prefix      string
	Routes      []Route
	Groups      []RouteGroup
	Middlewares []gin.HandlerFunc
}

type Routes []RouteGroup

type Router struct {
	*gin.Engine
	//CommonMiddlewares []MiddlewareFunc
}

func NewRouter(routes Routes, cfg config.Econfig, logger *logger.Logger) *Router {
	router := gin.New()
	router.Use(Exception(logger))
	router.Use(RequestID(RequestIDOptions{AllowSetting: true}, logger))
	router.Use(StructuredLogger(logger))
	router.Use(Cors(cfg))

	// for _, middleware := range routes {
	//     router.Use(gin.HandlerFunc(middleware))
	// }
	// for _, group := range routes {
	//     rg := router.Group(group.Prefix)
	//     for _, mw := range group.Middlewares {
	// 		rg.Use(gin.HandlerFunc(mw))
	//     }
	//     for _, route := range group.Routes {
	//         switch route.Method {
	//         case GET:
	//             rg.GET(route.Path, gin.HandlerFunc(route.Handler))
	//         case POST:
	//             rg.POST(route.Path, gin.HandlerFunc(route.Handler))
	//         case PUT:
	//             rg.PUT(route.Path, gin.HandlerFunc(route.Handler))
	//         case DELETE:
	//             rg.DELETE(route.Path, gin.HandlerFunc(route.Handler))
	//         }
	//     }
	// }

	for _, group := range routes {
		registerGroup(router.Group(""), group)
	}
	return &Router{
		Engine: router,
		//Log:    log,
	}
	//return router
}

func registerGroup(router *gin.RouterGroup, group RouteGroup) {
	groupRouter := router.Group(group.Prefix)
	//fmt.Println("Middleware len", len(group.Middlewares))
	for _, middleware := range group.Middlewares {
		//fmt.Println("Middleware", middleware)
		groupRouter.Use(middleware)
		// groupRouter.Use(func(c *gin.Context) {
		// 	middleware(c)
		// })
	}
	for _, route := range group.Routes {
		// if !validMethods[route.Method] {
		// 	panic("Invalid HTTP method: " + route.Method)
		// }
		handler := route.Handler
		// for _, middleware := range route.Middlewares {
		// 	handler = wrapMiddleware(handler, middleware)
		// }

		switch route.Method {
		case GET:
			groupRouter.GET(route.Path, gin.HandlerFunc(handler))
		case POST:
			groupRouter.POST(route.Path, gin.HandlerFunc(handler))
		case PUT:
			groupRouter.PUT(route.Path, gin.HandlerFunc(handler))
		case DELETE:
			groupRouter.DELETE(route.Path, gin.HandlerFunc(handler))
		}

		//groupRouter.Handle(string(route.Method), route.Path, gin.HandlerFunc(handler))
	}
	for _, nestedGroup := range group.Groups {
		registerGroup(groupRouter, nestedGroup)
	}
}

func wrapMiddleware(handler HandlerFunc, middleware MiddlewareFunc) HandlerFunc {
	return func(c *gin.Context) {
		middleware(c)
		handler(c)
	}
}

func RegisterRoutes(routes Routes) fx.Option {
	return fx.Provide(func() Routes {
		return routes
	})
}

var RouterModule = fx.Options(
	fx.Provide(NewRouter),
	fx.Provide(NewServer),
	fx.Invoke(startserver),
)

func NewServer(lc fx.Lifecycle, c config.Econfig, router *Router) *http.Server {
	s := &http.Server{
		Addr:    ":" + c.HttpPort,
		Handler: router.Engine,
	}

	return s
}

func startserver(lc fx.Lifecycle, log *logger.Logger, server *http.Server, c config.Econfig) {

	duration, err := time.ParseDuration(c.ShutDownTime)
	if err != nil {
		log.Error("Error in parsing duration", err.Error())
	}
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go server.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			//SetIsShuttingDown(true)
			shutdownCtx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()
			return server.Shutdown(shutdownCtx)
		},
	})
}
