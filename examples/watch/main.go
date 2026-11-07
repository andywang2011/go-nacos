// Watch remote Nacos config and write conf.yaml on change (requires env NACOS_*).
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andywang2011/go-nacos"
)

func main() {
	if nacos.ShouldSkip(os.Getenv("SKIP_NACOS")) {
		fmt.Println("nacos bootstrap skipped")
		return
	}
	cfg, err := nacos.LoadConfigFromEnvAndMeta("./config/nacos.yaml", nacos.NamespaceFromEnv())
	if err != nil {
		log.Fatal(err)
	}
	client, err := nacos.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	out := "./config/conf.yaml"
	if content, err := client.Download(context.Background()); err == nil {
		_ = nacos.WriteConfigFile(out, []byte(content))
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	err = client.Watch(ctx, func(data []byte) error {
		return nacos.WriteConfigFile(out, data)
	})
	if err != nil {
		log.Fatal(err)
	}
	// Watch returns immediately; block until a signal cancels ctx (which also
	// cancels the Nacos listener), then Close releases anything left.
	<-ctx.Done()
}
