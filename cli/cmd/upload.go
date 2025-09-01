package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/everFinance/goar/utils"
	"github.com/project-kardeshev/go-ardrive-turbo/pkg/signers"
	"github.com/project-kardeshev/go-ardrive-turbo/pkg/turbo"
	"github.com/project-kardeshev/go-ardrive-turbo/pkg/types"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/pkg/wasm"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tui"
	"gopkg.in/yaml.v2"
)

// BuildConfig represents the structure of the build configuration file
type BuildConfig struct {
	ComputeLimit    string `yaml:"compute_limit"`
	MaximumMemory   int    `yaml:"maximum_memory"`
	ModuleFormat    string `yaml:"module_format"`
	AOSGitHash      string `yaml:"aos_git_hash"`
	DataProtocol    string `yaml:"data_protocol,omitempty"`
	Variant         string `yaml:"variant,omitempty"`
	Type            string `yaml:"type,omitempty"`
	InputEncoding   string `yaml:"input_encoding,omitempty"`
	OutputEncoding  string `yaml:"output_encoding,omitempty"`
	ContentType     string `yaml:"content_type,omitempty"`
	AppName         string `yaml:"app_name,omitempty"`
	AppVersion      string `yaml:"app_version,omitempty"`
	Author          string `yaml:"author,omitempty"`
}

// HandleUploadCommand handles the upload command for uploading built modules to Arweave
func HandleUploadCommand(ctx context.Context, args []string) {
	debug.Printf("Handling upload command with args: %v", args)

	// If no arguments provided, run interactive mode
	if len(args) == 0 {
		err := runInteractiveUpload(ctx)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	var dryRun bool
	var walletPath string
	var wasmPath string
	var configPath string
	var version string
	var gitHash string

	// Parse command line arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			dryRun = true
		case "--wallet-file", "-w":
			if i+1 < len(args) {
				walletPath = args[i+1]
				i++
			}
		case "--wasm-file", "-f":
			if i+1 < len(args) {
				wasmPath = args[i+1]
				i++
			}
		case "--config", "-c":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "--version", "-v":
			if i+1 < len(args) {
				version = args[i+1]
				i++
			}
		case "--git-hash", "-g":
			if i+1 < len(args) {
				gitHash = args[i+1]
				i++
			}
		case "--help", "-h":
			PrintUploadUsage()
			return
		default:
			if !strings.HasPrefix(args[i], "-") && wasmPath == "" {
				wasmPath = args[i]
			}
		}
	}

	// Set default paths
	if wasmPath == "" {
		wasmPath = "dist/process.wasm"
	}
	if configPath == "" {
		// Try common config locations
		if _, err := os.Stat(".harlequin.yaml"); err == nil {
			configPath = ".harlequin.yaml"
		} else if _, err := os.Stat("build_configs/ao-build-config.yml"); err == nil {
			configPath = "build_configs/ao-build-config.yml"
		} else if _, err := os.Stat("ao-build-config.yml"); err == nil {
			configPath = "ao-build-config.yml"
		} else {
			fmt.Println("Error: No configuration file found. Please specify with --config or create .harlequin.yaml")
			PrintUploadUsage()
			os.Exit(1)
		}
	}
	if walletPath == "" {
		if os.Getenv("WALLET_PATH") != "" {
			walletPath = os.Getenv("WALLET_PATH")
		} else {
			walletPath = "key.json"
		}
	}
	if version == "" {
		version = "dev"
	}
	if gitHash == "" {
		gitHash = os.Getenv("GITHUB_SHA")
		if gitHash == "" {
			gitHash = "no-git-hash"
		}
	}

	err := uploadModule(ctx, wasmPath, configPath, walletPath, version, gitHash, dryRun)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// runInteractiveUpload runs the interactive TUI for module upload
func runInteractiveUpload(ctx context.Context) error {
	// Use the main TUI which has the enhanced balance checking
	return runMainTUIForUpload(ctx)
}

// runMainTUIForUpload starts the main TUI specifically for upload workflow
func runMainTUIForUpload(ctx context.Context) error {
	return tui.RunBuildTUI(ctx)
}

