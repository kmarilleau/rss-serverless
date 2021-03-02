package gcfuncs

import "os"

func GetEnv() string {
	switch os.Getenv("GOOGLE_CLOUD_PROJECT") {
	case "PROD":
		return "PROD"
	default:
		return "TEST"
	}
}

func GetProjectID(env string) string {
	switch env {
	case "PROD":
		return os.Getenv("GOOGLE_CLOUD_PROJECT")
	default:
		return "test-project"
	}
}
