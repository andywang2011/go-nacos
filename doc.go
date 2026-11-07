// Package nacos provides a testable Nacos config client for downloading and watching remote YAML.
//
// Design constraints:
//   - Configuration via Config struct (no viper, no log.Fatal inside the library).
//   - NewClientWithAPI injects the config API for unit tests.
//   - NACOS_* env vars are owned by this package; use LoadConfigFromEnvAndMeta,
//     NamespaceFromEnv, and ShouldSkip rather than reading those keys elsewhere.
package nacos
