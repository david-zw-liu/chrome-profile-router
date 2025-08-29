package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Rule struct {
	Pattern          string `json:"pattern"`
	ProfileDirectory string `json:"profile_directory"`
}

type Config struct {
	ChromeAppPath           string `json:"chrome_app_path"`
	DefaultProfileDirectory string `json:"default_profile_directory"`
	Rules                   []Rule `json:"rules"`
}

type compiledRule struct {
	re               *regexp.Regexp
	profileDirectory string
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "chrome-profile-router", "config.json")
}

func loadConfig(path string) (Config, []compiledRule, error) {
	var cfg Config

	f, err := os.Open(path)
	if err != nil {
		return cfg, nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return cfg, nil, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, nil, fmt.Errorf("parse config JSON: %w", err)
	}

	if cfg.ChromeAppPath == "" {
		cfg.ChromeAppPath = "/Applications/Google Chrome.app"
	}
	if cfg.DefaultProfileDirectory == "" {
		cfg.DefaultProfileDirectory = "Default"
	}

	var cr []compiledRule
	for i, r := range cfg.Rules {
		if r.Pattern == "" || r.ProfileDirectory == "" {
			return cfg, nil, fmt.Errorf("rule %d invalid: pattern and profile_directory are required", i)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return cfg, nil, fmt.Errorf("rule %d: compile regexp: %w", i, err)
		}
		cr = append(cr, compiledRule{re: re, profileDirectory: r.ProfileDirectory})
	}

	return cfg, cr, nil
}

func chooseProfile(urlStr string, compiled []compiledRule, fallback string) string {
	for _, r := range compiled {
		if r.re.MatchString(urlStr) {
			return r.profileDirectory
		}
	}
	return fallback
}

// macOS-friendly launcher for Chrome with profile.
// Uses: open -na "Google Chrome" --args --profile-directory="X" "URL"
func openInChrome(chromeAppPath, profileDir, urlStr string) error {
	// Sanity: ensure it's a URL we can hand off (http/https/file/custom schemes may arrive).
	// We’ll pass anything we got; but prefer http/https/mailto like a normal browser.
	// macOS will pass the exact URL given to the default browser.
	u, err := url.Parse(urlStr)
	if err != nil {
		// still try; Chrome might handle it
	} else if u.Scheme == "" && !strings.HasPrefix(urlStr, "http") {
		// If it’s bare text, try to force http
		urlStr = "http://" + urlStr
	}

	args := []string{
		"-na", chromeAppPath,
		"--args",
		fmt.Sprintf("--profile-directory=%s", profileDir),
		urlStr,
	}

	cmd := exec.Command("open", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func isLikelyURL(s string) bool {
	// When set as default browser, macOS usually passes one argument that’s the URL.
	// But users might run it manually or drag-drop etc. Accept common patterns.
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "file://") || strings.HasPrefix(s, "mailto:") {
		return true
	}
	// Simple domain-ish heuristic
	return strings.Contains(s, ".") || strings.Contains(s, ":")
}

func main() {
	cfgPathFlag := flag.String("config", "", "Path to config JSON (default: ~/.config/chrome-profile-router/config.json)")
	dryRun := flag.Bool("dry-run", false, "Don’t launch Chrome; just print routing decisions")
	version := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *version {
		fmt.Println("chrome-profile-router 1.0.0")
		return
	}

	cfgPath := *cfgPathFlag
	if cfgPath == "" {
		cfgPath = defaultConfigPath()
	}

	cfg, compiledRules, err := loadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(2)
	}

	// Collect URLs from args. macOS typically passes one arg (the URL),
	// but we’ll route multiple if provided.
	args := flag.Args()
	if len(args) == 0 {
		// Some launchers might pass URLs via STDIN (rare). Try that as a fallback.
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			stdinBytes, _ := io.ReadAll(os.Stdin)
			txt := strings.TrimSpace(string(stdinBytes))
			if txt != "" {
				args = append(args, txt)
			}
		}
	}

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "No URL provided.")
		os.Exit(1)
	}

	// Process each argument that looks like a URL
	var firstErr error
	for _, raw := range args {
		if !isLikelyURL(raw) {
			continue
		}
		urlStr := strings.TrimSpace(raw)
		profile := chooseProfile(urlStr, compiledRules, cfg.DefaultProfileDirectory)
		fmt.Fprintf(os.Stderr, "Routing: %s  ->  profile-directory=%q\n", urlStr, profile)

		if *dryRun {
			continue
		}

		if err := openInChrome(cfg.ChromeAppPath, profile, urlStr); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open URL in Chrome: %v\n", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	if firstErr != nil {
		// exit non-zero if any failed
		fmt.Fprintln(os.Stderr, "One or more URLs failed to open.")
		os.Exit(3)
	}
}
