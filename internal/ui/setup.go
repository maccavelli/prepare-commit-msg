package ui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/catalog"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/providers"
	"github.com/maccavelli/go-llmprovider-sdk/wizard"
	"github.com/maccavelli/prepare-commit-msg/internal/config"
)

const (
	providerGemini = string(llmprovider.ProviderGemini)
	providerOpenAI = string(llmprovider.ProviderOpenAI)
	providerClaude = string(llmprovider.ProviderClaude)
	providerGrok   = string(llmprovider.ProviderGrok)

	// DiscoveryTimeout bounds model list API calls during configure.
	DiscoveryTimeout = 45 * time.Second
)

var osGetenv = os.Getenv

// registry is the provider set configure offers and accepts.
var registry = providers.Default()

// listingClient carries configure's live model listing. Nil uses the SDK's
// default client; tests set one that never reaches the network.
var listingClient *http.Client

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
	if v, ok := llmprovider.ProviderEnvVars()[llmprovider.ProviderID(provider)]; ok {
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
	opts := catalog.OptionsFromEnv()
	if listingClient != nil {
		opts = append(opts, llmprovider.WithHTTPClient(listingClient))
	}
	cat, err := catalog.List(dCtx, llmprovider.ProviderID(provider), llmprovider.NewStaticToken(apiKey), opts...)
	if err != nil {
		return nil
	}
	return cat.Recommended
}

// defaultModels returns the curated catalog the SDK ships for a provider. It is
// a thin wrapper so this file has one name for the concept; the catalog itself
// is no longer duplicated here.
func defaultModels(provider string) []string {
	return catalog.Static(llmprovider.ProviderID(provider))
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
	store, err := config.NewOAuthStore()
	if err != nil {
		return fmt.Errorf("create OAuth store: %w", err)
	}

	defProvider := opts.Provider
	if defProvider == "" {
		defProvider = conf.ActiveProvider
	}
	existing := wizard.Result{Provider: llmprovider.ProviderID(defProvider)}
	if pc, ok := conf.Providers[defProvider]; ok {
		existing.Model = pc.Model
		existing.Organization = pc.Organization
		if config.IsOAuth(pc) {
			session, loadErr := store.Load(ctx, llmprovider.ProviderID(defProvider))
			if loadErr != nil {
				return fmt.Errorf("load existing OAuth session: %w", loadErr)
			}
			if session != nil {
				existing.Kind = wizard.CredOAuth
				existing.TokenExpiry = session.Expiry
				existing.Issuer = session.Issuer
				existing.ClientID = session.ClientID
				existing.AccountID = session.AccountID
			}
		} else {
			existing.APIKey = pc.APIKey
			if pc.APIKey != "" {
				existing.Kind = wizard.CredAPIKey
			}
		}
	}
	if opts.Model != "" {
		existing.Model = opts.Model
	}

	// Provider menu, key precedence, model discovery and fallback selection all
	// live in go-llmprovider-sdk now, so a provider added there appears here with no change
	// to this file. TestSetup_OffersEveryDescriptor guards that.
	res, err := wizard.ConfigureLLM(ctx, &wizard.TextPrompter{In: in, Out: os.Stdout}, wizard.Options{
		Existing:        existing,
		AllowEnv:        !opts.NoEnv,
		LookupEnv:       config.LookupEnv(osGetenv),
		Discover:        true,
		DiscoverLimit:   DiscoveryTimeout,
		HTTPClient:      listingClient,
		Registry:        registry,
		ProviderOptions: catalog.OptionsFromEnv(),
		NeedFallbacks:   true,
		TokenStore:      store,
		OpenURL:         openBrowser,
	})
	if err != nil {
		return err
	}

	pc, ok := conf.Providers[string(res.Provider)]
	if !ok {
		pc = config.ProviderConfig{}
	}
	pc.Model = res.Model
	pc.FallbackModels = res.Fallbacks
	pc.Organization = res.Organization
	pc.VendorAuthPath = ""
	switch res.Kind {
	case wizard.CredOAuth:
		pc.AuthKind = string(wizard.CredOAuth)
		pc.APIKey = ""
		if err := config.ValidateOAuth(ctx, string(res.Provider), pc, store); err != nil {
			return err
		}
	case wizard.CredVendorCLI:
		pc.AuthKind = config.AuthKindVendorCLI
		pc.APIKey = ""
		pc.VendorAuthPath = res.VendorAuthPath
		if err := config.ValidateVendorCLI(string(res.Provider), pc); err != nil {
			return err
		}
		// A session an older release copied from the CLI shares the CLI's
		// refresh token; drop it so it is never refreshed again.
		if err := store.Delete(ctx, res.Provider); err != nil {
			return fmt.Errorf("delete stale OAuth session: %w", err)
		}
	default:
		pc.AuthKind = ""
		pc.APIKey = res.APIKey
		d, _ := registry.Descriptor(res.Provider)
		if d.RequiresAPIKey && strings.TrimSpace(pc.APIKey) == "" {
			return fmt.Errorf("API key is required")
		}
		if err := store.Delete(ctx, res.Provider); err != nil {
			return fmt.Errorf("delete stale OAuth session: %w", err)
		}
	}

	if err := promptOperational(reader, conf, opts); err != nil {
		return err
	}

	conf.ActiveProvider = string(res.Provider)
	conf.Providers[string(res.Provider)] = pc

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
	if _, ok := registry.Descriptor(llmprovider.ProviderID(provider)); !ok {
		descriptors := registry.Descriptors()
		known := make([]string, 0, len(descriptors))
		for _, d := range descriptors {
			known = append(known, string(d.ID))
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
	pc.AuthKind = ""
	pc.VendorAuthPath = ""
	pc.Organization = ""

	model := strings.TrimSpace(opts.Model)
	if model == "" {
		models := discoverModels(ctx, provider, apiKey)
		if len(models) == 0 {
			return fmt.Errorf("no live models listed for %s; pass --model", provider)
		}
		model = models[0]
		if len(opts.Fallbacks) == 0 {
			opts.Fallbacks = recommendedFallbacks(models, model)
		}
	}
	pc.Model = model

	if len(opts.Fallbacks) > 0 {
		pc.FallbackModels = config.ClampFallbacks(opts.Fallbacks)
	} else if len(pc.FallbackModels) == 0 {
		pc.FallbackModels = config.ClampFallbacks(recommendedFallbacks(discoverModels(ctx, provider, apiKey), model))
	} else {
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