// winstonToCredits converts Winston Credits to Credits (AR denomination)
func winstonToCredits(winston string) string {
	if winston == "" || winston == "0" {
		return "0"
	}

	// Convert string to big.Int for precision
	winstonBig := new(big.Int)
	winstonBig, ok := winstonBig.SetString(winston, 10)
	if !ok {
		debug.Printf("Error parsing winston string: %s", winston)
		return winston + " Winston" // Fallback to raw winston
	}

	// Convert Winston to AR
	ar := utils.WinstonToAR(winstonBig)

	// Format with reasonable precision (6 decimal places)
	return ar.Text('f', 6)
}

// formatCreditsDisplay formats credits with appropriate units
func formatCreditsDisplay(winston string) string {
	credits := winstonToCredits(winston)
	if credits == "0" {
		return "0 Credits"
	}
	return credits + " Credits"
}

// uploadModule uploads the WASM module to Arweave using the Turbo client
func uploadModule(ctx context.Context, wasmPath, configPath, walletPath, version, gitHash string, dryRun bool) error {
	// Read WASM binary
	wasmBinary, err := ioutil.ReadFile(wasmPath)
	if err != nil {
		return fmt.Errorf("failed to read WASM file %s: %w", wasmPath, err)
	}

	fmt.Println("🎭 ===============================================")
	fmt.Println("   HARLEQUIN MODULE UPLOAD")
	fmt.Println("   ===============================================")
	fmt.Println()

	// File Analysis Section
	fmt.Println("📦 FILE ANALYSIS")
	fmt.Println("   ───────────────")
	fmt.Printf("   • WASM File: %s\n", wasmPath)
	fmt.Printf("   • File Size: %s (%d bytes)\n", wasm.FormatMemorySize(uint32(len(wasmBinary))), len(wasmBinary))

	// Parse WASM metadata
	wasmInfo, err := wasm.ParseWasmBinary(wasmBinary)
	if err != nil {
		fmt.Printf("   ⚠️  Could not parse WASM metadata: %v\n", err)
		wasmInfo = nil
	} else {
		fmt.Println("   • WASM Metadata:")
		if wasmInfo.InitialMemory > 0 {
			fmt.Printf("     - Initial Memory: %s\n", wasm.FormatMemorySize(wasmInfo.InitialMemory))
		}
		if wasmInfo.MaxMemory > 0 {
			fmt.Printf("     - Maximum Memory: %s\n", wasm.FormatMemorySize(wasmInfo.MaxMemory))
		}
		if wasmInfo.StackSize > 0 {
			fmt.Printf("     - Stack Size: %s\n", wasm.FormatMemorySize(wasmInfo.StackSize))
		}
		fmt.Printf("     - Target: %s\n", wasmInfo.Target)

		// Display runtime structure
		fmt.Printf("     - Functions: %d", wasmInfo.FunctionCount)
		if len(wasmInfo.Imports) > 0 {
			importFuncs := 0
			for _, imp := range wasmInfo.Imports {
				if imp.Type == "function" {
					importFuncs++
				}
			}
			if importFuncs > 0 {
				fmt.Printf(" (%d imported)", importFuncs)
			}
		}
		fmt.Printf("\n")

		if wasmInfo.GlobalCount > 0 {
			fmt.Printf("     - Globals: %d\n", wasmInfo.GlobalCount)
		}
		if wasmInfo.TableCount > 0 {
			fmt.Printf("     - Tables: %d\n", wasmInfo.TableCount)
		}

		// Display key exports
		if len(wasmInfo.Exports) > 0 {
			funcExports := make([]string, 0)
			memoryExports := make([]string, 0)
			otherExports := make([]string, 0)

			for _, exp := range wasmInfo.Exports {
				switch exp.Type {
				case "function":
					funcExports = append(funcExports, exp.Name)
				case "memory":
					memoryExports = append(memoryExports, exp.Name)
				default:
					otherExports = append(otherExports, fmt.Sprintf("%s (%s)", exp.Name, exp.Type))
				}
			}

			fmt.Printf("     - Exports: %d total\n", len(wasmInfo.Exports))
			if len(funcExports) > 0 {
				fmt.Printf("       • Functions: %s", funcExports[0])
				if len(funcExports) > 1 {
					fmt.Printf(" (+%d more)", len(funcExports)-1)
				}
				fmt.Printf("\n")
			}
			if len(memoryExports) > 0 {
				fmt.Printf("       • Memory: %s\n", memoryExports[0])
			}
			if len(otherExports) > 0 {
				fmt.Printf("       • Other: %s", otherExports[0])
				if len(otherExports) > 1 {
					fmt.Printf(" (+%d more)", len(otherExports)-1)
				}
				fmt.Printf("\n")
			}
		}

		if len(wasmInfo.CustomSections) > 0 {
			fmt.Printf("     - Custom Sections: %d found\n", len(wasmInfo.CustomSections))
		}
	}
	fmt.Println()

	// Configuration Section
	fmt.Println("📋 CONFIGURATION")
	fmt.Println("   ─────────────")
	configContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config BuildConfig
	err = yaml.Unmarshal(configContent, &config)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	fmt.Printf("   • Config File: %s\n", configPath)
	fmt.Printf("   • Compute Limit: %s\n", config.ComputeLimit)
	fmt.Printf("   • Module Format: %s\n", config.ModuleFormat)
	if config.AOSGitHash != "" {
		fmt.Printf("   • AOS Git Hash: %s\n", config.AOSGitHash)
	}
	fmt.Println()

	// Create publishing tags - prefer WASM metadata over config where available
	publishingTags := map[string]string{
		// AO tags
		"Data-Protocol":    getOrDefault(config.DataProtocol, "ao"),
		"Variant":          getOrDefault(config.Variant, "ao.TN.1"),
		"Type":             getOrDefault(config.Type, "Module"),
		"Input-Encoding":   getOrDefault(config.InputEncoding, "JSON-1"),
		"Output-Encoding":  getOrDefault(config.OutputEncoding, "JSON-1"),
		"Content-Type":     getOrDefault(config.ContentType, "application/wasm"),
		"Compute-Limit":    config.ComputeLimit,
		"Module-Format":    config.ModuleFormat,
		// Extra tags
		"App-Name":         getOrDefault(config.AppName, "Harlequin-CLI"),
		"App-Version":      getOrDefault(config.AppVersion, "1.0.0"),
		"AO-Module-Version": version,
		"Author":           getOrDefault(config.Author, "Harlequin Toolkit"),
		"Git-Hash":         gitHash,
		"AOS-Git-Hash":     config.AOSGitHash,
	}

	// Use WASM metadata for memory configuration if available
	if wasmInfo != nil {
		if wasmInfo.MaxMemory > 0 {
			publishingTags["Memory-Limit"] = strconv.Itoa(int(wasmInfo.MaxMemory)) + "-b"
		} else if config.MaximumMemory > 0 {
			publishingTags["Memory-Limit"] = strconv.Itoa(config.MaximumMemory) + "-b"
		}

		if wasmInfo.InitialMemory > 0 {
			publishingTags["Initial-Memory"] = strconv.Itoa(int(wasmInfo.InitialMemory)) + "-b"
		}

		if wasmInfo.StackSize > 0 {
			publishingTags["Stack-Size"] = strconv.Itoa(int(wasmInfo.StackSize)) + "-b"
		}

		// Update target based on WASM metadata
		if wasmInfo.Target != "" {
			publishingTags["Target"] = wasmInfo.Target
		}

		// Add runtime metadata as tags
		if wasmInfo.FunctionCount > 0 {
			publishingTags["Function-Count"] = strconv.Itoa(int(wasmInfo.FunctionCount))
		}
		if len(wasmInfo.Exports) > 0 {
			publishingTags["Export-Count"] = strconv.Itoa(len(wasmInfo.Exports))

			// Create JSON arrays for detailed export information
			funcExports := make([]string, 0)
			memoryExports := make([]string, 0)
			globalExports := make([]string, 0)
			tableExports := make([]string, 0)

			for _, exp := range wasmInfo.Exports {
				switch exp.Type {
				case "function":
					funcExports = append(funcExports, exp.Name)
				case "memory":
					memoryExports = append(memoryExports, exp.Name)
				case "global":
					globalExports = append(globalExports, exp.Name)
				case "table":
					tableExports = append(tableExports, exp.Name)
				}
			}

			// Add JSON arrays as tags
			if len(funcExports) > 0 {
				funcJSON, _ := json.Marshal(funcExports)
				publishingTags["Exported-Functions"] = string(funcJSON)

				// Prioritize key AO and WASM runtime exports
				keyExports := make([]string, 0)
				priorityFuncs := []string{"handle", "main", "__wasm_call_ctors", "malloc", "free"}

				// Add priority functions if they exist
				for _, priority := range priorityFuncs {
					for _, exp := range funcExports {
						if exp == priority {
							keyExports = append(keyExports, exp)
							break
						}
					}
				}

				// Fill remaining slots with other exports
				for _, exp := range funcExports {
					if len(keyExports) >= 3 {
						break
					}
					found := false
					for _, key := range keyExports {
						if key == exp {
							found = true
							break
						}
					}
					if !found {
						keyExports = append(keyExports, exp)
					}
				}

				if len(keyExports) > 0 {
					publishingTags["Key-Exports"] = strings.Join(keyExports, ",")
				}
			}

			if len(globalExports) > 0 {
				globalJSON, _ := json.Marshal(globalExports)
				publishingTags["Exported-Globals"] = string(globalJSON)
			}

			if len(memoryExports) > 0 {
				memoryJSON, _ := json.Marshal(memoryExports)
				publishingTags["Exported-Memory"] = string(memoryJSON)
			}

			if len(tableExports) > 0 {
				tableJSON, _ := json.Marshal(tableExports)
				publishingTags["Exported-Tables"] = string(tableJSON)
			}
		}
	} else {
		// Fallback to config memory limit
		if config.MaximumMemory > 0 {
			publishingTags["Memory-Limit"] = strconv.Itoa(config.MaximumMemory) + "-b"
		}
	}

	// Remove empty values
	for key, value := range publishingTags {
		if value == "" {
			delete(publishingTags, key)
		}
	}

	// Tags Section
	fmt.Println("🏷️  UPLOAD TAGS")
	fmt.Println("   ──────────")

	// Organize tags by category
	aaTags := []string{"Data-Protocol", "Variant", "Type", "Input-Encoding", "Output-Encoding", "Content-Type"}
	memoryTags := []string{"Memory-Limit", "Initial-Memory", "Stack-Size", "Compute-Limit"}
	buildTags := []string{"Module-Format", "Target", "AO-Module-Version", "Git-Hash", "AOS-Git-Hash"}
	runtimeTags := []string{"Function-Count", "Export-Count", "Key-Exports", "Exported-Functions", "Exported-Globals", "Exported-Memory", "Exported-Tables"}
	appTags := []string{"App-Name", "App-Version", "Author"}

	fmt.Println("   • AO Protocol:")
	for _, tag := range aaTags {
		if value, exists := publishingTags[tag]; exists {
			fmt.Printf("     - %s: %s\n", tag, value)
		}
	}

	fmt.Println("   • Memory & Performance:")
	for _, tag := range memoryTags {
		if value, exists := publishingTags[tag]; exists {
			if strings.HasSuffix(value, "-b") {
				// Format byte values nicely
				if bytes, err := strconv.Atoi(strings.TrimSuffix(value, "-b")); err == nil {
					fmt.Printf("     - %s: %s\n", tag, wasm.FormatMemorySize(uint32(bytes)))
				} else {
					fmt.Printf("     - %s: %s\n", tag, value)
				}
			} else {
				fmt.Printf("     - %s: %s\n", tag, value)
			}
		}
	}

	fmt.Println("   • Build Information:")
	for _, tag := range buildTags {
		if value, exists := publishingTags[tag]; exists {
			fmt.Printf("     - %s: %s\n", tag, value)
		}
	}

	fmt.Println("   • Runtime Metadata:")
	for _, tag := range runtimeTags {
		if value, exists := publishingTags[tag]; exists {
			// Format JSON arrays nicely for display
			if strings.HasPrefix(tag, "Exported-") && strings.HasPrefix(value, "[") {
				// It's a JSON array, format it better
				var items []string
				if err := json.Unmarshal([]byte(value), &items); err == nil {
					if len(items) <= 3 {
						fmt.Printf("     - %s: %s\n", tag, value)
					} else {
						// Show first few items + count for long arrays
						preview := fmt.Sprintf("[\"%s\", \"%s\", \"%s\" ... +%d more]",
							items[0], items[1], items[2], len(items)-3)
						fmt.Printf("     - %s: %s\n", tag, preview)
						fmt.Printf("       (Full list: %s)\n", value)
					}
				} else {
					fmt.Printf("     - %s: %s\n", tag, value)
				}
			} else {
				fmt.Printf("     - %s: %s\n", tag, value)
			}
		}
	}

	fmt.Println("   • Application:")
	for _, tag := range appTags {
		if value, exists := publishingTags[tag]; exists {
			fmt.Printf("     - %s: %s\n", tag, value)
		}
	}
	fmt.Println()

	if dryRun {
		fmt.Println("🌵 DRY RUN MODE")
		fmt.Println("   ────────────")
		fmt.Println("   • No actual upload will be performed")
		fmt.Println("   • All metadata and tags have been generated")
		fmt.Println("   • Ready for production upload")
		fmt.Println()
		fmt.Println("✅ Dry run completed successfully!")
		return nil
	}

	// Wallet Section
	fmt.Println("🔑 WALLET & AUTHENTICATION")
	fmt.Println("   ──────────────────────")

	// Read wallet (only needed for actual uploads)
	var jwk map[string]interface{}
	if os.Getenv("WALLET") != "" {
		err = json.Unmarshal([]byte(os.Getenv("WALLET")), &jwk)
		if err != nil {
			return fmt.Errorf("failed to parse WALLET environment variable: %w", err)
		}
		fmt.Println("   • Source: Environment variable")
	} else {
		walletContent, err := ioutil.ReadFile(walletPath)
		if err != nil {
			return fmt.Errorf("failed to read wallet file %s: %w", walletPath, err)
		}
		err = json.Unmarshal(walletContent, &jwk)
		if err != nil {
			return fmt.Errorf("failed to parse wallet file: %w", err)
		}
		fmt.Printf("   • Source: %s\n", walletPath)
	}
	fmt.Println("   • Status: ✅ Wallet loaded successfully")
	fmt.Println()

	// Create Arweave signer from JWK
	signer, err := signers.NewArweaveSigner(jwk)
	if err != nil {
		return fmt.Errorf("failed to create Arweave signer: %w", err)
	}

	// Convert tags to turbo format
	var tags []types.Tag
	for key, value := range publishingTags {
		tags = append(tags, types.Tag{
			Name:  key,
			Value: value,
		})
	}

	// Initialize authenticated Turbo client (using default config for production)
	turboClient := turbo.Authenticated(nil, signer)

	fmt.Println("💰 BALANCE & COST CHECK")
	fmt.Println("   ─────────────────────")
	fmt.Println("   • Checking wallet balance...")

	// Check wallet balance
	balance, err := turboClient.GetBalanceForSigner(ctx)
	if err != nil {
		// Check if it's a 404 User Not Found error - treat as 0 balance
		if strings.Contains(err.Error(), "HTTP 404") || strings.Contains(err.Error(), "User Not Found") {
			fmt.Println("   • User not found (new wallet) - treating as 0 balance")
			balance = &types.Balance{
				WinC:     "0",
				Credits:  "0",
				Currency: "winston",
			}
		} else {
			return fmt.Errorf("failed to check wallet balance: %w", err)
		}
	}

		fmt.Printf("   • Current Balance: %s\n", formatCreditsDisplay(balance.WinC))

	// Estimate upload cost
	fmt.Println("   • Estimating upload cost...")
	unauthenticatedClient := turbo.Unauthenticated(nil)
	fileSize := int64(len(wasmBinary))
	uploadCosts, err := unauthenticatedClient.GetUploadCosts(ctx, []int64{fileSize})
	if err != nil {
		return fmt.Errorf("failed to estimate upload cost: %w", err)
	}

	if len(uploadCosts) == 0 {
		return fmt.Errorf("no upload cost estimate received")
	}

	estimatedCost := uploadCosts[0].Winc
	fmt.Printf("   • Estimated Cost: %s\n", formatCreditsDisplay(estimatedCost))

	// Parse balance and cost as integers for comparison
	balanceInt, err := strconv.ParseInt(balance.WinC, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse balance: %w", err)
	}

	costInt, err := strconv.ParseInt(estimatedCost, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse cost estimate: %w", err)
	}

		// Check if balance is sufficient
	if balanceInt < costInt {
		shortfallStr := strconv.FormatInt(costInt-balanceInt, 10)
		fmt.Printf("   • ❌ Insufficient balance!\n")
		fmt.Printf("   • Required: %s\n", formatCreditsDisplay(estimatedCost))
		fmt.Printf("   • Available: %s\n", formatCreditsDisplay(balance.WinC))
		fmt.Printf("   • Shortfall: %s\n", formatCreditsDisplay(shortfallStr))
		fmt.Println()
		return fmt.Errorf("insufficient wallet balance: need %s, have %s", formatCreditsDisplay(estimatedCost), formatCreditsDisplay(balance.WinC))
	}

	remainingStr := strconv.FormatInt(balanceInt-costInt, 10)
	fmt.Printf("   • ✅ Balance sufficient for upload\n")
	fmt.Printf("   • Remaining after upload: %s\n", formatCreditsDisplay(remainingStr))
	fmt.Println()

	fmt.Println("🚀 UPLOAD PROCESS")
	fmt.Println("   ──────────────")
	fmt.Println("   • Initializing Turbo client...")
	fmt.Println("   • Preparing data item for upload...")

	// Create upload request
	uploadRequest := &types.UploadRequest{
		Data: wasmBinary,
		Tags: tags,
		Events: &types.UploadEvents{
			OnProgress: func(event types.ProgressEvent) {
				switch event.Step {
				case "signing":
					if event.ProcessedBytes == 0 {
						fmt.Println("   • Signing data item...")
					}
				case "uploading":
					if event.ProcessedBytes == 0 {
						fmt.Println("   • Uploading to Arweave...")
					}
					if event.ProcessedBytes > 0 && event.ProcessedBytes < event.TotalBytes {
						percentage := float64(event.ProcessedBytes) / float64(event.TotalBytes) * 100
						fmt.Printf("   • Upload progress: %.1f%% (%s/%s)\n",
							percentage,
							wasm.FormatMemorySize(uint32(event.ProcessedBytes)),
							wasm.FormatMemorySize(uint32(event.TotalBytes)))
					}
				}
			},
			OnSigningSuccess: func() {
				fmt.Println("   • ✅ Data signing completed successfully")
			},
			OnUploadSuccess: func(result *types.UploadResult) {
				fmt.Printf("   • 🎉 Upload completed! Transaction ID: %s\n", result.ID)
			},
			OnUploadError: func(err error) {
				fmt.Printf("   • ❌ Upload failed: %v\n", err)
			},
		},
	}

	// Upload data item
	result, err := turboClient.Upload(ctx, uploadRequest)
	if err != nil {
		return fmt.Errorf("failed to upload WASM binary: %w", err)
	}

	dataItemId := result.ID
	fmt.Println()

	fmt.Println("✅ UPLOAD SUCCESSFUL!")
	fmt.Println("   ─────────────────")
	fmt.Printf("   • Transaction ID: %s\n", dataItemId)
	fmt.Printf("   • Arweave URL: https://arweave.net/%s\n", dataItemId)
	fmt.Printf("   • Module Version: %s\n", version)
	fmt.Printf("   • Git Hash: %s\n", gitHash)
	fmt.Println()
	fmt.Println("🎭 Module successfully deployed to Arweave!")

	return nil
}

