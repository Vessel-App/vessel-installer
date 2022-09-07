# Vessel Installer

This is a tiny HTTP server that periodically retrieves the latest stable version based on GitHub Releases.

It offers just a few HTTP endpoints:

1. `/stable/version` - Get the latest version back in JSON
2. `/stable/install.sh` - Get a `sh` script to install `vessel` to `~/.vessel/bin` (MacOS / Linux)

To install Vessel, you can run this script:

```bash
curl <hostname>/stable/install.sh | sh
```

For installation instructions, see the [Vessel CLI repository](https://github.com/Vessel-App/vessel-cli).