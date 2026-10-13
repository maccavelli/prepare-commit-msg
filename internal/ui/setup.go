package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

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
	Provider string
	APIKey   string
	// BaseURL is the provider's endpoint; empty keeps the saved one, else the
	// default (0014-MADR Amendment 2, B2).
	BaseURL           string
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
	// ClientOptions name the application in configure's SDK requests; none
	// leaves the SDK naming itself (0014-MADR D5).
	ClientOptions []llmprovider.Option
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

func providerEnvVar(provider string) string {
	if v, ok := llmprovider.ProviderEnvVars()[llmprovider.ProviderID(provider)]; ok {
		return v
	}
	return ""
}

// promptInt asks for a positive integer on the wizard's prompter, so the
// wizard and these prompts share one buffered input (0014-MADR D3). A blank
// answer, the end of input, or anything but a positive integer keeps current.
func promptInt(p wizard.Prompter, label string, current int) (int, error) {
	line, err := p.Input(label, strconv.Itoa(current))
	if errors.Is(err, io.EOF) {
		return current, nil
	}
	if err != nil {
		return current, fmt.Errorf("read %s: %w", label, err)
	}
	line = strings.TrimSpace(line)
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

func promptOperational(p wizard.Prompter, conf *config.Config, opts SetupOptions) error {
	fmt.Println("\n--- Operational settings (Enter keeps current) ---")
	var err error
	timeout := conf.TimeoutSeconds
	if opts.TimeoutSeconds > 0 {
		timeout = opts.TimeoutSeconds
	}
	conf.TimeoutSeconds, err = promptInt(p, "Timeout seconds", timeout)
	if err != nil {
		return err
	}

	maxDiff := conf.MaxDiffBytes
	if opts.MaxDiffBytes > 0 {
		maxDiff = opts.MaxDiffBytes
	}
	conf.MaxDiffBytes, err = promptInt(p, "Max diff bytes", maxDiff)
	if err != nil {
		return err
	}

	retry := conf.RetryCount
	if opts.RetryCount > 0 {
		retry = opts.RetryCount
	}
	conf.RetryCount, err = promptInt(p, "Retry count", retry)
	if err != nil {
		return err
	}

	delay := conf.RetryDelaySeconds
	if opts.RetryDelaySeconds > 0 {
		delay = opts.RetryDelaySeconds
	}
	conf.RetryDelaySeconds, err = promptInt(p, "Retry delay seconds", delay)
	if err != nil {
		return err
	}
	return nil
}

// discoverModels lists a provider's short list of models; catalog.List bounds
// the listing itself. The interactive path goes through wizard.ConfigureLLM;
// this exists for the non-interactive --yes path, which takes flags rather
// than prompts. A key the service refuses is an error, as in the wizard; any
// other failed listing gives the static catalog (0014-MADR D4). A non-empty
// baseURL is the endpoint listed (0014-MADR Amendment 2, B3); extra carries
// the application's name (0014-MADR D5).
func discoverModels(ctx context.Context, provider, apiKey, baseURL string, extra ...llmprovider.Option) ([]string, error) {
	opts := append(catalog.OptionsFromEnv(), catalog.WithLimit(catalog.MaxListed))
	opts = append(opts, extra...)
	if baseURL != "" {
		opts = append(opts, llmprovider.WithBaseURL(baseURL))
	}
	if listingClient != nil {
		opts = append(opts, llmprovider.WithHTTPClient(listingClient))
	}
	cat, err := catalog.List(ctx, llmprovider.ProviderID(provider), llmprovider.NewStaticToken(apiKey), opts...)
	if err != nil {
		if refusedCredential(err) {
			return nil, err
		}
		return nil, nil
	}
	if refusedCredential(cat.Err) {
		return nil, cat.Err
	}
	// A live listing that curates to nothing keeps Recommended empty, so the
	// wizard can open its search; this path has no search, so it takes the
	// curated catalog, as the SDK itself did before v1.2.0 (0009-MADR D3).
	if len(cat.Recommended) == 0 {
		return defaultModels(provider), nil
	}
	return cat.Recommended, nil
}

// refusedKeyError is configure --yes's error for a key the service refused;
// nothing is saved.
func refusedKeyError(provider string, err error) error {
	return fmt.Errorf("%s refused the API key: %w", provider, err)
}

// refusedCredential reports whether err is the service refusing the
// credential: an authentication failure that is not a refusal to permit
// something. It is the wizard's own rule, which the SDK does not export
// (go-llmprovider-sdk wizard/configure.go, 0028-MADR D-H2).
func refusedCredential(err error) bool {
	return errors.Is(err, llmprovider.ErrAuthFailure) && !errors.Is(err, llmprovider.ErrNotPermitted)
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
	// One prompter reads the whole run: a second reader on in would find the
	// prompter's buffer had already drained it (0014-MADR D3).
	prompter := &wizard.TextPrompter{In: in, Out: os.Stdout}
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
		existing.BaseURL = pc.BaseURL
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
	if opts.BaseURL != "" {
		existing.BaseURL = opts.BaseURL
	}

	// Provider menu, key precedence, model discovery and fallback selection all
	// live in go-llmprovider-sdk now, so a provider added there appears here with no change
	// to this file. TestSetup_OffersEveryDescriptor guards that.
	res, err := wizard.ConfigureLLM(ctx, prompter, wizard.Options{
		Existing:        existing,
		AllowEnv:        !opts.NoEnv,
		LookupEnv:       config.LookupEnv(osGetenv),
		Discover:        true,
		ModelLimit:      catalog.MaxListed,
		HTTPClient:      listingClient,
		Registry:        registry,
		ProviderOptions: append(catalog.OptionsFromEnv(), opts.ClientOptions...),
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
	pc.BaseURL = savedBaseURL(res.Provider, res.BaseURL)
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

	if err := promptOperational(prompter, conf, opts); err != nil {
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

// savedBaseURL is the endpoint configure chose when it is not the provider's
// default, which needs no saving (0014-MADR D6).
func savedBaseURL(id llmprovider.ProviderID, baseURL string) string {
	d, _ := registry.Descriptor(id)
	if baseURL == d.DefaultBaseURL {
		return ""
	}
	return baseURL
}

// checkBaseURL checks configure --yes's --base-url by the wizard's rule for a
// typed endpoint, which the SDK does not export (go-llmprovider-sdk
// wizard/base_url.go, 0021-MADR Z7): http or https with a host, no user name
// or password, no query and no fragment. --yes cannot ask the wizard's
// confirmation, so plain http to a host that is not loopback is refused
// where a key would travel over it (0014-MADR Amendment 2, B2). No error
// quotes a URL that could carry a password.
func checkBaseURL(d llmprovider.Descriptor, raw string) (string, error) {
	const plain, secure = "http", "https"
	if !d.SupportsBaseURL {
		return "", fmt.Errorf("%s takes no --base-url", d.ID)
	}
	u, err := url.Parse(raw)
	switch {
	case err != nil:
		return "", errors.New("--base-url is not a URL")
	case u.User != nil:
		return "", errors.New("--base-url carries no user name or password")
	case u.Scheme != plain && u.Scheme != secure || u.Host == "":
		return "", fmt.Errorf("--base-url %s needs http:// or https:// and a host", raw)
	case u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(raw, "#"):
		return "", fmt.Errorf("--base-url %s has a query or a fragment", raw)
	case u.Scheme == plain && d.RequiresAPIKey && !loopbackHost(u.Hostname()):
		return "", fmt.Errorf("--base-url %s is plain http to a remote host, so the key would travel unencrypted; use https", raw)
	}
	return raw, nil
}

// loopbackHost reports localhost or a loopback IP address, as the wizard's
// own check does.
func loopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func runSetupNonInteractive(ctx context.Context, conf *config.Config, opts SetupOptions) error {
	provider := opts.Provider
	if provider == "" {
		provider = conf.ActiveProvider
	}
	if provider == "" {
		provider = providerGemini
	}
	d, ok := registry.Descriptor(llmprovider.ProviderID(provider))
	if !ok {
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
	// Only a provider whose descriptor needs a key is refused without one
	// (0014-MADR Amendment 2, B1).
	if apiKey == "" && d.RequiresAPIKey {
		envName := providerEnvVar(provider)
		return fmt.Errorf("API key required: pass --api-key or set %s", envName)
	}
	baseURL := pc.BaseURL
	if raw := strings.TrimSpace(opts.BaseURL); raw != "" {
		checked, err := checkBaseURL(d, raw)
		if err != nil {
			return err
		}
		baseURL = checked
	}
	pc.APIKey = apiKey
	pc.BaseURL = savedBaseURL(d.ID, baseURL)
	pc.AuthKind = ""
	pc.VendorAuthPath = ""
	pc.Organization = ""

	model := strings.TrimSpace(opts.Model)
	if model == "" {
		models, err := discoverModels(ctx, provider, apiKey, baseURL, opts.ClientOptions...)
		if err != nil {
			return refusedKeyError(provider, err)
		}
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
		models, err := discoverModels(ctx, provider, apiKey, baseURL, opts.ClientOptions...)
		if err != nil {
			return refusedKeyError(provider, err)
		}
		pc.FallbackModels = config.ClampFallbacks(recommendedFallbacks(models, model))
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
