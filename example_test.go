package nacos_test

import (
	"context"

	"github.com/andywang2011/go-nacos"
)

func ExampleNewClient() {
	if nacos.ShouldSkip("") {
		return
	}
	cfg, err := nacos.LoadConfigFromEnvAndMeta("./config/nacos.yaml", nacos.NamespaceFromEnv())
	if err != nil {
		return
	}
	client, err := nacos.NewClient(cfg)
	if err != nil {
		return
	}
	content, err := client.Download(context.Background())
	if err != nil {
		return
	}
	_ = content
}
