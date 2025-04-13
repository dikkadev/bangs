package main

import (
	"bangs/internal/watcher"
	"bangs/pkg/bangs"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dikkadev/prettyslog"
	flag "github.com/spf13/pflag"
)

//go:embed all:frontend/dist
var embeddedFrontend embed.FS

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

	// --- Serve Frontend from Embedded Filesystem ---

	// Create a sub-filesystem rooted at "frontend/dist"
	frontendFS, err := fs.Sub(embeddedFrontend, "frontend/dist")
	if err != nil {
		slog.Error("Failed to create sub VFS for frontend", "err", err)
		os.Exit(1)
	}
	frontendFileServer := http.FileServer(http.FS(frontendFS))

	// Serve static assets (/assets, /favicon.ico, etc.)
	mainRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Clean the path to prevent traversal issues within the embedded FS
		// Although less critical now, it's good practice.
		path := filepath.Clean(r.URL.Path)

		// Check if the requested path (excluding '/') corresponds to a file
		// in the embedded FS. We check paths like /assets/..., /favicon.ico
		if path != "/" {
			// fs.Stat needs a path without the leading slash
			if _, err := fs.Stat(frontendFS, strings.TrimPrefix(path, "/")); err == nil {
				frontendFileServer.ServeHTTP(w, r)
				return
			}
		}

		// If it's the root path or the file wasn't found, serve index.html
		// We need to manually open and serve index.html because http.FileServer
		// doesn't automatically serve index.html for directories in embedded FS.
		index, err := frontendFS.Open("index.html")
		if err != nil {
			slog.Error("Failed to open embedded index.html", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer index.Close()

		// Get file info for headers (optional but good)
		info, err := index.Stat()
		if err != nil {
			slog.Error("Failed to stat embedded index.html", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Serve the index.html content
		http.ServeContent(w, r, "index.html", info.ModTime(), index.(io.ReadSeeker))
	})

	// --- API and Bang Handlers ---

	// API endpoints (like listing bangs)
	apiHandler := bangs.Handler(false, ".") // Create a handler instance for API routes
	mainRouter.Handle("/api/", http.StripPrefix("/api", apiHandler))

	// Handle bang redirection logic at /bang
	mainRouter.Handle("/bang", bangs.Handler(allowNoBang, ignoreChar))

	// --- Start Server ---

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
