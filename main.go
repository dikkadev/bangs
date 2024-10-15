package main

import (
	"bangs/pkg/bangs"
	"bangs/web/assets"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sett17/prettyslog"
	flag "github.com/spf13/pflag"
)

var version = "dev"

func main() {
	getEnv := func(key, fallback string) string {
		if value, exists := os.LookupEnv(key); exists {
			return value
		}
		return fallback
	}

	getEnvBool := func(key string, fallback bool) bool {
		if value, exists := os.LookupEnv(key); exists {
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				slog.Warn("Invalid boolean value in environment variables", "env", key, "value", value)
				return fallback
			}
			return parsed
		}
		return fallback
	}

	bangsFileDefault := getEnv("BANGS_BANGFILE", "")
	debugLogsDefault := getEnvBool("BANGS_VERBOSE", false)
	showHelpDefault := getEnvBool("BANGS_HELP", false)
	portDefault := getEnv("BANGS_PORT", "8080")
	watchBangFileDefault := getEnvBool("BANGS_WATCH", false)

	var bangsFile string
	flag.StringVarP(&bangsFile, "bangs", "b", bangsFileDefault, "Path to the yaml file containing bang definitions")

	var debugLogs bool
	flag.BoolVarP(&debugLogs, "verbose", "v", debugLogsDefault, "Show debug logs")

	var showHelp bool
	flag.BoolVarP(&showHelp, "help", "h", showHelpDefault, "Show this help")

	var port string
	flag.StringVarP(&port, "port", "p", portDefault, "Port to listen on")

	var watchBangFile bool
	flag.BoolVarP(&watchBangFile, "watch", "w", watchBangFileDefault, "Reload bangs file on change")

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
	slog.Info("Starting bangs", "version", version)
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

	if watchBangFile {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			slog.Error("Error creating watcher", "err", err)
			return
		}
		defer watcher.Close()

		err = watcher.Add(bangsFile)
		if err != nil {
			slog.Error("Error adding file to watcher", "err", err)
			return
		}
		slog.Info("Watching bangs file", "file", bangsFile)

		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					slog.Debug("Watcher event", "op", event.Op, "file", event.Name)

					if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Rename == fsnotify.Rename {
						// Delay before re-adding to handle rename issues (apprently vim renames the file and overwrites somehing? not sure, but this seems to work)
						if event.Op&fsnotify.Rename == fsnotify.Rename {
							time.Sleep(100 * time.Millisecond)

							// Ensure file exists before re-adding to watcher
							if _, err := os.Stat(bangsFile); err == nil {
								err = watcher.Add(bangsFile)
								if err != nil {
									slog.Error("Error re-adding file to watcher", "err", err)
								}
							} else {
								slog.Error("File does not exist after rename; cannot re-add", "file", bangsFile)
							}
						}

						slog.Info("Bangs file changed; reloading", "file", event.Name)
						err := bangs.Load(bangsFile)
						if err != nil {
							slog.Error("Error reloading bangs", "err", err)
						} else {
							slog.Info("Bangs reloaded successfully", "file", bangsFile)
						}
					}

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					slog.Error("Watcher error", "err", err)
				}
			}
		}()
	}

	mainRouter := http.NewServeMux()

	mainRouter.Handle("/assets/", assets.Handler())
	mainRouter.Handle("/", bangs.Handler())

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
