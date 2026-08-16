package ui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/maccavelli/mcplib/llmprovider"
	"github.com/maccavelli/prepare-commit-msg/internal/config"
)

const (
	providerGemini = llmprovider.ProviderGemini
	providerOpenAI = llmprovider.ProviderOpenAI
	providerClaude = llmprovider.ProviderClaude

	// DiscoveryTimeout bounds model list API calls during configure.
	DiscoveryTimeout = 45 * time.Second
)

var osGetenv = os.Getenv
var readPassword = term.ReadPassword // Mockable for tests

// SetupOptions holds non-interactive / flag-driven configure settings.
// Zero values mean "unset" and interactive mode will prompt (unless Yes is set).
type SetupOptions struct {
	Provider          string
	APIKey            string
	Model             string
	Fallbacks         []string
	TimeoutSeconds    int
	MaxDiffBytes      int
	RetryCount        int
	RetryDelaySeconds int
	// NoEnv disables reading provider API keys from the environment.
	NoEnv bool
	// Yes runs non-interactively; missing required fields become errors or defaults.
	Yes bool
}

// RunSetup runs the interactive configuration wizard.
func RunSetup(ctx context.Context, conf *config.Config) error {
	return RunSetupWithOptions(ctx, conf, SetupOptions{}, os.Stdin)
}

// RunSetupWithOptions runs configure using flags and/or an interactive wizard.
func RunSetupWithOptions(ctx context.Context, conf *config.Config, opts SetupOptions, in io.Reader) error {
	if opts.Yes {
		return runSetupNonInteractive(ctx, conf, opts)
	}
	// If any configure flag was provided without --yes, seed the interactive flow.
	return runSetupInteractive(ctx, conf, opts, in)
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptProvider(reader *bufio.Reader, defaultProvider string) (string, error) {
	if defaultProvider == "" {
		defaultProvider = providerGemini
	}
	fmt.Println("Choose LLM Provider:")
	fmt.Println("  1) gemini")
	fmt.Println("  2) openai")
	fmt.Println("  3) claude")
	defIdx := "1"
	switch defaultProvider {
	case providerOpenAI:
		defIdx = "2"
	case providerClaude:
		defIdx = "3"
	}
	fmt.Printf("Select [1/2/3] (default %s): ", defIdx)
	choice, err := readLine(reader)
	if err != nil {
		return "", fmt.Errorf("read provider choice: %w", err)
	}
	if choice == "" {
		choice = defIdx
	}
	switch choice {
	case "2", providerOpenAI:
		return providerOpenAI, nil
	case "3", providerClaude:
		return providerClaude, nil
	default:
		return providerGemini, nil
	}
}

func providerEnvVar(provider string) string {
	if v, ok := llmprovider.ProviderEnvVars[provider]; ok {
		return v
	}
	return ""
}

func resolveAPIKey(reader *bufio.Reader, in io.Reader, provider string, pc config.ProviderConfig, opts SetupOptions) (string, error) {
	if opts.APIKey != "" {
		return strings.TrimSpace(opts.APIKey), nil
	}

	if !opts.NoEnv {
		if envVar := providerEnvVar(provider); envVar != "" {
			if envVal := osGetenv(envVar); envVal != "" {
				fmt.Printf("\n%s detected in environment!\n", envVar)
				fmt.Print("Use this key? [Y/n]: ")
				useEnv, err := readLine(reader)
				if err != nil {
					return "", fmt.Errorf("read env key choice: %w", err)
				}
				useEnv = strings.ToLower(useEnv)
				if useEnv == "" || useEnv == "y" || useEnv == "yes" {
					fmt.Println("Using environment key.")
					return envVal, nil
				}
			}
		}
	}

	if pc.APIKey != "" {
		fmt.Printf("\nExisting key found in config.\n")
		fmt.Print("Keep existing key? [Y/n]: ")
		keep, err := readLine(reader)
		if err != nil {
			return "", fmt.Errorf("read keep key choice: %w", err)
		}
		keep = strings.ToLower(keep)
		if keep == "" || keep == "y" || keep == "yes" {
			fmt.Println("Keeping existing key.")
			return pc.APIKey, nil
		}
	}

	fmt.Printf("\nEnter %s API Key: ", provider)
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		bytes, err := readPassword(int(f.Fd()))
		fmt.Println()
		if err == nil {
			key := strings.TrimSpace(string(bytes))
			if key != "" {
				return key, nil
			}
			// Empty password entry — fall through to visible prompt.
			fmt.Println("Empty key; try again (input will be visible):")
		} else {
			// Git Bash / mintty / some terminals fail hidden read — fall back.
			fmt.Printf("(hidden input unavailable: %v — typing will be visible)\n", err)
			fmt.Printf("Enter %s API Key: ", provider)
		}
	}
	input, err := readLine(reader)
	if err != nil {
		return "", fmt.Errorf("read api key: %w", err)
	}
	return input, nil
}

