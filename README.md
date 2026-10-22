# APIX (API eXecuter)

**apix** is a CLI HTTP and gRPC client that manages requests across multiple API environments.

It handles authentication, domain switching, and [OpenAPI Specification (OAS)](https://swagger.io/specification/) (when using http) or [server reflection](https://grpc.io/docs/guides/reflection/) (when using gRPC).

## ✨ Features

- Manage multiple API domains under one tool
- Simple declarative `yaml` config
- Endpoint `<tab>` completions and discoverability via OAS or reflection

## 📦 Getting Started

```sh
go install github.com/rhydianjenkins/apix/v3/cmd/apix@latest

# ...or with nix
nix run github:rhydianjenkins/apix
```

<details>
<summary>Want to <strong>build from source instead?</strong></summary>

```sh
git clone https://github.com/rhydianjenkins/apix && cd apix
nix run .#build && ./apix
```

</details>

Want to try it out first? Point `-c` at the bundled example config to query the public [PokéAPI](https://pokeapi.co/) without setting anything up:

```sh
apix -c ./examples/pokeapi.yaml get /api/v2/pokemon/ditto
```

## 🚀 Usage

```
Usage:
  apix [command] [flags]
  apix [command]

Available Commands:
  auth        Run a domain's configured auth command (see `auth.command` in config)
  completion  Generate the autocompletion script for the specified shell
  delete      Send a DELETE request to the active domain
  discover    List the endpoints or gRPC methods available on the active domain
  edit        Open config in your $EDITOR
  get         Send a GET request to the active domain
  grpc        Invoke a unary gRPC method on the active domain
  help        Help about any command
  list        List all domain names saved in config
  new         Create a new API domain
  patch       Send a PATCH request to the active domain
  post        Send a POST request to the active domain
  put         Send a PUT request to the active domain
  remove      Remove a domain from the config
  set-header  Set a header/metadata value on a domain, without changing the active domain
  show        Show a summary of an HTTP endpoint (from the connected OpenAPI spec) or a gRPC method (via server reflection)
  switch      Sets the active domain to the specified name
  version     Print version information

Flags:
  -c, --config stringArray   path to an apix config file; repeat to layer files, later ones override earlier ones and receive all writes (default ~/.apix.yaml)
  -h, --help                 help for apix
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

Assuming you've set up shell completions, you should be able to `<tab>` complete your endpoints! You can also list them non-interactively with `apix discover` (see above).

</details>

<details>
<summary><strong>How do I configure auth?</strong></summary>

`apix` is unopinionated about how a domain's auth works - add an `auth.command` to the domain in config (`apix edit`, or `apix new ... --auth-command '<cmd>'`) and run `apix auth` whenever you need to (re-)authenticate:

```sh
apix auth              # runs auth.command for the active domain
apix auth my-domain    # ...or for a named domain, without switching to it
apix auth --all        # runs it for every domain that has one configured
```

`command` is run through `sh -c`, with stdin/stdout/stderr all inherited from your terminal - so interactive flows (an OAuth CLI, a "approve on your phone" prompt, etc.) work as expected. `apix` never captures or interprets the command's output - if it needs to save a header/metadata value (e.g. a bearer token), the command is responsible for doing that itself, via `apix set-header`:

```yaml
domains:
  # the command saves its own header, using `apix set-header`
  some-oauth-gated-api:
    auth:
      command: "apix set-header authorization \"Bearer $(my-auth-cli get_token)\""

  # ...or just run for its side effects (e.g. allow-listing your IP) - no header needed
  some-ip-gated-api:
    auth:
      command: "curl -sf -X POST https://admin.example.com/allowlist/me"
```

</details>

<details>
<summary><strong>How do I contribute?</strong></summary>

Fork, branch, and open a PR. Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, etc.) - versioning and `CHANGELOG.md` are generated from them automatically.

</details>
