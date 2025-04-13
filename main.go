package main

import (
	"bangs/internal/watcher"
	"bangs/pkg/bangs"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/dikkadev/prettyslog"
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
	portDefault := getEnv("BANGS_PORT", "8080")
	watchBangFileDefault := getEnvBool("BANGS_WATCH", false)
	allowNoBangDefault := getEnvBool("BANGS_ALLOW_NO_BANG", false)
	ignoreCharDefault := getEnv("BANGS_IGNORE_CHAR", ".")

	var bangsFile string
	flag.StringVarP(&bangsFile, "bangs", "b", bangsFileDefault, "Path to the yaml file containing bang definitions")

	var debugLogs bool
	flag.BoolVarP(&debugLogs, "verbose", "v", debugLogsDefault, "Show debug logs")

	var showHelp bool
	flag.BoolVarP(&showHelp, "help", "h", false, "Show this help")

	var port string
	flag.StringVarP(&port, "port", "p", portDefault, "Port to listen on")

	var watchBangFile bool
	flag.BoolVarP(&watchBangFile, "watch", "w", watchBangFileDefault, "Reload bangs file on change")

	var allowNoBang bool
	flag.BoolVarP(&allowNoBang, "allow-no-bang", "a", allowNoBangDefault, "Allow requests with no bang to be handled as if they have a bang")

	var ignoreChar string
	flag.StringVarP(&ignoreChar, "ignore-char", "i", ignoreCharDefault, "Start with this character to ignore bangs (only uses first character of the string)")

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
		go watcher.WatchFile(bangsFile, func() error {
			return bangs.Load(bangsFile)
		})
	}

	mainRouter := http.NewServeMux()

	// Serve static files from frontend/dist
	fs := http.FileServer(http.Dir("./frontend/dist"))
	mainRouter.Handle("/assets/", fs) // Serve specific assets like JS, CSS

	// API endpoints (like listing bangs)
	apiHandler := bangs.Handler(false, ".")                          // Create a handler instance for API routes (params don't matter much here)
	mainRouter.Handle("/api/", http.StripPrefix("/api", apiHandler)) // Mount API handler under /api/

	// Handle bang redirection logic at /bang (uses a separate instance with correct params)
	mainRouter.Handle("/bang", bangs.Handler(allowNoBang, ignoreChar))

	// Serve index.html for the root and any other paths (SPA routing)
	mainRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If the requested file exists in dist, serve it (e.g., favicon.ico)
		filePath := "./frontend/dist" + r.URL.Path
		if _, err := os.Stat(filePath); err == nil {
			// Check if it's actually a file and not a directory
			info, _ := os.Stat(filePath)
			if !info.IsDir() {
				http.ServeFile(w, r, filePath)
				return
			}
		}
		// Otherwise, serve index.html
		http.ServeFile(w, r, "./frontend/dist/index.html")
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mainRouter,
	}

	slog.Info("Starting server on", "port", port)

	err = server.ListenAndServe()
	if err != nil {
		slog.Error("Error starting server", "err", err)
		return
	}
}
