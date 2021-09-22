package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"github.com/trafficstars/statuspage/handler/echostatuspage"
)

// HealthCheckResponse is response to health check
type HealthCheckResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// NewHealthCheckResponse returns marshaled json from `HealthCheckResponse` structure
func NewHealthCheckResponse(code int, msg string) []byte {
	r := HealthCheckResponse{code, msg}
	jsonBytes, _ := json.Marshal(r)
	return jsonBytes
}

// Server represents HTTP server for API
type Server struct {
	// internal http server can be replaced with fasthttp or
	// any other http server implementation if needed
	echo   *echo.Echo
	logger logrus.FieldLogger
}

// NewServer creates and setups new instance of HTTP server
func NewServer(l logrus.FieldLogger) *Server {
	echoServer := echo.New()

	// disable fancy echo terminal output
	echoServer.HideBanner = true
	echoServer.HidePort = true

	// add middleware
	echoServer.Pre(middleware.Recover())
	echoServer.Pre(middleware.RemoveTrailingSlash())
	echoServer.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Skipper: middleware.DefaultSkipper,
		Generator: func() string {
			UUID, err := uuid.NewUUID()
			if err == nil {
				return UUID.String()
			}
			return ""
		},
	}))

	api := echoServer.Group("/api")

	// init status and metric endpoints
	api.GET("/health-check", func(c echo.Context) error {
		// any system checks like database connection etc.
		return c.JSONBlob(http.StatusOK, NewHealthCheckResponse(http.StatusOK, `System is operational`))
	})
	api.GET("/health-check-simple", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	api.GET("/status.json", func(c echo.Context) error {
		return echostatuspage.StatusJSON(c)
	})
	api.GET("/status.prometheus", func(c echo.Context) error {
		return echostatuspage.StatusPrometheus(c)
	})

	return &Server{
		echo:   echoServer,
		logger: l,
	}
}

// Start runs echo server
func (s Server) Start(listen string) error {
	return s.echo.Start(listen)
}

func (s Server) Stop() error {
	return s.echo.Close()
}
