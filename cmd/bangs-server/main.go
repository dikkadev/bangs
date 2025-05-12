package main

import (
	"bangs/internal/watcher" // Will need adjustment based on final project structure
	"bangs/pkg/bangs"        // Will need adjustment based on final project structure
	"bangs/web"              // Add import for the new web package
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

// //go:embed all:../../frontend/dist
// var embeddedFrontend embed.FS

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
	allowMultiBangDefault := getEnvBool("BANGS_ALLOW_MULTI_BANG", false)
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

	var allowMultiBang bool
	flag.BoolVarP(&allowMultiBang, "allow-multi-bang", "m", allowMultiBangDefault, "Allow multiple bangs in a single request")

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

	bangHandler := bangs.Handler(allowNoBang, allowMultiBang, ignoreChar)

	mainRouter.Handle("/bang/", http.StripPrefix("/bang", bangHandler))

	frontendFS, err := web.FrontendFS()
	if err != nil {
		slog.Error("Failed to get embedded frontend filesystem", "err", err)
		os.Exit(1)
	}

	frontendFileServer := http.FileServer(http.FS(frontendFS))

	mainRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "" {
			bangHandler.ServeHTTP(w, r)
			return
		}

		path := filepath.Clean(r.URL.Path)

		if path != "/" {
			if _, err := fs.Stat(frontendFS, strings.TrimPrefix(path, "/")); err == nil {
				frontendFileServer.ServeHTTP(w, r)
				return
			}
		}

		index, err := frontendFS.Open("index.html")
		if err != nil {
			slog.Error("Failed to open embedded index.html", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer index.Close()

		info, err := index.Stat()
		if err != nil {
			slog.Error("Failed to stat embedded index.html", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, "index.html", info.ModTime(), index.(io.ReadSeeker))
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
