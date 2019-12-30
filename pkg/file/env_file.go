package file

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

// set a single env var
func SetEnvVar(key, variable string) error {
	return errors.WithMessage(os.Setenv(key, variable), "setting env variable")
}

// Set env vars from .env file globally FIXME: Not working
func SetEnvFileVars() error {
	return godotenv.Load()
}

func RetrieveFromEnvFile() {
	_, err := os.Open("./env_file.go")
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadFile(".env")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	barr := bytes.Split(b, []byte("="))
	fmt.Println(barr)

}
