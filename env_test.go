package nacos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvConstantsUseNACOSPrefix(t *testing.T) {
	want := map[string]string{
		EnvNamespace: "NACOS_NAMESPACE",
		EnvServer:    "NACOS_SERVER",
		EnvPort:      "NACOS_PORT",
		EnvUsername:  "NACOS_USERNAME",
		EnvPassword:  "NACOS_PASSWORD",
		EnvSkip:      "NACOS_SKIP",
	}
	for got, exp := range want {
		if got != exp {
			t.Errorf("env const = %q, want %q", got, exp)
		}
		if strings.Contains(got, "VERY_STAR") {
			t.Errorf("legacy prefix in %q", got)
		}
	}
}

func TestLoadConfigFromEnvAndMeta_UsesNACOSEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nacos.yaml")
	if err := os.WriteFile(path, []byte("ns1:\n  data_id: app.yaml\n  group: g\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NACOS_SERVER", "nacos.example")
	t.Setenv("NACOS_PORT", "8848")
	t.Setenv("NACOS_USERNAME", "u")
	t.Setenv("NACOS_PASSWORD", "p")
	t.Setenv("NACOS_NAMESPACE", "ns1")

	if got := NamespaceFromEnv(); got != "ns1" {
		t.Fatalf("NamespaceFromEnv = %q", got)
	}
	cfg, err := LoadConfigFromEnvAndMeta(path, NamespaceFromEnv())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Address != "nacos.example" || cfg.Port != 8848 || cfg.Username != "u" || cfg.Password != "p" {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestLoadConfigFromEnvAndMeta_LegacyEnvIgnored(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nacos.yaml")
	if err := os.WriteFile(path, []byte("ns1:\n  data_id: app.yaml\n  group: g\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NACOS_PORT", "")
	t.Setenv("VERY_STAR_NACOS_PORT", "8848")
	if _, err := LoadConfigFromEnvAndMeta(path, "ns1"); err == nil {
		t.Fatal("legacy VERY_STAR_NACOS_PORT must not be honored")
	}
}
