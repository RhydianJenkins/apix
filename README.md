# APIX (API eXecuter)

**apix** is a CLI HTTP and gRPC client that manages requests across multiple API environments.

It handles authentication, domain switching, and [OpenAPI Specification (OAS)](https://swagger.io/specification/) (when using http) or [server reflection](https://grpc.io/docs/guides/reflection/) (when using gRPC).

## ✨ Features

- Manage multiple API domains under one tool
- Simple declarative `yaml` config
- Endpoint `<tab>` completions and discoverability via OAS or reflection

## 📦 Getting Started

```sh
go install github.com/rhydianjenkins/apix@latest

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
apix -c .apix.yaml.example get /api/v2/pokemon/ditto
```

## 🚀 Usage

```
API eXecuter (APIX) is a CLI tool to manage API domains and make requests

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
  show        Show a summary of an endpoint from the connected OpenAPI spec
  switch      Sets the active domain to the specified name
  version     Print version information

Flags:
  -c, --config string   path to the apix config file (default ~/.apix.yaml)
  -h, --help            help for apix

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
<summary><strong>How do I discover endpoints/methods without a shell (e.g. scripts, AI agents)?</strong></summary>

Shell `<tab>` completion only works in an interactive shell with a completion script loaded, so it's not usable by scripts, CI, or AI agents driving `apix` directly. Use `apix discover` instead - it prints the same information non-interactively, with a non-zero exit code on failure:

```sh
apix discover                # every endpoint/method on the active domain
apix discover --method post  # http domains only - filter by verb
```

Output depends on the active domain's protocol:

```
# http (from the connected OAS spec)
GET     /users - List users
POST    /users - Create a user

# grpc (via server reflection) - includes request/response message types
mypkg.UserService/GetUser(mypkg.GetUserRequest) -> mypkg.User
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
<summary><strong>How do I connect to a non-default port?</strong></summary>

For both `http` and `grpc` domains, `base` must be host-only - it can never include a port itself. If your API listens on a non-default port, set it separately with `--port` (or `http.port`/`grpc.port` in `.apix.yaml`):

```sh
apix new myapi https://api.example.com --port 8443
```

If you omit `--port`, the scheme's default port is used as normal (`http` domains only - `grpc` domains always require `--port`, since there's no equivalent default).

</details>

<details>
<summary><strong>How do I use gRPC?</strong></summary>

Create a domain with `--protocol grpc`, pointing `base` at the host of your gRPC server and `--port` at its port. `base` must be host-only - the port always comes from `--port` (or `grpc.port` in `.apix.yaml`), never from `base` itself:

```sh
# TLS (default)
apix new mygrpc grpc.example.com --protocol grpc --port 443

# plaintext, e.g. for local development
apix new mygrpc localhost --protocol grpc --insecure --port 50051
```

Then invoke a unary method by name using `apix grpc`. If the method takes no input (or you're happy sending an empty request), just run it directly:

```sh
apix grpc GetUser
# ...or qualify it if the method name is ambiguous across services
apix grpc mypkg.UserService/GetUser
```

If the method expects a request body, pipe it in as JSON, same as `apix post`/`apix put`. Field names must match the input message's `protojson` representation:

```sh
echo '{"name": "world"}' | apix grpc GetUser
# ...or from a file
cat req_body.json | apix grpc mypkg.UserService/GetUser
```

`apix` uses [server reflection](https://grpc.io/docs/guides/reflection/) to discover services and methods, so the target server must have reflection enabled. List them with `<tab>` completion or `apix discover` (see above).

</details>

<details>
<summary><strong>How do I send headers?</strong></summary>

You can add a headers attribute to the `yaml` config file. Edit it with `apix edit`, then add the following to the domain you want to send headers with:

```yaml
...
domains:
  my-domain:
    ...
    headers:
      x-my-header: foobar
```

</details>

<details>
<summary><strong>How do I configure auth?</strong></summary>

`apix` is unopinionated about how a domain's auth works - add an `auth.command` to the domain in config (`apix edit`, or `apix new ... --auth-command '<cmd>'`) and run `apix auth` whenever you need to (re-)authenticate:

```sh
apix auth              # runs auth.command for the active domain
apix auth my-domain    # ...or for a named domain, without switching to it
apix auth --all        # runs it for every domain that has one configured
```

`command` is run through `sh -c`, with stdin/stderr inherited from your terminal - so interactive flows (an OAuth CLI, a "approve on your phone" prompt, etc.) work as expected. What happens to its stdout depends on whether `header` is set:

```yaml
domains:
  # command's stdout becomes the authorization header/metadata value
  some-oauth-gated-api:
    auth:
      command: "my-auth-cli get_token"
      header: authorization

  # no header - command is just run for its side effects (e.g. allow-listing your IP)
  some-ip-gated-api:
    auth:
      command: "curl -sf -X POST https://admin.example.com/allowlist/me"
```

If a request to a domain with `auth.command` configured gets back a 401 (http) or an Unauthenticated status (grpc), `apix` prints a one-line hint to run `apix auth` for that domain.

</details>
