# Chrome Profile Router

A Go-based utility that automatically routes URLs to different Chrome profiles based on configurable rules. Perfect for managing multiple Chrome profiles for different purposes (work, personal, development, etc.).

## Features

- **Smart URL Routing**: Automatically opens URLs in the appropriate Chrome profile based on regex patterns
- **macOS Optimized**: Uses native macOS `open` command for seamless Chrome integration
- **Flexible Configuration**: JSON-based configuration with support for multiple routing rules
- **Fallback Support**: Default profile for URLs that don't match any rules
- **Dry Run Mode**: Test your routing configuration without actually opening Chrome

## Use Cases

- **Work vs Personal**: Route work-related URLs to your work Chrome profile
- **Development**: Route development URLs to a profile with developer tools and extensions
- **Multiple Projects**: Separate Chrome profiles for different client projects
- **Security**: Isolate sensitive browsing to specific profiles

## Installation

### Prerequisites

- Go 1.16 or later
- macOS (uses macOS-specific Chrome launching)
- Google Chrome installed

### Build from Source

```bash
git clone <repository-url>
cd chrome-profile-router
go build -o chrome-profile-router main.go
```

### Install to System Path (Optional)

```bash
# Copy to a directory in your PATH
sudo cp chrome-profile-router /usr/local/bin/
```

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
- Custom names you've set

## Usage

### Basic Usage

```bash
# Route a single URL
./chrome-profile-router "https://github.com/yourcompany/project"

# Route multiple URLs
./chrome-profile-router "https://github.com/yourcompany/project" "https://stackoverflow.com/questions/123"
```

### Command Line Options

```bash
./chrome-profile-router [options] [URLs...]

Options:
  -config string
        Path to config JSON (default: ~/.config/chrome-profile-router/config.json)
  -dry-run
        Don't launch Chrome; just print routing decisions
  -version
        Print version
```

### Examples

```bash
# Test configuration without opening Chrome
./chrome-profile-router -dry-run "https://github.com/yourcompany/project"

# Use custom config file
./chrome-profile-router -config ./my-config.json "https://example.com"

# Route URLs from stdin
echo "https://github.com/yourcompany/project" | ./chrome-profile-router
```

### Setting as Default Browser

To use Chrome Profile Router as your default browser on macOS:

1. Build the application
2. In System Preferences > General > Default web browser, select Chrome Profile Router
3. Or use the command line:
   ```bash
   open -a "Chrome Profile Router"
   ```

## How It Works

1. **URL Analysis**: The router receives URLs from command line arguments or stdin
2. **Pattern Matching**: Each URL is tested against the regex patterns in your configuration
3. **Profile Selection**: The first matching rule determines which Chrome profile to use
4. **Chrome Launch**: Chrome is launched with the selected profile using macOS's `open` command
5. **Fallback**: If no rules match, the default profile is used

## Troubleshooting

### Common Issues

**Chrome doesn't open**
- Verify Chrome is installed at the configured path
- Check that profile directories exist
- Ensure you have permission to launch Chrome

**URLs not routing correctly**
- Test your regex patterns with `-dry-run` flag
- Verify profile directory names match exactly
- Check the configuration file syntax

**Permission denied errors**
- Ensure the binary is executable: `chmod +x chrome-profile-router`
- Check file permissions on your config directory

### Debug Mode

Use the `-dry-run` flag to see routing decisions without opening Chrome:

```bash
./chrome-profile-router -dry-run "https://example.com"
```

This will output:
```
Routing: https://example.com  ->  profile-directory="Default"
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[Add your license information here]

## Version

Current version: 1.0.0
