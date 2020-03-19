package file

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

// set a single env var
func SetEnvVar(key, variable string) error {
	return errors.WithMessage(os.Setenv(key, variable), "setting env variable")
}

// Set env vars from .env file globally
func SetEnvFileVars(dotPath string) (path string, err error) {
	if err := os.Chdir(dotPath); err != nil {
		return "", errors.Cause(err)
	}
	path, err = os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "change path err:")
	}
	if err := godotenv.Load(); err != nil {
		return "", errors.Wrap(err, "load dotenv file")
	}
	return path, err
}
