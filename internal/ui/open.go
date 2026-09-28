package ui

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var openBrowser = openBrowserDefault

func openBrowserDefault(rawURL string) error {
	if err := validateBrowserURL(rawURL); err != nil {
		return err
	}
	fmt.Printf("Open this URL in your browser: %s\n", rawURL)
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		// #nosec G204 -- rawURL is validated and passed directly without a shell.
		command = exec.Command("open", rawURL)
	case "linux":
		// #nosec G204 -- rawURL is validated and passed directly without a shell.
		command = exec.Command("xdg-open", rawURL)
	case "windows":
		// #nosec G204 -- rawURL is validated and passed directly without a shell.
		command = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", rawURL)
	default:
		return fmt.Errorf("opening a browser is not supported on %s", runtime.GOOS)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	// Reap the launcher without blocking the sign-in flow. The URL is already
	// printed, so a launcher that fails only needs saying so.
	go func() {
		if err := command.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "could not open a browser (%v); open the URL above instead\n", err)
		}
	}()
	return nil
}

func validateBrowserURL(rawURL string) error {
	if strings.ContainsAny(rawURL, "\x00\r\n") {
		return fmt.Errorf("invalid browser URL: control character")
	}
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return fmt.Errorf("invalid browser URL: absolute HTTP(S) URL required")
	}
	return nil
}
