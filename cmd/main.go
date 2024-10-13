package main

import (
	"bangs/pkg/bangs"
	"bangs/web/assets"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/sett17/prettyslog"
	flag "github.com/spf13/pflag"
)

func main() {
	var bangsFile string
	flag.StringVarP(&bangsFile, "bangs", "b", "", "Path to the yaml file containg bang definitions")

	var debugLogs bool
	flag.BoolVarP(&debugLogs, "verbose", "v", false, "Show debug logs")

	var showHelp bool
	flag.BoolVarP(&showHelp, "help", "h", false, "Show this help")

	var port string
	flag.StringVarP(&port, "port", "p", "8080", "Port to listen on")

	flag.Parse()

	if showHelp {
		fmt.Println("help :(")
		flag.PrintDefaults()
		os.Exit(0)
	}

	logOptions := make([]prettyslog.Option, 0)
	if debugLogs {
		logOptions = append(logOptions, prettyslog.WithLevel(slog.LevelDebug))
	}
	slog.SetDefault(slog.New(prettyslog.NewPrettyslogHandler("APP", logOptions...)))
	if debugLogs {
		slog.Debug("Activated debug log entries")
	}

	if bangsFile == "" {
		slog.Error("No bangs definition file given")
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := bangs.Load("bangs.yaml")
	if err != nil {
		slog.Error("Error loading bangs", "err", err)
		return
	}

	mainRouter := http.NewServeMux()

	mainRouter.Handle("/assets/", assets.Handler())
	mainRouter.Handle("/", bangs.Handler())
	//
	// h2s := &http2.Server{}
	server := &http.Server{
		Addr: ":" + port,
		// Handler: h2c.NewHandler(mainRouter, h2s),
		Handler: mainRouter,
	}

	slog.Info("Starting server on", "port", port)

	err = server.ListenAndServe()
	if err != nil {
		slog.Error("Error starting server", "err", err)
		return
	}
}
