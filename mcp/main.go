package main

import (
	"bangs/internal/watcher"
	"bangs/pkg/bangs"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/dikkadev/prettyslog"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	flag "github.com/spf13/pflag"
)

var version = "dev"

type BangResult struct {
	Bang string `json:"bang"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type MultiBangResult struct {
	Results []BangResult `json:"results"`
	Errors  []string     `json:"errors,omitempty"`
}

type BangInfo struct {
	Name        string `json:"name"`
	Bang        string `json:"bang"`
	Description string `json:"description"`
	Category    string `json:"category"`
	URL         string `json:"url"`
}

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

	// Environment variables
	bangsFileDefault := getEnv("BANGS_BANGFILE", "")
	debugLogsDefault := getEnvBool("BANGS_VERBOSE", false)
	watchBangFileDefault := getEnvBool("BANGS_WATCH", false)
	httpModeDefault := getEnvBool("BANGS_MCP_HTTP", false)
	portDefault := getEnv("BANGS_MCP_PORT", "8081")

	// Command line flags
	var bangsFile string
	flag.StringVarP(&bangsFile, "bangs", "b", bangsFileDefault, "Path to the yaml file containing bang definitions")

	var debugLogs bool
	flag.BoolVarP(&debugLogs, "verbose", "v", debugLogsDefault, "Show debug logs")

	var showHelp bool
	flag.BoolVarP(&showHelp, "help", "h", false, "Show this help")

	var watchBangFile bool
	flag.BoolVarP(&watchBangFile, "watch", "w", watchBangFileDefault, "Reload bangs file on change")

	var httpMode bool
	flag.BoolVar(&httpMode, "http", httpModeDefault, "Run in HTTP mode instead of stdio")

	var port string
	flag.StringVarP(&port, "port", "p", portDefault, "Port to listen on (HTTP mode only)")

	flag.Parse()

	if showHelp {
		fmt.Println("Bangs MCP Server")
		fmt.Println("Provides Model Context Protocol access to bangs search functionality")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  go run github.com/dikkadev/bangs/mcp@latest -b ~/bangs.yaml\n")
		fmt.Println()
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Setup logging
	logOptions := make([]prettyslog.Option, 0)
	if debugLogs {
		logOptions = append(logOptions, prettyslog.WithLevel(slog.LevelDebug))
	}
	slog.SetDefault(slog.New(prettyslog.NewPrettyslogHandler("MCP", logOptions...)))
	slog.Info("Starting bangs MCP server", "version", version)
	if debugLogs {
		slog.Debug("Activated debug log entries")
	}

	// Validate bangs file
	if bangsFile == "" {
		fmt.Fprintf(os.Stderr, "Error: bangs.yaml not found\n\n")
		fmt.Fprintf(os.Stderr, "To get started:\n")
		fmt.Fprintf(os.Stderr, "1. Download the default configuration to your home directory:\n")
		fmt.Fprintf(os.Stderr, "   curl -o ~/bangs.yaml https://raw.githubusercontent.com/dikkadev/bangs/main/bangs.yaml\n\n")
		fmt.Fprintf(os.Stderr, "2. Then run: go run github.com/dikkadev/bangs/mcp@latest -b ~/bangs.yaml\n\n")
		fmt.Fprintf(os.Stderr, "Or create your own following the format at:\n")
		fmt.Fprintf(os.Stderr, "https://github.com/dikkadev/bangs#configuration-bangsyaml\n")
		os.Exit(1)
	}

	// Load bangs registry
	err := bangs.Load(bangsFile)
	if err != nil {
		slog.Error("Error loading bangs", "err", err)
		os.Exit(1)
	}

	// Setup file watching
	if watchBangFile {
		go watcher.WatchFile(bangsFile, func() error {
			slog.Info("Reloading bangs configuration")
			return bangs.Load(bangsFile)
		})
	}

	// Create MCP server
	var server *mcp.Server
	if httpMode {
		slog.Info("Starting MCP server in HTTP mode", "port", port)
		httpTransport := http.NewHTTPTransport("/mcp").WithAddr(":" + port)
		server = mcp.NewServer(
			httpTransport,
			mcp.WithName("bangs-mcp-server"),
			mcp.WithVersion(version),
			mcp.WithInstructions("Provides access to bangs search functionality via MCP"),
		)
	} else {
		slog.Info("Starting MCP server in stdio mode")
		server = mcp.NewServer(
			stdio.NewStdioServerTransport(),
			mcp.WithName("bangs-mcp-server"),
			mcp.WithVersion(version),
			mcp.WithInstructions("Provides access to bangs search functionality via MCP"),
		)
	}

	// Register tools
	err = registerTools(server)
	if err != nil {
		slog.Error("Error registering tools", "err", err)
		os.Exit(1)
	}

	// Register resources
	err = registerResources(server)
	if err != nil {
		slog.Error("Error registering resources", "err", err)
		os.Exit(1)
	}

	// Start server
	err = server.Serve()
	if err != nil {
		slog.Error("Error starting MCP server", "err", err)
		os.Exit(1)
	}

	// Keep running
	select {}
}

// Tool argument types
type ExecuteBangArgs struct {
	Bang  string `json:"bang" jsonschema:"required,description=The bang to use (e.g. 'gh', 'g')"`
	Query string `json:"query" jsonschema:"required,description=The search query"`
}

type ExecuteMultiBangArgs struct {
	Bangs []string `json:"bangs" jsonschema:"required,description=List of bangs to use"`
	Query string   `json:"query" jsonschema:"required,description=The search query"`
}

type GetBangsByCategoryArgs struct {
	Category string `json:"category" jsonschema:"required,description=The category to filter by"`
}

func registerTools(server *mcp.Server) error {
	// ExecuteBang tool
	err := server.RegisterTool("execute_bang", "Execute a search using a specific bang", func(arguments ExecuteBangArgs) (*mcp.ToolResponse, error) {
		slog.Debug("Executing bang", "bang", arguments.Bang, "query", arguments.Query)

		// Get all bangs to find the one we want
		allBangs := bangs.All()
		var targetEntry *bangs.Entry
		for _, entry := range allBangs.Entries {
			if entry.Bang == arguments.Bang {
				targetEntry = &entry
				break
			}
		}

		if targetEntry == nil {
			return nil, fmt.Errorf("bang '%s' not found", arguments.Bang)
		}

		// Generate the URL
		finalURL, err := targetEntry.URL.Augment(arguments.Query)
		if err != nil {
			return nil, fmt.Errorf("error generating URL: %v", err)
		}

		result := BangResult{
			Bang: targetEntry.Bang,
			Name: findBangName(targetEntry.Bang),
			URL:  finalURL.String(),
		}

		return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("Generated URL: %s", result.URL))), nil
	})
	if err != nil {
		return fmt.Errorf("failed to register execute_bang tool: %v", err)
	}

	// ExecuteMultiBang tool
	err = server.RegisterTool("execute_multi_bang", "Execute a search using multiple bangs", func(arguments ExecuteMultiBangArgs) (*mcp.ToolResponse, error) {
		slog.Debug("Executing multi-bang", "bangs", arguments.Bangs, "query", arguments.Query)

		allBangs := bangs.All()
		var results []BangResult
		var errors []string

		for _, bangName := range arguments.Bangs {
			var targetEntry *bangs.Entry
			for _, entry := range allBangs.Entries {
				if entry.Bang == bangName {
					targetEntry = &entry
					break
				}
			}

			if targetEntry == nil {
				errors = append(errors, fmt.Sprintf("Bang '%s' not found", bangName))
				continue
			}

			finalURL, err := targetEntry.URL.Augment(arguments.Query)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Error generating URL for bang '%s': %v", bangName, err))
				continue
			}

			results = append(results, BangResult{
				Bang: targetEntry.Bang,
				Name: findBangName(targetEntry.Bang),
				URL:  finalURL.String(),
			})
		}

		if len(results) == 0 {
			return nil, fmt.Errorf("no valid bangs found")
		}

		_ = MultiBangResult{
			Results: results,
			Errors:  errors,
		}

		// Format response
		var response strings.Builder
		response.WriteString(fmt.Sprintf("Generated %d URLs:\n", len(results)))
		for _, r := range results {
			response.WriteString(fmt.Sprintf("- %s (%s): %s\n", r.Name, r.Bang, r.URL))
		}
		if len(errors) > 0 {
			response.WriteString(fmt.Sprintf("\nErrors: %s", strings.Join(errors, ", ")))
		}

		return mcp.NewToolResponse(mcp.NewTextContent(response.String())), nil
	})
	if err != nil {
		return fmt.Errorf("failed to register execute_multi_bang tool: %v", err)
	}

	// GetBangsByCategory tool
	err = server.RegisterTool("get_bangs_by_category", "Get all bangs in a specific category", func(arguments GetBangsByCategoryArgs) (*mcp.ToolResponse, error) {
		slog.Debug("Getting bangs by category", "category", arguments.Category)

		allBangs := bangs.All()
		var categoryBangs []BangInfo

		for name, entry := range allBangs.Entries {
			if entry.Category == arguments.Category {
				categoryBangs = append(categoryBangs, BangInfo{
					Name:        name,
					Bang:        entry.Bang,
					Description: entry.Description,
					Category:    entry.Category,
					URL:         string(entry.URL),
				})
			}
		}

		if len(categoryBangs) == 0 {
			return nil, fmt.Errorf("no bangs found in category '%s'", arguments.Category)
		}

		var response strings.Builder
		response.WriteString(fmt.Sprintf("Found %d bangs in category '%s':\n", len(categoryBangs), arguments.Category))
		for _, bang := range categoryBangs {
			response.WriteString(fmt.Sprintf("- %s (!%s): %s\n", bang.Name, bang.Bang, bang.Description))
		}

		return mcp.NewToolResponse(mcp.NewTextContent(response.String())), nil
	})
	if err != nil {
		return fmt.Errorf("failed to register get_bangs_by_category tool: %v", err)
	}

	return nil
}

func registerResources(server *mcp.Server) error {
	// Registry resource
	err := server.RegisterResource("bangs://registry", "bangs_registry", "Complete bangs registry", "application/json", func() (*mcp.ResourceResponse, error) {
		allBangs := bangs.All()

		// Convert to a more structured format
		registry := make(map[string]BangInfo)
		for name, entry := range allBangs.Entries {
			registry[name] = BangInfo{
				Name:        name,
				Bang:        entry.Bang,
				Description: entry.Description,
				Category:    entry.Category,
				URL:         string(entry.URL),
			}
		}

		return mcp.NewResourceResponse(mcp.NewTextEmbeddedResource("bangs://registry", fmt.Sprintf("%+v", registry), "application/json")), nil
	})
	if err != nil {
		return fmt.Errorf("failed to register registry resource: %v", err)
	}

	// Categories resource
	err = server.RegisterResource("bangs://categories", "bangs_categories", "Available bang categories", "application/json", func() (*mcp.ResourceResponse, error) {
		allBangs := bangs.All()

		// Collect unique categories
		categories := make(map[string]bool)
		for _, entry := range allBangs.Entries {
			if entry.Category != "" {
				categories[entry.Category] = true
			}
		}

		var categoryList []string
		for category := range categories {
			categoryList = append(categoryList, category)
		}

		return mcp.NewResourceResponse(mcp.NewTextEmbeddedResource("bangs://categories", fmt.Sprintf("%+v", categoryList), "application/json")), nil
	})
	if err != nil {
		return fmt.Errorf("failed to register categories resource: %v", err)
	}

	return nil
}

// Helper function to find bang name by bang characters
func findBangName(bangChars string) string {
	allBangs := bangs.All()
	for name, entry := range allBangs.Entries {
		if entry.Bang == bangChars {
			return name
		}
	}
	return bangChars // fallback to bang chars if name not found
}
