package envx

import "os"

const (
	ENV_LOCAL = "local"
	ENV_DEV   = "dev"
	ENV_TEST  = "test"
	ENV_PROD  = "prod"
)

func GetEnv(name string) string {
	return os.Getenv(name)
}
