package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include "handler.h"
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

var urlListener chan string = make(chan string)
var config Config
var compiledRules []compiledRule

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
	// We'll pass anything we got; but prefer http/https/mailto like a normal browser.
	// macOS will pass the exact URL given to the default browser.
	u, err := url.Parse(urlStr)
	if err != nil {
		// still try; Chrome might handle it
	} else if u.Scheme == "" && !strings.HasPrefix(urlStr, "http") {
		// If it's bare text, try to force http
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

func processURL(urlStr string) {
	profile := chooseProfile(urlStr, compiledRules, config.DefaultProfileDirectory)
	fmt.Fprintf(os.Stderr, "Routing: %s  ->  profile-directory=%q\n", urlStr, profile)

	if err := openInChrome(config.ChromeAppPath, profile, urlStr); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open URL in Chrome: %v\n", err)
	}
}

func main() {
	var err error
	config, compiledRules, err = loadConfig(defaultConfigPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(2)
	}

	go func() {
		timeout := time.After(4 * time.Second)
		select {
		case url := <-urlListener:
			processURL(url)
			os.Exit(0)
		case <-timeout:
			fmt.Fprintln(os.Stderr, "No URLs received within timeout")
			os.Exit(1)
		}
	}()

	C.RunApp()
}

//export HandleURL
func HandleURL(u *C.char) {
	urlListener <- C.GoString(u)
}
