# apix (API eXecuter)

⚠️ **Warning: apix is in active development.**

**apix** is a lightweight CLI tool to manage and interact with multiple API domains. It supports user/password auth, domain switching, and prettified HTTP responses—all configured via a simple local YAML file.

## ✨ Features

- Add and manage multiple API domains
- Store credentials securely in a local config
- Switch between domains easily
- Make GET requests to API endpoints
- Pretty JSON output for responses

## 📦 Installation

```sh
# from source
go build -o apix && mv apix /usr/local/bin
```

## 🚀 Usage

```
API eXecuter (APIX) is a CLI tool to manage API domains and make requests

Usage:
   [flags]
   [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  get         Send a GET request to the active domain
  help        Help about any command
  list        List all domain names saved in config
  remove      Remove a domain from the config
  set         Set a new or existing API domain
  use         Sets the active domain to the specified name

Flags:
  -h, --help   help for this command

Use " [command] --help" for more information about a command.
```
