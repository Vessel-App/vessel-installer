package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Version struct {
	Tag string `json:"tag_name"`
}

var stableVersion string

// getStableVersion retrieves the latest stable release for the Vessel CLI project
// If an error is encountered, the latest version is not updated
func setStableVersion() {
	resp, err := http.Get("https://api.github.com/repos/vessel-app/vessel-cli/releases/latest")

	if err != nil {
		log.Printf("Error retrieving Vessel CLI version: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		v := &Version{}
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			log.Printf("Error decoding JSON response: %v", err)
		}

		// Update the stable version that API calls to this endpoint will see
		stableVersion = v.Tag
	} else {
		r, _ := io.ReadAll(resp.Body) // Yes, I ignore a possible error here. Sorry.
		log.Printf("HTTP status error retrieving Vessel CLI version. Status: %d, Body: %s", resp.StatusCode, r)
	}
}

func GetStableVersion(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"version": "%s"}`, stableVersion)
}

func GetStableInstall(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, `#!/bin/sh
# TODO: Please keep this script extremely simple

set -e

# See https://superuser.com/a/1734874
case $(uname -s) in
    Darwin)    os='darwin';;
    Linux)     os='linux';;
    *)         echo "Vessel is not built for this OS!" >&2; exit 1;;
esac

arch=$(uname -m)
version=${1:-%s}

if [ "$arch" == "x86_64" ]; then
	arch="amd64"
fi

vessel_install="$HOME/.vessel"
bin_dir="$vessel_install/bin"
exe="$bin_dir/vessel"

if [ ! -d "$bin_dir" ]; then
	mkdir -p "$bin_dir"
fi

curl -q --fail --location --progress-bar --output "$exe.tar.gz" "https://github.com/Vessel-App/vessel-cli/releases/download/${version}/vessel-cli_${version}_${os}_${arch}.tar.gz"
cd "$bin_dir"
tar xzf "$exe.tar.gz"
chmod +x "$exe"
rm "$exe.tar.gz"

if [ "$os" == "darwin" ]; then
	echo "\n\033[0;32m\xE2\x9C\x94\033[0m vessel ${version} was installed successfully to $exe\n"
else
	echo -e "\n\033[0;32m\xE2\x9C\x94\033[0m vessel ${version} was installed successfully to $exe\n"
fi

cd $HOME
if command -v vessel >/dev/null; then
	echo "Run 'vessel --help' to get started\n"
else
	case $SHELL in
	/bin/zsh) shell_profile=".zshrc" ;;
	*) shell_profile=".bash_profile" ;;
	esac

	if [ "$os" == "darwin" ]; then
		echo "\033[0;33mNote:\033[0m Manually add the following to your \$HOME/$shell_profile (or similar)"
	else
		echo -e "\033[0;33mNote:\033[0m Manually add the following to your \$HOME/$shell_profile (or similar)"
	fi
	echo "  export PATH=\"${bin_dir}:\$PATH\"\n"
	echo "Run '$exe --help' to get started\n"
fi
`, stableVersion)
}

func main() {
	// Get the latest stable version on boot
	setStableVersion()

	// Get the latest stable version every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				setStableVersion()
			case <-quit:
				ticker.Stop()
			}
		}
	}()

	// Return the latest stable version
	http.Handle("/", http.RedirectHandler("https://github.com/Vessel-App/vessel-cli", 302))
	http.HandleFunc("/stable/version", GetStableVersion)
	http.HandleFunc("/stable/install.sh", GetStableInstall)
	http.ListenAndServe(":8080", nil)
}