// getOrDefault returns the value if not empty, otherwise returns the default
func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// PrintUploadUsage prints usage information for the upload command
func PrintUploadUsage() {
	fmt.Println("🎭 Harlequin Upload Module - Upload Built Modules to Arweave")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("    harlequin upload-module                    # Interactive mode")
	fmt.Println("    harlequin upload-module [WASM_FILE] [OPTIONS]  # Command-line mode")
	fmt.Println()
	fmt.Println("ARGUMENTS:")
	fmt.Println("    WASM_FILE              Path to the WASM file to upload (default: dist/process.wasm)")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("    -f, --wasm-file <FILE>   Path to the WASM file to upload")
	fmt.Println("    -c, --config <FILE>      Path to build configuration file")
	fmt.Println("    -w, --wallet-file <FILE> Path to Arweave wallet JSON file (default: key.json)")
	fmt.Println("    -v, --version <VERSION>  Module version for tagging (default: dev)")
	fmt.Println("    -g, --git-hash <HASH>    Git commit hash for tagging")
	fmt.Println("    --dry-run                Show what would be uploaded without actually uploading")
	fmt.Println("    -h, --help               Show this help message")
	fmt.Println()
	fmt.Println("ENVIRONMENT VARIABLES:")
	fmt.Println("    WALLET                   Base64 encoded wallet JSON (alternative to --wallet-file)")
	fmt.Println("    WALLET_PATH              Path to wallet file (alternative to --wallet-file)")
	fmt.Println("    GITHUB_SHA               Git commit hash (used if --git-hash not provided)")
	fmt.Println()
	fmt.Println("CONFIGURATION:")
	fmt.Println("    The command looks for configuration files in this order:")
	fmt.Println("    1. .harlequin.yaml (project root)")
	fmt.Println("    2. build_configs/ao-build-config.yml")
	fmt.Println("    3. ao-build-config.yml")
	fmt.Println("    4. File specified with --config")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("    # Upload with default settings")
	fmt.Println("    harlequin upload-module")
	fmt.Println()
	fmt.Println("    # Upload specific WASM file")
	fmt.Println("    harlequin upload-module dist/my-module.wasm")
	fmt.Println()
	fmt.Println("    # Upload with custom wallet and version")
	fmt.Println("    harlequin upload-module --wallet-file ~/.arweave/wallet.json --version v1.2.3")
	fmt.Println()
	fmt.Println("    # Dry run to see what would be uploaded")
	fmt.Println("    harlequin upload-module --dry-run")
	fmt.Println()
	fmt.Println("    # Upload with custom configuration")
	fmt.Println("    harlequin upload-module --config custom-config.yml --git-hash abc123")
	fmt.Println()
	fmt.Println("TAGS:")
	fmt.Println("    The following tags are automatically added to uploads:")
	fmt.Println("    • Data-Protocol: ao")
	fmt.Println("    • Variant: ao.TN.1")
	fmt.Println("    • Type: Module")
	fmt.Println("    • Input-Encoding: JSON-1")
	fmt.Println("    • Output-Encoding: JSON-1")
	fmt.Println("    • Content-Type: application/wasm")
	fmt.Println("    • Compute-Limit: (from config)")
	fmt.Println("    • Memory-Limit: (from WASM metadata or config)")
	fmt.Println("    • Initial-Memory: (from WASM metadata)")
	fmt.Println("    • Stack-Size: (from WASM metadata)")
	fmt.Println("    • Target: (from WASM metadata or wasm32 default)")
	fmt.Println("    • Module-Format: (from config)")
	fmt.Println("    • App-Name: Harlequin-CLI")
	fmt.Println("    • AO-Module-Version: (version specified)")
	fmt.Println("    • Author: (from config or default)")
	fmt.Println("    • Git-Hash: (git commit)")
	fmt.Println("    • AOS-Git-Hash: (from config)")
	fmt.Println()
	fmt.Println("WASM METADATA:")
	fmt.Println("    The command automatically parses WASM binaries to extract:")
	fmt.Println("    • Initial memory configuration")
	fmt.Println("    • Maximum memory limits")
	fmt.Println("    • Stack size (if available)")
	fmt.Println("    • Target architecture (wasm32/wasm64)")
	fmt.Println("    • Custom sections")
	fmt.Println("    WASM metadata takes precedence over config file values.")
	fmt.Println()
}
