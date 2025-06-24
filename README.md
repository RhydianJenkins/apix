# APIX (API eXecuter)

⚠️ **Warning: APIX is in active development.**

**apix** is a lightweight CLI tool to manage and interact with multiple API domains. It supports user/password auth, domain switching, and prettified HTTP responses—all configured via a simple local YAML file.

## ✨ Features

- Add and manage multiple API domains
- Store credentials in a local yaml config
- OAS support (openapi 3+) and endpoint completion
- Switch between domains easily
- Make requests to your stored API endpoints

## 📦 Getting Started

```sh
go install github.com/rhydianjenkins/apix@latest
```

<details>
<summary>Want to <strong>build from source instead?</strong></summary>

```sh
# fetch the project
git clone https://github.com/rhydianjenkins/apix && cd apix

# then build from source
go build -o apix && ./apix

# or run with nix
nix run
```

</details>

## 🚀 Usage

```
API eXecuter (APIX) is a CLI tool to manage API domains and make requests

Usage:
  apix [command] [flags]
  apix [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  delete      Send a DELETE request to the active domain
  edit        Open config in your $EDITOR
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
<summary><strong>How do I send a request body?</strong></summary>

`apix` aims to make use of pipelines as much as possible and takes the request body from `stdin`.

```sh
# Make a POST request to {base}/users with `my_req_body.json` as a body
cat my_req_body.json | apix post /users
```
</details>

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
<summary><strong>How do I get it working with OAS?</strong></summary>

`APIX` can do:

- `json` and `yaml` spec formats
- openapi versions `3.0.0` and `3.1.0`
- remote and local specs

You can specify which spec belongs to which domain in config:

```sh
apix new myapi https://api.example.com --oas "https://api.example.domain/myOAS.json"
# ... or
apix new myapi https://api.example.com --oas "/local/path/to/myOAS.yaml"
```

Assuming you've set up shell completions, you should be able to `<tab>` complete your endpoints!

</details>
