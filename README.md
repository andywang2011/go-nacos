# go-nacos

Testable Nacos configuration client: download remote YAML, watch changes, and write local config files with MD5 verification.

## Install

```bash
go get github.com/andywang2011/go-nacos@v0.1.0
```

## Quick start

```go
if nacos.ShouldSkip("") {
	return
}
cfg, err := nacos.LoadConfigFromEnvAndMeta("./config/nacos.yaml", nacos.NamespaceFromEnv())
if err != nil {
	// handle
}
client, err := nacos.NewClient(cfg)
if err != nil {
	// handle
}
content, err := client.Download(context.Background())
_ = content
```

## Examples

```bash
cd go-nacos/examples/watch && go run .
```

## License

MIT — see [LICENSE](LICENSE).
