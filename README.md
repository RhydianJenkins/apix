# apix (API eXecuter)

⚠️ **Warning: apix is in active development.**

**apix** is a lightweight CLI tool to manage and interact with multiple API domains. It supports user/password auth, domain switching, and prettified HTTP responses—all configured via a simple local YAML file.

## ✨ Features

- Add and manage multiple API domains
- Store credentials in a local yaml config
- OAS support and endpoint completion
- Switch between domains easily
- Make requests to your stored API endpoints

## 📦 Getting Started

```sh
# fetch the project
git clone https://github.com/rhydianjenkins/apix && cd apix

# then build from source
go build -o apix && ./apix

# or run with nix
nix run
```

## 🚀 Usage

```
API eXecuter (APIX) is a CLI tool to manage API domains and make requests

Usage:
  apix [command] [flags]
  apix [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  delete      Send a DELETE request to the active domain
  get         Send a GET request to the active domain
  help        Help about any command
  list        List all domain names saved in config
  new         Create a new API domain
  patch       Send a PATCH request to the active domain
  post        Send a POST request to the active domain
  put         Send a PUT request to the active domain
  remove      Remove a domain from the config
  switch      Sets the active domain to the specified name

Flags:
  -h, --help   help for apix

Use "apix [command] --help" for more information about a command.
```

## ❓ FAQ

<details>
<summary><strong>How do I enable shell completion?</strong></summary>

If you haven't already, add your completion directory to `fpath`:

```sh
# in .zshrc, or wherever
fpath=($HOME/.local/zsh-completions $fpath)
compinit
```

Then generate the completion script:

```sh
# if you're using zsh...
apix completion zsh > ~/.local/zsh-completions/_apix
```
</details>

<details>
<summary><strong>How do I get it working with my OAS?</strong></summary>

TODO Rhydian

`apix` supports `json` and `yaml` OAS, as well as remote (http) hosted and local hosted specs.

You can specifiy which spec belongs to which domain in config:

```sh
apix new myapi https://api.example.domain --oas "https://api.example.domain/myOAS.json"
# ... or
apix new myapi https://api.example.domain --oas "/local/path/to/myOAS.json"
```

</details>
