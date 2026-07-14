# Hike

Hike contains the bits and pieces of code we reuse across our Go projects: layered configuration,
encryption and secret token handling, and HTTP request parsing.

It is a library — there is no binary. Add it with:

```bash
go get github.com/urban-connect/hike-go
```

## Packages

### `config` — layered configuration

`Reader` is the core abstraction. Each reader fills in a config struct you own, and you compose
them by running several in order — later readers override earlier ones.

```go
type Config struct {
    config.BaseConfig

    API struct {
        Host string `json:"host" env:"api_host"`
        Port int    `json:"port" env:"api_port"`
    } `json:"api"`
}

func NewConfig() (*Config, error) {
    cfg := &Config{}

    readers := []config.Reader{
        config.FromEnv("MYAPP_"),                        // MYAPP_API_HOST, MYAPP_CRYPTO_KEY, ...
        config.FromFile("config/settings.json", false),  // required
        config.FromFile("config/settings.local.json", true), // optional, gitignored
    }

    for _, reader := range readers {
        if err := reader.Read(cfg); err != nil {
            return nil, err
        }
    }

    return cfg, nil
}
```

Embed `config.BaseConfig` to pick up `env`, `version` and the `crypto` key/nonce fields. Fields carry
both `json` and `env` tags because the same struct is filled from either source. Env lookup is
prefixed and upper-cased, so `env:"crypto_key"` with prefix `MYAPP_` reads `MYAPP_CRYPTO_KEY`.

For secrets committed to the repo, `FromEncryptedFile` decrypts before parsing. It takes anything
satisfying `config.Crypto` (`Encrypt`/`Decrypt`) — `crypto.AES` does:

```go
aes, err := crypto.NewAES(cfg.Crypto.Key, cfg.Crypto.Nonce)
if err != nil {
    return nil, err
}

reader := config.FromEncryptedFile("config/settings.json.encrypted", true, aes)
```

`config.NewLogger(cfg.BaseConfig)` returns a stdlib `*slog.Logger` — a JSON handler at info level
when `env` is `production`, and a human-readable text handler at debug level everywhere else.

### `crypto` — encryption and tokens

`crypto.AES` is AES-GCM with a fixed key and nonce, both base64-encoded. The fixed nonce is
deliberate: it makes encrypting a config file reproducible, so re-encrypting unchanged settings
produces no diff. That also means **it is not suitable for encrypting many distinct messages under
one key** — use it for config at rest, not as a general-purpose cipher.

`crypto.Token` wraps a secret string, hashes it with bcrypt for storage, and checks a candidate
against a stored digest — the usual pattern behind static access-token header auth:

```go
token, err := crypto.RandomToken()  // give this to the caller
digest, err := token.Digest()       // store this in settings

// later, on each request
err := crypto.NewToken(header).Validate(digest)
```

### `api` — request parameter binding

`api.ParseParams` binds request data into a struct via `params:"key,source"` tags, where `source` is
`path`, `query` (the default), or `header`. Path values use `r.PathValue`, i.e. Go 1.22 `net/http`
routing patterns.

```go
type Params struct {
    ID    string `params:"id,path"`
    Token string `params:"authorization,header"`
    Limit int    `params:"limit,query"`
}

var params Params
if err := api.ParseParams(r, &params); err != nil {
    // ...
}
```

Supported field kinds are string, int, int64, float32, float64, and bool. An absent value leaves the
field untouched rather than erroring.

## Design notes

The three packages are independent — none of them imports another, so you can use `api` without
pulling in `config`, and so on.

Hike stays deliberately thin on dependencies. Logging is stdlib `log/slog`, so the library brings no
logging framework of its own and leaves the choice of handler to you. That keeps the direct
requirements down to `golang.org/x/crypto` and `go-envconfig`.

The packages are written to stay generic rather than to serve one application. Readers take `any` and
fill a config struct you own, `NewLogger` takes the embeddable `BaseConfig` instead of your concrete
type, and `FromEncryptedFile` accepts any `config.Crypto` implementation rather than depending on the
`crypto` package directly.

## Development

```bash
go build ./...
go test ./...
go vet ./...
gofmt -l .   # must print nothing

# CI also gates on the modernize analyzer
go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest ./...
```

## License

MIT — see [LICENSE](LICENSE).
