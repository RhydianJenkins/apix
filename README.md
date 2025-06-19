# apix (API eXecuter)

⚠️ **Warning: apix is in active development.**

**apix** is a lightweight CLI tool to manage and interact with multiple API domains. It supports user/password auth, domain switching, and prettified HTTP responses—all configured via a simple local YAML file.

---

## ✨ Features

- Add and manage multiple API domains
- Store credentials securely in a local config
- Switch between domains easily
- Make GET requests to API endpoints
- Pretty JSON output for responses

---

## 📦 Installation

```sh
git clone https://github.com/yourusername/apix.git
cd apix
go build -o apix
sudo mv apix /usr/local/bin/
```

---

## 🚀 Usage

```sh
# Add a new domain named 'my-domain'
apix add my-domain https://api.example.com --user foo --pass bar

# List domains
apix list

# Switch to a domain
apix switch my-domain

# Make API calls to that domain
apix get /users # equivalent to `curl -u 'foo:bar' https://api.example.com/users`
```
