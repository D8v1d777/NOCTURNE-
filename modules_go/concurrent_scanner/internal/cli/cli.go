package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"nocturne/scanner/internal/correlation"
	"nocturne/scanner/internal/engine"
	"nocturne/scanner/internal/models"
	"nocturne/scanner/internal/validation"
	"nocturne/scanner/sources/external"
	"nocturne/scanner/sources/username"
	"os"
	"strings"
	"time"
)

type CLI struct {
	Manager *engine.Manager
}

func NewCLI() *CLI {
	m := engine.NewManager()
	m.Register(username.NewPlugin())
	m.Register(external.NewPlugin())
	m.Register(external.NewRustModulePlugin())
	return &CLI{Manager: m}
}

func (c *CLI) Run(args []string) {
	if len(args) < 1 {
		c.PrintGeneralHelp()
		return
	}

	command := args[0]
	switch command {
	case "scan":
		c.handleScan(args[1:])
	case "test":
		validation.RunValidationSuite()
	case "monitor":
		c.handleMonitor(args[1:])
	case "worker":
		c.handleWorker(args[1:])
	case "serve":
		c.handleServe(args[1:])
	case "correlate":
		c.handleCorrelate(args[1:])
	case "help":
		c.PrintGeneralHelp()
	case "-h", "--help":
		c.PrintGeneralHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		c.PrintGeneralHelp()
		os.Exit(1)
	}
}

func (c *CLI) handleScan(args []string) {
	if len(args) < 1 {
		c.PrintScanHelp()
		return
	}

	subCommand := args[0]
	if subCommand != "username" {
		fmt.Printf("Unknown scan type: %s\n", subCommand)
		c.PrintScanHelp()
		os.Exit(1)
	}

	// Define flags for the scan username command
	fs := flag.NewFlagSet("scan username", flag.ExitOnError)
	jsonOutput := fs.Bool("json", false, "Output results in JSON format")
	outputFile := fs.String("output", "", "Save results to a file")
	enableExternal := fs.Bool("enable-external", false, "Enable external API-based plugins")
	enableRust := fs.Bool("enable-rust", false, "Enable future Rust-based modules")

	fs.Parse(args[1:])

	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Println("Error: <value> (username) is required")
		c.PrintScanHelp()
		os.Exit(1)
	}

	target := remaining[0]

	// Configuration
	enabled := []string{"username_scanner"}
	if *enableExternal {
		enabled = append(enabled, "external_api")
	}
	if *enableRust {
		enabled = append(enabled, "rust_bridge")
	}

	if !*jsonOutput {
		fmt.Printf("🕯️  NOCTURNE | Starting scan for: %s\n", target)
		fmt.Printf("📦 Plugins: %s\n", strings.Join(enabled, ", "))
		fmt.Println(strings.Repeat("-", 45))
	}

	start := time.Now()
	results := c.Manager.RunPlugins(target, enabled)
	duration := time.Since(start)

	if *jsonOutput {
		c.outputJSON(results, *outputFile)
	} else {
		c.outputTable(results, duration)
		if *outputFile != "" {
			c.saveToFile(results, *outputFile)
		}
	}
}

func (c *CLI) outputTable(results []models.Result, duration time.Duration) {
	fmt.Printf("\n%-15s %-15s %-10s %-10s %-30s\n", "SOURCE", "PLATFORM", "STATUS", "CONF.", "URL")
	fmt.Println(strings.Repeat("-", 90))

	foundCount := 0
	for _, res := range results {
		status := "NOT FOUND"
		if res.Exists {
			status = "MATCH"
			foundCount++
		}
		if res.Error != "" {
			status = "ERROR"
		}

		fmt.Printf("%-15s %-15s %-10s %-10.2f %-30s\n",
			strings.ToUpper(res.Source),
			res.Platform,
			status,
			res.Confidence,
			res.URL)

		if res.Error != "" {
			fmt.Printf("  [!] Error: %s\n", res.Error)
		}
	}

	fmt.Println(strings.Repeat("-", 90))
	fmt.Printf("✨ Scan completed in %s. Found %d matches.\n", duration.Round(time.Millisecond), foundCount)
}

func (c *CLI) outputJSON(results []models.Result, filePath string) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return
	}

	if filePath != "" {
		err := os.WriteFile(filePath, data, 0644)
		if err != nil {
			fmt.Printf("Error saving JSON to file: %v\n", err)
			return
		}
		fmt.Printf("Results saved to %s\n", filePath)
	} else {
		fmt.Println(string(data))
	}
}

