package main

import (
	"crypto/tls"
	"io/ioutil"
	stdLog "log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	loggerElasticHook "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/getsentry/sentry-go"
	"github.com/hasansino/environment"
	loggerSentryHook "github.com/makasim/sentryhook"
	"github.com/sirupsen/logrus"
	"github.com/trafficstars/metrics"
	"github.com/trafficstars/registry"

	"github.com/hasansino/apptemplate/internal/api"
	"github.com/hasansino/apptemplate/internal/config"
)

// this variables are passed as arguments upon build
var (
	buildDate   string
	buildCommit string
)

// global variables
var (
	cfg *config.Config
	log logrus.StdLogger
)

func init() {
	rand.Seed(time.Now().Unix())

	if len(buildDate) == 0 {
		buildDate = "dev"
	}
	if len(buildCommit) == 0 {
		buildCommit = "dev"
	}

	log = stdLog.New(os.Stdout, "", 0)

	log.Printf("Build date: %s\n", buildDate)
	log.Printf("Build commit: %s\n", buildCommit)

	initRegistry()
	initMetrics()
	initLogger()
	initSentry()
	initProfiling()

	// check for debug mode
	if cfg.Debug {
		log.Println("!!! Application is running in debug mode !!!")
	}
}

// initRegistry initializes configuration and connects to consul k/v storage
func initRegistry() {
	cfg = new(config.Config)

	r, err := registry.New(os.Getenv("REGISTRY_DSN"), os.Args)
	if err != nil {
		log.Fatalf("Failed to initialize registry: %s\n", err.Error())
	}

	// bind registry to configuration
	if err = r.Bind(cfg); err != nil {
		log.Fatalf("Failed to bind registry to config struct: %s\n", err.Error())
	}

	// print config to stdout, can expose sensitive data
	log.Printf("%s\n", cfg.String())
}

// initMetrics initializes prometheus metrics
func initMetrics() {
	// ignore error, if there's no hostname we use empty string instead
	hostname, _ := os.Hostname()

	// this labels will be inherited by all metrics
	metrics.SetDefaultTags(metrics.Tags{
		"service":     cfg.ServiceName,
		"hostname":    hostname,
		"environment": environment.GetEnvironment().String(),
	})

	// write metrics with build date and commit hash
	buildMetric := metrics.GaugeInt64("build", metrics.Tags{
		"date":   buildDate,
		"commit": buildCommit,
	})
	buildMetric.Set(1)
	buildMetric.SetGCEnabled(false)

	// call GC every 5 minutes
	go func() {
		timer := time.NewTicker(5 * time.Minute)
		for {
			<-timer.C
			metrics.GC()
		}
	}()
}

// initLogger initializes logger instance and output connections
func initLogger() {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		DisableTimestamp: true,
	}

	switch cfg.Logger.Output {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	default:
		logger.SetOutput(ioutil.Discard)
	}

	// set logger level
	if level, err := logrus.ParseLevel(cfg.Logger.Level); err == nil {
		log.Printf("Logger level set to %s\n", level.String())
		logger.SetLevel(level)
	} else {
		log.Printf("Failed to parse log level: %s\n", err.Error())
		logger.SetLevel(logrus.ErrorLevel)
	}

	// connect to logstash
	if len(cfg.Logger.Address) > 0 {
		conn, err := net.Dial("udp", cfg.Logger.Address)
		if err != nil {
			log.Printf("Could not connect to logstash server: %v\n", err)
		} else {
			hook := loggerElasticHook.New(conn, loggerElasticHook.DefaultFormatter(
				logrus.Fields{
					"service":     cfg.ServiceName,
					"environment": environment.GetEnvironment().String(),
				}),
			)
			logger.AddHook(hook)
			log.Println("Connected to logstash service")
		}
	} else {
		log.Println("No logstash address was provided, using default logger")
	}

	// construct base logger
	log = logger
}

// initSentry connects to sentry and hooks to logger instance
func initSentry() {
	if len(cfg.Sentry.DSN) == 0 {
		return
	}

	// connect to sentry server
	if err := sentry.Init(sentry.ClientOptions{
		// do not check SSL certificate
		HTTPTransport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Dsn:              cfg.Sentry.DSN,
		Debug:            cfg.Sentry.Debug,
		AttachStacktrace: cfg.Sentry.AttachStacktrace,
		SampleRate:       cfg.Sentry.SampleRate,
		Release:          buildCommit,
		Environment:      environment.GetEnvironment().String(),
	}); err != nil {
		log.Printf("Failed to initialize sentry client: %s", err.Error())
		return
	}

	// hook to logger instance
	hook := loggerSentryHook.New(
		[]logrus.Level{
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
		loggerSentryHook.WithTags(map[string]string{
			"service": cfg.ServiceName,
		}),
	)

	log.(*logrus.Logger).AddHook(hook)
	log.Println("Connected to sentry service")
}

// initProfiling starts basic http server with pprof
func initProfiling() {
	if len(cfg.ListenPprof) == 0 {
		return
	}
	go func() {
		log.Printf("Profiler listening on %s", cfg.ListenPprof)
		if err := http.ListenAndServe(cfg.ListenPprof, nil); err != nil {
			log.Printf("Failed to start pprof http server: %s", err.Error())
		}
	}()
}

func main() {
	var (
		logger    = log.(logrus.FieldLogger)
		apiServer = api.NewServer(logger)
	)

	// start HTTP API server
	go func() {
		log.Printf("API server listening on %s", cfg.Listen)
		if err := apiServer.Start(cfg.Listen); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start API server: %s", err.Error())
		}
	}()

	log.Println("Hello World")

	// listen for exit signals
	sys := make(chan os.Signal, 1)
	signal.Notify(sys, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	shutdown(<-sys, apiServer)
}

// shutdown implements all graceful shutdown logic
func shutdown(_ os.Signal, api *api.Server) {
	log.Println("Shutting down...")
	if err := api.Stop(); err != nil {
		log.Println(err)
	}
	os.Exit(0)
}
