package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codeownersnet/atlas/internal/auth"
	"github.com/codeownersnet/atlas/internal/config"
	"github.com/codeownersnet/atlas/internal/mcp"
	confluencetools "github.com/codeownersnet/atlas/internal/tools/confluence"
	jiratools "github.com/codeownersnet/atlas/internal/tools/jira"
	opsgenietools "github.com/codeownersnet/atlas/internal/tools/opsgenie"
	"github.com/codeownersnet/atlas/pkg/atlassian/confluence"
	"github.com/codeownersnet/atlas/pkg/atlassian/jira"
	"github.com/codeownersnet/atlas/pkg/atlassian/opsgenie"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	version    = "0.1.0"
	commit     = "dev"
	date       = "unknown"
	rootCmd    *cobra.Command
	configFile string
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "atlas-mcp",
		Short: "MCP server for Atlassian products (Jira, Confluence, and Opsgenie)",
		Long: `MCP Atlassian is a Model Context Protocol server that provides AI assistants
with seamless access to Atlassian products (Jira and Confluence).

Supports both Cloud and Server/Data Center deployments with multiple
authentication methods: API Token, Personal Access Token, and Bearer Token.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(configFile)
		},
	}

	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file (default is .env)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(configFile string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Setup logging
	logger := setupLogger(cfg.Logging)

	logger.Info().
		Str("version", version).
		Str("commit", commit).
		Str("transport", cfg.Server.Transport).
		Bool("read_only_mode", cfg.Security.ReadOnlyMode).
		Msg("starting MCP Atlassian server")

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.ServerConfig{
		Logger:       &logger,
		ReadOnlyMode: cfg.Security.ReadOnlyMode,
		EnabledTools: cfg.Security.EnabledTools,
	})

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Jira client and register tools if configured
	if cfg.IsJiraConfigured() {
		logger.Info().
			Str("url", cfg.Jira.URL).
			Str("auth_method", cfg.Jira.AuthMethod.String()).
			Msg("initializing Jira client")

		jiraClient, err := createJiraClient(cfg, &logger)
		if err != nil {
			return fmt.Errorf("failed to create Jira client: %w", err)
		}

		// Store Jira client in context
		ctx = jiratools.WithJiraClient(ctx, jiraClient)

		// Register all Jira tools
		if err := jiratools.RegisterJiraTools(mcpServer); err != nil {
			return fmt.Errorf("failed to register Jira tools: %w", err)
		}

		logger.Info().Int("count", 29).Msg("registered Jira tools")
	} else {
		logger.Info().Msg("Jira not configured, skipping Jira tools")
	}

	// Initialize Confluence client and register tools if configured
	if cfg.IsConfluenceConfigured() {
		logger.Info().
			Str("url", cfg.Confluence.URL).
			Str("auth_method", cfg.Confluence.AuthMethod.String()).
			Msg("initializing Confluence client")

		confluenceClient, err := createConfluenceClient(cfg, &logger)
		if err != nil {
			return fmt.Errorf("failed to create Confluence client: %w", err)
		}

		// Store Confluence client in context
		ctx = confluencetools.WithConfluenceClient(ctx, confluenceClient)

		// Register all Confluence tools
		if err := confluencetools.RegisterConfluenceTools(mcpServer); err != nil {
			return fmt.Errorf("failed to register Confluence tools: %w", err)
		}

		logger.Info().Int("count", 11).Msg("registered Confluence tools")
	} else {
		logger.Info().Msg("Confluence not configured, skipping Confluence tools")
	}

	// Initialize Opsgenie client and register tools if configured
	if cfg.IsOpsgenieConfigured() {
		logger.Info().
			Str("url", cfg.Opsgenie.URL).
			Msg("initializing Opsgenie client")

		opsgenieClient, err := createOpsgenieClient(cfg, &logger)
		if err != nil {
			return fmt.Errorf("failed to create Opsgenie client: %w", err)
		}

		// Store Opsgenie client in context
		ctx = opsgenietools.WithOpsgenieClient(ctx, opsgenieClient)

		// Register all Opsgenie tools
		if err := opsgenietools.RegisterOpsgenieTools(mcpServer); err != nil {
			return fmt.Errorf("failed to register Opsgenie tools: %w", err)
		}

		logger.Info().Int("count", 25).Msg("registered Opsgenie tools")
	} else {
		logger.Info().Msg("Opsgenie not configured, skipping Opsgenie tools")
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info().Str("signal", sig.String()).Msg("received shutdown signal")
		cancel()
	}()

	// Start the appropriate transport
	switch cfg.Server.Transport {
	case "stdio":
		return runStdioTransport(ctx, mcpServer, &logger)
	case "sse":
		return fmt.Errorf("SSE transport not yet implemented")
	case "streamable-http":
		return fmt.Errorf("streamable-http transport not yet implemented")
	default:
		return fmt.Errorf("unknown transport: %s", cfg.Server.Transport)
	}
}

func runStdioTransport(ctx context.Context, server *mcp.Server, logger *zerolog.Logger) error {
	logger.Info().Msg("starting stdio transport")

	transport := mcp.NewStdioTransport(server, logger)

	if err := transport.Start(ctx); err != nil {
		if err == context.Canceled {
			logger.Info().Msg("stdio transport stopped gracefully")
			return nil
		}
		return fmt.Errorf("stdio transport error: %w", err)
	}

	return nil
}

func setupLogger(cfg *config.LoggingConfig) zerolog.Logger {
	// Determine log level
	level := zerolog.WarnLevel
	if cfg.VeryVerbose {
		level = zerolog.DebugLevel
	} else if cfg.Verbose {
		level = zerolog.InfoLevel
	}

	// Determine output stream
	output := os.Stderr
	if cfg.LogToStdout {
		output = os.Stdout
	}

	// Create logger with human-friendly console output
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        output,
		TimeFormat: time.RFC3339,
	}).
		Level(level).
		With().
		Timestamp().
		Logger()

	return logger
}

// createJiraClient creates a Jira client with the appropriate authentication
func createJiraClient(cfg *config.Config, logger *zerolog.Logger) (*jira.Client, error) {
	authProvider, err := createJiraAuthProvider(cfg.Jira)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}

	logger.Debug().
		Str("auth_type", authProvider.Type()).
		Str("auth_masked", authProvider.Mask()).
		Msg("created Jira auth provider")

	jiraClient, err := jira.NewClient(&jira.Config{
		BaseURL:       cfg.Jira.URL,
		Auth:          authProvider,
		CustomHeaders: cfg.Jira.CustomHeaders,
		SSLVerify:     cfg.Jira.SSLVerify,
		HTTPProxy:     cfg.Jira.HTTPProxy,
		HTTPSProxy:    cfg.Jira.HTTPSProxy,
		SOCKSProxy:    cfg.Jira.SOCKSProxy,
		NoProxy:       cfg.Jira.NoProxy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Jira client: %w", err)
	}

	return jiraClient, nil
}

// createJiraAuthProvider creates the appropriate auth provider for Jira
func createJiraAuthProvider(cfg *config.JiraConfig) (auth.Provider, error) {
	switch cfg.AuthMethod {
	case config.AuthMethodBasic:
		return auth.NewBasicAuth(cfg.Username, cfg.APIToken)
	case config.AuthMethodPAT:
		return auth.NewPATAuth(cfg.PersonalToken)
	case config.AuthMethodOAuth:
		// BYO (Bring Your Own) OAuth token
		if cfg.OAuthAccessToken == "" {
			return nil, fmt.Errorf("ATLASSIAN_OAUTH_ACCESS_TOKEN is required for OAuth authentication")
		}
		return auth.NewOAuthAuth(cfg.OAuthAccessToken, cfg.OAuthCloudID)
	default:
		return nil, fmt.Errorf("no authentication configured - set JIRA_USERNAME+JIRA_API_TOKEN or JIRA_PERSONAL_TOKEN or ATLASSIAN_OAUTH_ACCESS_TOKEN")
	}
}

// createConfluenceClient creates a Confluence client with the appropriate authentication
func createConfluenceClient(cfg *config.Config, logger *zerolog.Logger) (*confluence.Client, error) {
	authProvider, err := createConfluenceAuthProvider(cfg.Confluence)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}

	logger.Debug().
		Str("auth_type", authProvider.Type()).
		Str("auth_masked", authProvider.Mask()).
		Msg("created Confluence auth provider")

	confluenceClient, err := confluence.NewClient(&confluence.Config{
		BaseURL:       cfg.Confluence.URL,
		Auth:          authProvider,
		CustomHeaders: cfg.Confluence.CustomHeaders,
		SSLVerify:     cfg.Confluence.SSLVerify,
		HTTPProxy:     cfg.Confluence.HTTPProxy,
		HTTPSProxy:    cfg.Confluence.HTTPSProxy,
		SOCKSProxy:    cfg.Confluence.SOCKSProxy,
		NoProxy:       cfg.Confluence.NoProxy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Confluence client: %w", err)
	}

	return confluenceClient, nil
}

// createConfluenceAuthProvider creates the appropriate auth provider for Confluence
func createConfluenceAuthProvider(cfg *config.ConfluenceConfig) (auth.Provider, error) {
	switch cfg.AuthMethod {
	case config.AuthMethodBasic:
		return auth.NewBasicAuth(cfg.Username, cfg.APIToken)
	case config.AuthMethodPAT:
		return auth.NewPATAuth(cfg.PersonalToken)
	case config.AuthMethodOAuth:
		// BYO (Bring Your Own) OAuth token
		if cfg.OAuthAccessToken == "" {
			return nil, fmt.Errorf("ATLASSIAN_OAUTH_ACCESS_TOKEN is required for OAuth authentication")
		}
		return auth.NewOAuthAuth(cfg.OAuthAccessToken, cfg.OAuthCloudID)
	default:
		return nil, fmt.Errorf("no authentication configured - set CONFLUENCE_USERNAME+CONFLUENCE_API_TOKEN or CONFLUENCE_PERSONAL_TOKEN or ATLASSIAN_OAUTH_ACCESS_TOKEN")
	}
}

// createOpsgenieClient creates an Opsgenie client with the appropriate authentication
func createOpsgenieClient(cfg *config.Config, logger *zerolog.Logger) (*opsgenie.Client, error) {
	authProvider, err := createOpsgenieAuthProvider(cfg.Opsgenie)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}

	logger.Debug().
		Str("auth_type", authProvider.Type()).
		Str("auth_masked", authProvider.Mask()).
		Msg("created Opsgenie auth provider")

	opsgenieClient, err := opsgenie.NewClient(&opsgenie.Config{
		BaseURL:       cfg.Opsgenie.URL,
		Auth:          authProvider,
		CustomHeaders: cfg.Opsgenie.CustomHeaders,
		SSLVerify:     cfg.Opsgenie.SSLVerify,
		HTTPProxy:     cfg.Opsgenie.HTTPProxy,
		HTTPSProxy:    cfg.Opsgenie.HTTPSProxy,
		SOCKSProxy:    cfg.Opsgenie.SOCKSProxy,
		NoProxy:       cfg.Opsgenie.NoProxy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Opsgenie client: %w", err)
	}

	return opsgenieClient, nil
}

// createOpsgenieAuthProvider creates the appropriate auth provider for Opsgenie
func createOpsgenieAuthProvider(cfg *config.OpsgenieConfig) (auth.Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OPSGENIE_API_KEY is required for Opsgenie authentication")
	}
	return auth.NewAPIKeyAuth(cfg.APIKey)
}
