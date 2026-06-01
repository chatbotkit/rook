// Command rook is a standalone autonomous security agent for vulnerability
// research, bug hunting and source-code auditing. It is built on the
// ChatBotKit Go SDK and ships with an embedded library of security skills.
//
// Usage:
//
//	export CHATBOTKIT_API_SECRET="your-api-key"
//	rook "Audit the HTTP handlers in ./server for injection bugs"
//	rook --scope "repo: ./server, no network" "Hunt for auth bypasses"
//	rook version
//
// Rook is intended for authorized security testing only. Always pass an
// explicit --scope describing the systems you are permitted to assess.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"

	"github.com/chatbotkit/rook/internal/agent"
	"github.com/chatbotkit/rook/internal/config"
	"github.com/chatbotkit/rook/internal/version"
)

func main() {
	godotenv.Load()

	flags := pflag.NewFlagSet("rook", pflag.ContinueOnError)
	model := flags.String("model", config.DefaultModel, "model the agent reasons with")
	maxIter := flags.Int("max-iterations", config.DefaultMaxIterations, "maximum agent iterations before forced stop")
	scope := flags.String("scope", "", "authorization boundary (hosts, repos, paths) the agent must stay within")
	scopeFile := flags.String("scope-file", "", "read the authorization scope from a file")
	verbose := flags.BoolP("verbose", "v", false, "stream the agent's reasoning tokens to stdout")
	showVersion := flags.BoolP("version", "V", false, "print version and exit")

	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "rook — autonomous security research agent\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  rook [flags] <task>\n  rook version\n\nFlags:\n")
		flags.PrintDefaults()
	}

	// Allow `rook version` as a subcommand in addition to the --version flag.
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "version" {
		printVersion()
		return
	}

	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	if *showVersion {
		printVersion()
		return
	}

	task := strings.TrimSpace(strings.Join(flags.Args(), " "))
	if task == "" {
		flags.Usage()
		os.Exit(2)
	}

	apiSecret := os.Getenv("CHATBOTKIT_API_SECRET")
	if apiSecret == "" {
		fmt.Fprintln(os.Stderr, "Error: CHATBOTKIT_API_SECRET environment variable is not set")
		os.Exit(1)
	}

	resolvedScope := *scope
	if *scopeFile != "" {
		data, err := os.ReadFile(*scopeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: read scope file: %v\n", err)
			os.Exit(1)
		}
		resolvedScope = string(data)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	code, err := agent.Run(ctx, agent.Config{
		APISecret:     apiSecret,
		Model:         *model,
		MaxIterations: *maxIter,
		Task:          task,
		Scope:         resolvedScope,
		Verbose:       *verbose,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}

	notifyUpdate()

	os.Exit(code)
}

func printVersion() {
	fmt.Printf("rook %s\n", version.Version)
	notifyUpdate()
}

// notifyUpdate prints a one-line notice to stderr when a newer release exists.
// It is silently skipped for dev builds and on any network error.
func notifyUpdate() {
	result, err := version.Check()
	if err != nil {
		return
	}
	if notice := version.FormatUpdateNotice(result); notice != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", notice)
	}
}
