# Chrome Profile Router

A native macOS application written in Go that automatically routes URLs to different Chrome profiles based on configurable rules. Perfect for managing multiple Chrome profiles for different purposes (work, personal, development, etc.).

## Features

- **Smart URL Routing**: Automatically opens URLs in the appropriate Chrome profile based on regex patterns
- **Native macOS App**: Built as a proper macOS application bundle with Objective-C bindings
- **Default Browser Integration**: Can be set as your system's default web browser
- **Flexible Configuration**: JSON-based configuration with support for multiple routing rules
- **Fallback Support**: Default profile for URLs that don't match any rules
- **Automatic URL Handling**: Processes URLs passed from the system when set as default browser

## Use Cases

- **Work vs Personal**: Route work-related URLs to your work Chrome profile
- **Development**: Route development URLs to a profile with developer tools and extensions
- **Multiple Projects**: Separate Chrome profiles for different client projects
- **Security**: Isolate sensitive browsing to specific profiles
- **Productivity**: Automatically organize your browsing across different contexts

## Installation

### Prerequisites

- Go 1.16 or later
- macOS (uses macOS-specific Chrome launching and Objective-C bindings)
- Google Chrome installed
- Xcode Command Line Tools (for Objective-C compilation)

### Build from Source

```bash
git clone <repository-url>
cd chrome-profile-router
make
```

This will create a `chrome-profile-router` binary into `ChromeProfileRouter.app` macOS application bundle.

And move the `ChromeProfileRouter.app` to the `/Applications/` folder

## Configuration

Create a configuration file at `~/.config/chrome-profile-router/config.json`. You can use the included `config.json.example` as a starting point:

```bash
# Copy the example configuration
cp config.json.example ~/.config/chrome-profile-router/config.json

# Edit the configuration for your needs
nano ~/.config/chrome-profile-router/config.json
```

Here's a basic example of what your configuration should look like:

```json
{
  "chrome_app_path": "/Applications/Google Chrome.app",
  "default_profile_directory": "Default",
  "strategy_for_unknown_urls": "use-default-profile",
  "log_level": "info",
  "rules": [
    {
      "pattern": "github\\.com/yourcompany",
      "profile_directory": "Work"
    },
    {
      "pattern": "stackoverflow\\.com",
      "profile_directory": "Development"
    },
    {
      "pattern": "gmail\\.com",
      "profile_directory": "Personal"
    }
  ]
}
```

### Configuration Options

- **`chrome_app_path`**: Path to Chrome application (defaults to `/Applications/Google Chrome.app`)
- **`default_profile_directory`**: Profile to use when no rules match (defaults to `"Default"`)
- **`log_level`**: Sets the verbosity of logging output. Options include `"debug"`, `"info"`, `"warn"`, and `"error"`. (defaults to `"info"`)
- **`strategy_for_unknown_urls`**: Strategy for handling URLs that don't match any rules
  - **`"use-default-profile"`**: Use the profile specified in `default_profile_directory`
  - **`"use-browser-default"`**: Let the system's default browser handle the URL (Chrome Profile Router won't interfere)
- **`rules`**: Array of routing rules
  - **`pattern`**: Regex pattern to match against URLs
  - **`profile_directory`**: Chrome profile directory name to use for matching URLs

### Finding Profile Directories

Chrome profile directories are located at:
```
~/Library/Application Support/Google/Chrome/
```

Common profile names include:
- `Default` - Default profile
- `Profile 1`, `Profile 2` - Additional profiles

## Usage

### As Default Browser (Recommended)

1. build and move app into /Applications/ folder
2. use the command line:
   ```bash
   # cli tool to set default browser
   brew install defaultbrowser

   # register the app to launch service
   /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -f /Applications/ChromeProfileRouter.app

   # set default browser
   defaultbrowser chromeprofilerouter
   ```
3. Now when you click links in other applications, they'll automatically route to the appropriate Chrome profile

## How It Works

1. **URL Reception**: The router receives URLs from the system when set as default browser, or from command line arguments
2. **Pattern Matching**: Each URL is tested against the regex patterns in your configuration
3. **Profile Selection**: The first matching rule determines which Chrome profile to use
4. **Chrome Launch**: Chrome is launched with the selected profile using macOS's `open` command
5. **Fallback Strategy**: If no rules match, the behavior depends on your `strategy_for_unknown_urls` setting:
   - **`use-default-profile`**: Opens the URL in Chrome using the profile specified in `default_profile_directory`
   - **`use-browser-default`**: Passes the URL to the system's default browser (Chrome Profile Router won't interfere)

### Technical Details

- **Objective-C Integration**: Uses CGO to interface with macOS Cocoa framework
- **Native App Bundle**: Creates a proper `.app` bundle for system integration
- **URL Handling**: Implements the macOS URL handling protocol for default browser functionality
- **Profile Management**: Leverages Chrome's `--profile-directory` argument for profile switching

## Troubleshooting

### Common Issues

**Chrome doesn't open**
- Verify Chrome is installed at the configured path
- Check that profile directories exist
- Ensure you have permission to launch Chrome

**URLs not routing correctly**
- Verify profile directory names match exactly
- Check the configuration file syntax
- Ensure regex patterns are valid

**Permission denied errors**
- Ensure the binary is executable: `chmod +x chrome-profile-router`
- Check file permissions on your config directory

**Build errors**
- Ensure Xcode Command Line Tools are installed: `xcode-select --install`
- Verify Go version is 1.16 or later: `go version`

## Development

### Project Structure

- `main.go` - Main Go application with URL routing logic
- `handler.h` - C header file for Objective-C integration
- `handle.m` - Objective-C implementation for macOS URL handling
- `Makefile` - Build automation for the macOS app bundle

### Building

```bash
# Build the binary and app bundle
make

# Clean build artifacts
make clean
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License

## Version

Current version: 1.0.0
