# apix (API eXecuter)

⚠️ **Warning: apix is in active development.**

**apix** is a lightweight CLI tool to manage and interact with multiple API domains. It supports user/password auth, domain switching, and prettified HTTP responses—all configured via a simple local YAML file.

## ✨ Features

- Add and manage multiple API domains
- Store credentials in a local yaml config
- Switch between domains easily
- Make requests to your stored API endpoints

## 📦 Getting Started

```sh
# fetch the project
git clone https://github.com/rhydianjenkins/apix && cd apix

# build from source
go build -o apix
./apix

# or with nix
nix run
```

## 🚀 Usage

```
API eXecuter (APIX) is a CLI tool to manage API domains and make requests

Usage:
   [flags]
   [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  delete      Send a DELETE request to the active domain
  get         Send a GET request to the active domain
  help        Help about any command
  list        List all domain names saved in config
  new         create a new API domain
  patch       Send a PATCH request to the active domain
  post        Send a POST request to the active domain
  put         Send a PUT request to the active domain
  remove      Remove a domain from the config
  use         Sets the active domain to the specified name

Flags:
  -h, --help   help for this command

Use " [command] --help" for more information about a command.
```
