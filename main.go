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

	"github.com/sirupsen/logrus"
)

type Rule struct {
	Pattern          string `json:"pattern"`
	ProfileDirectory string `json:"profile_directory"`
}

type StrategyForUnknownUrls string

const (
	StrategyForUnknownUrlsUseBrowserDefault StrategyForUnknownUrls = "use-browser-default"
	StrategyForUnknownUrlsUseDefaultProfile StrategyForUnknownUrls = "use-default-profile"
)

type Config struct {
	ChromeAppPath           string                 `json:"chrome_app_path"`
	DefaultProfileDirectory string                 `json:"default_profile_directory"`
	StrategyForUnknownUrls  StrategyForUnknownUrls `json:"strategy_for_unknown_urls"`
	Rules                   []Rule                 `json:"rules"`
	LogLevel                string                 `json:"log_level"`
	compiledRules           []compiledRule
	parsedLogLevel          logrus.Level
}

type compiledRule struct {
	re               *regexp.Regexp
	profileDirectory string
}

var urlListener chan string = make(chan string)
var lockFilePath string = filepath.Join("/tmp", "chrome-profile-router.lock")
var logFilePath string = filepath.Join("/tmp", "chrome-profile-router.log")
var logger *logrus.Logger = nil

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "chrome-profile-router", "config.json")
}

func loadConfig(path string) (Config, error) {
	var cfg Config

	f, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config JSON: %w", err)
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
			return cfg, fmt.Errorf("rule %d invalid: pattern and profile_directory are required", i)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return cfg, fmt.Errorf("rule %d: compile regexp: %w", i, err)
		}
		cr = append(cr, compiledRule{re: re, profileDirectory: r.ProfileDirectory})
	}
	cfg.compiledRules = cr

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	parsedLogLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return cfg, fmt.Errorf("parse log level: %w", err)
	}
	cfg.parsedLogLevel = parsedLogLevel

	return cfg, nil
}

func chooseProfile(urlStr string, config Config) string {
	for _, r := range config.compiledRules {
		if r.re.MatchString(urlStr) {
			return r.profileDirectory
		}
	}
	if config.StrategyForUnknownUrls == StrategyForUnknownUrlsUseDefaultProfile {
		return config.DefaultProfileDirectory
	}
	return "" // StrategyForUnknownUrlsUseBrowserDefault
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
	}
	if profileDir != "" {
		args = append(args, fmt.Sprintf("--profile-directory=%s", profileDir))
	}
	args = append(args, urlStr)

	cmd := exec.Command("open", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func processURL(urlStr string, config Config) {
	profile := chooseProfile(urlStr, config)
	logger.Debugf("Routing: %s  ->  profile-directory=%q\n", urlStr, profile)

	if err := openInChrome(config.ChromeAppPath, profile, urlStr); err != nil {
		logger.Errorf("Failed to open URL in Chrome: %v\n", err)
	}
}

func main() {
	// load config
	config, err := loadConfig(defaultConfigPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(2)
		return
	}

	// initialize logger
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		os.Exit(2)
		return
	}
	logger = logrus.New()
	logger.SetOutput(logFile)
	logger.SetLevel(config.parsedLogLevel)
	defer logFile.Close()

	// exit if another instance is running
	f, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		logger.Debugln("Another instance is already running, exiting")
		os.Exit(0)
		return
	}
	defer func() {
		f.Close()
		os.Remove(lockFilePath)
	}()

	go func() {
		for url := range urlListener {
			processURL(url, config)
		}
	}()

	C.RunApp()
}

//export HandleURL
func HandleURL(u *C.char) {
	urlListener <- C.GoString(u)
}
