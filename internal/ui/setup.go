package ui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/maccavelli/mcplib/llmprovider"
	"github.com/maccavelli/mcplib/wizard"
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

func providerEnvVar(provider string) string {
	if v, ok := llmprovider.ProviderEnvVars[provider]; ok {
		return v
	}
	return ""
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

// discoverModels lists a provider's models, bounded by DiscoveryTimeout.
// The interactive path goes through wizard.ConfigureLLM; this exists for the
// non-interactive --yes path, which takes flags rather than prompts.
func discoverModels(ctx context.Context, provider, apiKey string) []string {
	dCtx, cancel := context.WithTimeout(ctx, DiscoveryTimeout)
	defer cancel()
	models, err := llmprovider.ListAvailableModels(dCtx, provider, apiKey)
	if err != nil {
		return nil
	}
	return models
}

// defaultModels returns the curated catalog mcplib ships for a provider. It is
// a thin wrapper so this file has one name for the concept; the catalog itself
// is no longer duplicated here.
func defaultModels(provider string) []string {
	return llmprovider.StaticModels(provider)
}

// recommendedFallbacks picks the models to try after the primary, preserving
// catalog order and skipping the primary itself.
func recommendedFallbacks(models []string, primary string) []string {
	var out []string
	for _, m := range models {
		if m != primary {
			out = append(out, m)
		}
		if len(out) == config.MaxFallbacks {
			break
		}
	}
	return out
}

func runSetupInteractive(ctx context.Context, conf *config.Config, opts SetupOptions, in io.Reader) error {
	reader := bufio.NewReader(in)
	fmt.Println("--- prepare-commit-msg Setup ---")

	defProvider := opts.Provider
	if defProvider == "" {
		defProvider = conf.ActiveProvider
	}
	existing := wizard.Result{Provider: defProvider}
	if pc, ok := conf.Providers[defProvider]; ok {
		existing.APIKey = pc.APIKey
		existing.Model = pc.Model
	}
	if opts.Model != "" {
		existing.Model = opts.Model
	}

	// Provider menu, key precedence, model discovery and fallback selection all
	// live in mcplib now, so a provider added there appears here with no change
	// to this file. TestSetup_OffersEveryDescriptor guards that.
	res, err := wizard.ConfigureLLM(ctx, &wizard.TextPrompter{In: in, Out: os.Stdout}, wizard.Options{
		Existing:      existing,
		AllowEnv:      !opts.NoEnv,
		LookupEnv:     osGetenv,
		Discover:      true,
		DiscoverLimit: DiscoveryTimeout,
		NeedFallbacks: true,
	})
	if err != nil {
		return err
	}

	pc, ok := conf.Providers[res.Provider]
	if !ok {
		pc = config.ProviderConfig{}
	}
	if res.APIKey != "" {
		pc.APIKey = res.APIKey
	}
	d, _ := llmprovider.DescriptorFor(res.Provider)
	if d.RequiresAPIKey && strings.TrimSpace(pc.APIKey) == "" {
		return fmt.Errorf("API key is required")
	}
	pc.Model = res.Model
	pc.FallbackModels = res.Fallbacks

	if err := promptOperational(reader, conf, opts); err != nil {
		return err
	}

	conf.ActiveProvider = res.Provider
	conf.Providers[res.Provider] = pc

	if err := conf.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if path, pathErr := config.GetConfigPath(); pathErr == nil {
		fmt.Printf("\nConfiguration saved to %s\n", path)
	} else {
		fmt.Println("\nConfiguration saved.")
	}
	fmt.Printf("Active provider: %s (%s)\n", res.Provider, pc.Model)
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
	if _, ok := llmprovider.DescriptorFor(provider); !ok {
		known := make([]string, 0, len(llmprovider.Descriptors()))
		for _, d := range llmprovider.Descriptors() {
			known = append(known, d.ID)
		}
		return fmt.Errorf("unsupported provider %q (known: %s)", provider, strings.Join(known, ", "))
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
