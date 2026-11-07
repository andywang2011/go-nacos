package nacos

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Meta holds data_id/group resolved from a local nacos meta YAML file.
type Meta struct {
	DataID   string
	Group    string
	PollTime time.Duration
	Timeout  time.Duration
}

// LoadMetaFromYAML reads namespace-specific keys from a nacos meta file.
func LoadMetaFromYAML(path, namespace string) (Meta, error) {
	if namespace == "" {
		return Meta{}, fmt.Errorf("nacos: namespace is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Meta{}, fmt.Errorf("nacos: read meta file %q: %w", path, err)
	}
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return Meta{}, fmt.Errorf("nacos: parse meta yaml: %w", err)
	}
	ns, ok := root[namespace].(map[string]any)
	if !ok {
		return Meta{}, fmt.Errorf("nacos: namespace %q not found in %q", namespace, path)
	}
	meta := Meta{
		DataID: getString(ns, "data_id"),
		Group:  getString(ns, "group"),
	}
	if meta.DataID == "" || meta.Group == "" {
		return Meta{}, fmt.Errorf("nacos: data_id and group required for namespace %q", namespace)
	}
	if v := getString(ns, "poll_time"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Meta{}, fmt.Errorf("nacos: invalid poll_time: %w", err)
		}
		meta.PollTime = d
	}
	if v := getString(ns, "timeout"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Meta{}, fmt.Errorf("nacos: invalid timeout: %w", err)
		}
		meta.Timeout = d
	}
	return meta, nil
}

// LoadConfigFromEnvAndMeta builds Config from NACOS_* env vars and a meta YAML file.
func LoadConfigFromEnvAndMeta(metaPath, namespace string) (Config, error) {
	meta, err := LoadMetaFromYAML(metaPath, namespace)
	if err != nil {
		return Config{}, err
	}
	port, err := strconv.ParseUint(os.Getenv(EnvPort), 10, 64)
	if err != nil || port == 0 {
		return Config{}, fmt.Errorf("nacos: %s is required", EnvPort)
	}
	cfg := Config{
		Address:     os.Getenv(EnvServer),
		Port:        port,
		Username:    os.Getenv(EnvUsername),
		Password:    os.Getenv(EnvPassword),
		NamespaceID: namespace,
		DataID:      meta.DataID,
		Group:       meta.Group,
		PollTime:    meta.PollTime,
		Timeout:     meta.Timeout,
	}
	return cfg, cfg.Validate()
}

// ShouldSkip reports whether Nacos bootstrap should be skipped.
func ShouldSkip(skipFlag string) bool {
	return skipFlag == "Y" || os.Getenv(EnvSkip) == "Y"
}

// NamespaceFromEnv returns EnvNamespace.
func NamespaceFromEnv() string {
	return os.Getenv(EnvNamespace)
}

func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}