var modelAnnotations = map[string]string{
	// Gemini fast models
	"gemini-3.7-flash":      "Gemini 3.7 Flash      [★ Recommended: Frontier coding intelligence]",
	"gemini-3.6-flash":      "Gemini 3.6 Flash      [High efficiency & reduced token overhead]",
	"gemini-3.5-flash":      "Gemini 3.5 Flash      [High-speed production workhorse]",
	"gemini-3.5-flash-lite": "Gemini 3.5 Flash-Lite [Ultra-low latency & high throughput]",
	"gemini-2.5-flash":      "Gemini 2.5 Flash      [Proven balanced fast baseline]",
	"gemini-2.5-flash-lite": "Gemini 2.5 Flash-Lite [Lightweight fast baseline]",
	// OpenAI fast models
	"gpt-4.1-mini": "GPT-4.1 Mini          [★ Recommended: Fast & cost-effective]",
	"gpt-4.1-nano": "GPT-4.1 Nano          [Ultra-low latency lightweight]",
	"gpt-4o-mini":  "GPT-4o Mini           [Stable fast chat model]",
	"gpt-4.1":      "GPT-4.1               [High-capability fast tier]",
	"gpt-4o":       "GPT-4o                [Flagship multimodal]",
	"o4-mini":      "o4-mini               [Fast reasoning]",
	// Claude fast models
	"claude-haiku-4-5":        "Claude Haiku 4.5      [★ Recommended: High speed & low latency]",
	"claude-sonnet-5":         "Claude Sonnet 5       [Balanced speed & high precision]",
	"claude-sonnet-4-6":       "Claude Sonnet 4.6     [Stable high precision]",
	"claude-opus-4-8":         "Claude Opus 4.8       [Maximum capability]",
	"claude-3-5-haiku-latest": "Claude 3.5 Haiku      [Fast lightweight]",
}

func annotateModel(modelID string) string {
	if ann, ok := modelAnnotations[modelID]; ok {
		return ann
	}
	return modelID
}

func defaultModels(provider string) []string {
	return llmprovider.StaticModels(provider)
}

// discoverModels returns a short curated list of production text models.
// Prefer free ListAvailableModels (catalog ∩ API) over DiscoverModels health
// probes so configure stays fast and never dumps huge Gemini specialty lists.
func discoverModels(ctx context.Context, provider, apiKey string) []string {
	dctx, cancel := context.WithTimeout(ctx, DiscoveryTimeout)
	defer cancel()

	fmt.Println("\nLoading recommended models for this provider...")
	listed, err := llmprovider.ListAvailableModels(dctx, provider, apiKey)
	if err == nil && len(listed) > 0 {
		return listed
	}

	// Fall back to static catalog only (no token-burning health probes).
	if fallback := defaultModels(provider); len(fallback) > 0 {
		return fallback
	}
	return nil
}

func promptModel(reader *bufio.Reader, models []string, defaultModel string) (string, error) {
	fmt.Println("\nRecommended fast models for commit messages (curated top options):")
	defaultIdx := 1
	for i, m := range models {
		fmt.Printf("  %d) %s\n", i+1, annotateModel(m))
		if defaultModel != "" && m == defaultModel {
			defaultIdx = i + 1
		}
	}
	fmt.Printf("  %d) Other (enter manually)\n", len(models)+1)
	fmt.Printf("Select Model [%d]: ", defaultIdx)

	modChoiceStr, err := readLine(reader)
	if err != nil {
		return "", fmt.Errorf("read model choice: %w", err)
	}
	if modChoiceStr == "" {
		if defaultModel != "" {
			return defaultModel, nil
		}
		return models[0], nil
	}

	var idx int
	if _, scanErr := fmt.Sscanf(modChoiceStr, "%d", &idx); scanErr != nil {
		// Non-numeric input is treated as a literal model name.
		return strings.TrimSpace(modChoiceStr), nil //nolint:nilerr // intentional: free-typed model name
	}
	if idx > 0 && idx <= len(models) {
		return models[idx-1], nil
	}
	if idx == len(models)+1 {
		fmt.Print("Enter model name: ")
		m, err := readLine(reader)
		if err != nil {
			return "", fmt.Errorf("read model name: %w", err)
		}
		return m, nil
	}
	return models[0], nil
}

