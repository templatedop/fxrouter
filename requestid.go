package fxrouter

import (
	//"context"
	//logger"gotemplate/logger"
	logger"github.com/templatedop/fxlogger"

	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
)

type RequestIDOptions struct {
	AllowSetting bool
}

func RequestID(options RequestIDOptions, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestID string

		if options.AllowSetting {
			// If Set-Request-Id header is set on request, use that for
			// Request-Id response header. Otherwise, generate a new one.
			requestID = c.Request.Header.Get("Request-Id")
		}

		if requestID == "" {
			requestID = uuid.New()
		}

		c.Writer.Header().Set("Request-Id", requestID)

	

		// l := logger.ToZerolog().With().Str("request_id", requestID).Logger()
		// logger = logger.FromZerolog(&l)
		// logger.Debug("Request ID from debug")
		// c.Set("logger", logger)



		/* This is working
		l := logger.ToZerolog().With().Str("request_id", requestID).Logger()
		logger= logger.FromZerolog(&l)
		logger.Debug("Request ID from debug")
		*/

		// type contextKey string

		// const loggerKey contextKey = "logger"

		// ctx := context.WithValue(c.Request.Context(), loggerKey, logger)

		//logger, ok := c.Request.Context().Value("logger").(*logger.Logger)
		//c.Set("logger", logger)

		// ctx = context.WithValue(, "logger", logger)
		//c.Request = c.Request.WithContext(ctx)

		// Set Transaction-Id header

		//logger= &l

		//logger.Debug("Request ID from debug")

		//n.Info().Msg("Request ID")

		c.Next()
	}
}

func GetLoggerFromContext(ctx *gin.Context, logger *logger.Logger) *logger.Logger {

	
	req := ctx.Writer.Header().Get("Request-Id")
	l := logger.ToZerolog().With().Str("request_id", req).Logger()
	logger = logger.FromZerolog(&l)
	logger.Debug("Request ID from debug")
	return logger
}