func (c *CLI) saveToFile(results []models.Result, filePath string) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("NOCTURNE SCAN REPORT - %s\n", time.Now().Format(time.RFC3339)))
	sb.WriteString(strings.Repeat("=", 45) + "\n\n")

	for _, res := range results {
		exists := "No"
		if res.Exists {
			exists = "Yes"
		}
		sb.WriteString(fmt.Sprintf("Platform: %s\nURL: %s\nExists: %s\nConfidence: %.2f\nSource: %s\n",
			res.Platform, res.URL, exists, res.Confidence, res.Source))
		if res.Error != "" {
			sb.WriteString(fmt.Sprintf("Error: %s\n", res.Error))
		}
		sb.WriteString("-" + "\n")
	}

	err := os.WriteFile(filePath, []byte(sb.String()), 0644)
	if err != nil {
		fmt.Printf("Error saving report to file: %v\n", err)
		return
	}
	fmt.Printf("Report saved to %s\n", filePath)
}

func (c *CLI) PrintGeneralHelp() {
	fmt.Println(`🕯️  NOCTURNE - Professional OSINT Scanner

Usage:
  nocturne <command> [subcommand] [flags] [value]

Commands:
  scan        Execute a scanning operation
  correlate   Run identity correlation on sample/input data
  help        Show this help message

Examples:
  nocturne scan username shadow_user --enable-rust
  nocturne correlate
  nocturne scan username shadow_user --json --output report.json`)
}

func (c *CLI) handleCorrelate(args []string) {
	fmt.Println("🧠 Running NOCTURNE Correlation Engine (Graph Mode)...")
	fmt.Println(strings.Repeat("-", 45))

	// Sample data for demonstration
	identities := []correlation.Identity{
		{
			ID: "1", Platform: "GitHub", Username: "shadow_coder",
			DisplayName: "Shadow Coder", Bio: "Security researcher and Go enthusiast.",
			Links: []string{"https://shadow.io"},
		},
		{
			ID: "2", Platform: "Twitter", Username: "shadow_coder",
			DisplayName: "Shadow", Bio: "I code in Go and look for bugs.",
			Links: []string{"https://shadow.io"},
		},
		{
			ID: "3", Platform: "Reddit", Username: "shadow_coder_alt",
			DisplayName: "Random", Bio: "Security researcher and developer. Focused on Go.",
		},
		{
			ID: "4", Platform: "Instagram", Username: "shadow_coder",
			DisplayName: "Shadow", Bio: "Photography and travel.",
		},
		{
			ID: "5", Platform: "Mastodon", Username: "shadow_coder",
			DisplayName: "Shadow Coder", Bio: "Security and Privacy.",
			Links:     []string{"https://shadow.io"},
			AvatarURL: "https://example.com/avatar1.png",
		},
		{
			ID: "6", Platform: "Bluesky", Username: "shadow_dev",
			DisplayName: "Shadow", Bio: "I build things.",
			AvatarURL: "https://example.com/avatar1.png",
		},
	}

	// 1. Show pairwise comparison with rejection reasons
	fmt.Println("\n🔍 Pairwise Comparison Analysis (Graph Edge Candidates):")
	for i := 0; i < len(identities); i++ {
		for j := i + 1; j < len(identities); j++ {
			a := correlation.NormalizeIdentity(identities[i])
			b := correlation.NormalizeIdentity(identities[j])
			score, reasons, rejections := correlation.Compare(a, b)

			if score >= 0.65 {
				fmt.Printf("\n🔗 EDGE: [%s] <-> [%s] (Weight: %.2f)\n", a.Username, b.Username, score)
				fmt.Printf("   Signals: %s\n", strings.Join(reasons, ", "))
			} else if len(rejections) > 0 {
				// Only show interesting rejections to avoid noise
				if !strings.Contains(rejections[0], "no strong signal") {
					fmt.Printf("\n🚫 NO EDGE: [%s] <-> [%s] (Score: %.2f)\n", a.Username, b.Username, score)
					fmt.Printf("   Rejections: %s\n", strings.Join(rejections, ", "))
				}
			}
		}
	}

	// 2. Show final clustering
	fmt.Println("\n" + strings.Repeat("-", 45))
	fmt.Println("📦 Graph-Based Identity Clusters (Connected Components):")
	clusters, _ := correlation.RunCorrelation(identities) // Ignore edges for CLI output

	for i, cluster := range clusters {
		if len(cluster.Members) > 1 {
			fmt.Printf("\n[Cluster %d] Level: %s (%.2f)\n", i+1, cluster.ConfidenceLevel, cluster.Confidence)
			fmt.Printf("   Analysis: %s\n", cluster.ConfidenceExplain)
		} else {
			fmt.Printf("\n[Cluster %d] Level: %s\n", i+1, cluster.ConfidenceLevel)
		}

		fmt.Println("Members:")
		for _, m := range cluster.Members {
			fmt.Printf("  - %-15s (%s)\n", m.Username, m.Platform)
		}
	}
}

func (c *CLI) PrintScanHelp() {
	fmt.Println(`Usage:
  nocturne scan username <value> [flags]

Flags:
  --json             Output results in JSON format
  --output <file>    Save results to a specific file
  --enable-external  Enable external API-based plugins (default: false)
  --enable-rust      Enable future Rust-based modules (default: false)`)
}
