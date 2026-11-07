package nacos

// Deployment environment variable names shared by all services using go-nacos.
// Callers must not read these keys directly; use LoadConfigFromEnvAndMeta,
// NamespaceFromEnv, and ShouldSkip instead.
const (
	EnvNamespace = "NACOS_NAMESPACE"
	EnvServer    = "NACOS_SERVER"
	EnvPort      = "NACOS_PORT"
	EnvUsername  = "NACOS_USERNAME"
	EnvPassword  = "NACOS_PASSWORD"
	EnvSkip      = "NACOS_SKIP"
)