// recommendedFallbacks returns up to MaxFallbacks models from list excluding primary.
func recommendedFallbacks(models []string, primary string) []string {
	var out []string
	for _, m := range models {
		if m == primary {
			continue
		}
		out = append(out, m)
		if len(out) >= config.MaxFallbacks {
			break
		}
	}
	return out
}

func promptFallbacks(reader *bufio.Reader, models []string, primary string, preset []string) ([]string, error) {
	candidates := make([]string, 0, len(models))
	for _, m := range models {
		if m != primary {
			candidates = append(candidates, m)
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	rec := recommendedFallbacks(models, primary)
	if len(preset) > 0 {
		rec = config.ClampFallbacks(preset)
	}

	fmt.Printf("\nFallback models (up to %d). Recommended: %s\n", config.MaxFallbacks, strings.Join(rec, ", "))
	for i, m := range candidates {
		mark := " "
		if slices.Contains(rec, m) {
			mark = "*"
		}
		fmt.Printf("  %d) %s %s\n", i+1, mark, m)
	}
	fmt.Print("Select fallbacks as comma-separated numbers (Enter = recommended, 0 = none): ")
	line, err := readLine(reader)
	if err != nil {
		return nil, fmt.Errorf("read fallbacks: %w", err)
	}
	if line == "" {
		return rec, nil
	}
	if line == "0" {
		return nil, nil
	}

	var selected []string
	for part := range strings.SplitSeq(line, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > len(candidates) {
			continue
		}
		selected = append(selected, candidates[n-1])
	}
	return config.ClampFallbacks(selected), nil
}

func promptInt(reader *bufio.Reader, label string, current int) (int, error) {
	fmt.Printf("%s [%d]: ", label, current)
	line, err := readLine(reader)
	if err != nil {
		return current, fmt.Errorf("read %s: %w", label, err)
	}
	if line == "" {
		return current, nil
	}
	n, parseErr := strconv.Atoi(line)
	if parseErr == nil && n > 0 {
		return n, nil
	}
	fmt.Printf("  invalid value %q — keeping %d\n", line, current)
	return current, nil
}

func promptOperational(reader *bufio.Reader, conf *config.Config, opts SetupOptions) error {
	fmt.Println("\n--- Operational settings (Enter keeps current) ---")
	var err error
	timeout := conf.TimeoutSeconds
	if opts.TimeoutSeconds > 0 {
		timeout = opts.TimeoutSeconds
	}
	conf.TimeoutSeconds, err = promptInt(reader, "Timeout seconds", timeout)
	if err != nil {
		return err
	}

	maxDiff := conf.MaxDiffBytes
	if opts.MaxDiffBytes > 0 {
		maxDiff = opts.MaxDiffBytes
	}
	conf.MaxDiffBytes, err = promptInt(reader, "Max diff bytes", maxDiff)
	if err != nil {
		return err
	}

	retry := conf.RetryCount
	if opts.RetryCount > 0 {
		retry = opts.RetryCount
	}
	conf.RetryCount, err = promptInt(reader, "Retry count", retry)
	if err != nil {
		return err
	}

	delay := conf.RetryDelaySeconds
	if opts.RetryDelaySeconds > 0 {
		delay = opts.RetryDelaySeconds
	}
	conf.RetryDelaySeconds, err = promptInt(reader, "Retry delay seconds", delay)
	if err != nil {
		return err
	}
	return nil
}

func runSetupInteractive(ctx context.Context, conf *config.Config, opts SetupOptions, in io.Reader) error {
	reader := bufio.NewReader(in)
	fmt.Println("--- prepare-commit-msg Setup ---")

	defProvider := opts.Provider
	if defProvider == "" {
		defProvider = conf.ActiveProvider
	}
	provider, err := promptProvider(reader, defProvider)
	if err != nil {
		return err
	}

	pc, ok := conf.Providers[provider]
	if !ok {
		pc = config.ProviderConfig{}
	}

	apiKey, err := resolveAPIKey(reader, in, provider, pc, opts)
	if err != nil {
		return err
	}
	if apiKey != "" {
		pc.APIKey = apiKey
	}
	if strings.TrimSpace(pc.APIKey) == "" {
		return fmt.Errorf("API key is required")
	}

	models := discoverModels(ctx, provider, pc.APIKey)
	if len(models) == 0 {
		models = defaultModels(provider)
	}

	defModel := opts.Model
	if defModel == "" {
		defModel = pc.Model
	}
	model, err := promptModel(reader, models, defModel)
	if err != nil {
		return err
	}
	if strings.TrimSpace(model) == "" {
		return fmt.Errorf("model is required")
	}
	pc.Model = model

	fallbacks, err := promptFallbacks(reader, models, model, opts.Fallbacks)
	if err != nil {
		return err
	}
	pc.FallbackModels = fallbacks

	if err := promptOperational(reader, conf, opts); err != nil {
		return err
	}

	conf.ActiveProvider = provider
	conf.Providers[provider] = pc

	if err := conf.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if path, pathErr := config.GetConfigPath(); pathErr == nil {
		fmt.Printf("\nConfiguration saved to %s\n", path)
	} else {
		fmt.Println("\nConfiguration saved.")
	}
	fmt.Printf("Active provider: %s (%s)\n", provider, pc.Model)
	if len(pc.FallbackModels) > 0 {
		fmt.Printf("Fallbacks: %s\n", strings.Join(pc.FallbackModels, ", "))
	}
	return nil
}

func runSetupNonInteractive(ctx context.Context, conf *config.Config, opts SetupOptions) error {
	provider := opts.Provider
	if provider == "" {
		provider = conf.ActiveProvider
	}
	if provider == "" {
		provider = providerGemini
	}
	switch provider {
	case providerGemini, providerOpenAI, providerClaude:
	default:
		return fmt.Errorf("unsupported provider %q (use gemini, openai, or claude)", provider)
	}

	pc, ok := conf.Providers[provider]
	if !ok {
		pc = config.ProviderConfig{}
	}

	apiKey := strings.TrimSpace(opts.APIKey)
	if apiKey == "" {
		apiKey = config.ResolveAPIKey(pc, provider, !opts.NoEnv, osGetenv)
	}
	if apiKey == "" {
		envName := providerEnvVar(provider)
		return fmt.Errorf("API key required: pass --api-key or set %s", envName)
	}
	pc.APIKey = apiKey

	model := strings.TrimSpace(opts.Model)
	if model == "" {
		models := discoverModels(ctx, provider, apiKey)
		if len(models) == 0 {
			models = defaultModels(provider)
		}
		model = models[0]
		if len(opts.Fallbacks) == 0 {
			opts.Fallbacks = recommendedFallbacks(models, model)
		}
	}
	pc.Model = model

	if len(opts.Fallbacks) > 0 {
		pc.FallbackModels = config.ClampFallbacks(opts.Fallbacks)
	} else {
		// Keep existing fallbacks if any; otherwise recommend from static list.
		if len(pc.FallbackModels) == 0 {
			pc.FallbackModels = recommendedFallbacks(defaultModels(provider), model)
		}
		pc.FallbackModels = config.ClampFallbacks(pc.FallbackModels)
	}

	if opts.TimeoutSeconds > 0 {
		conf.TimeoutSeconds = opts.TimeoutSeconds
	}
	if opts.MaxDiffBytes > 0 {
		conf.MaxDiffBytes = opts.MaxDiffBytes
	}
	if opts.RetryCount > 0 {
		conf.RetryCount = opts.RetryCount
	}
	if opts.RetryDelaySeconds > 0 {
		conf.RetryDelaySeconds = opts.RetryDelaySeconds
	}

	conf.ActiveProvider = provider
	conf.Providers[provider] = pc

	if err := conf.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if path, pathErr := config.GetConfigPath(); pathErr == nil {
		fmt.Printf("Configuration saved to %s\n", path)
	} else {
		fmt.Println("Configuration saved.")
	}
	fmt.Printf("Active provider: %s (%s)\n", provider, pc.Model)
	return nil
}
