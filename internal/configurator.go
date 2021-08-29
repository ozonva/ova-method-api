package internal

import (
	"encoding/json"
	"os"
)

type Application struct {
	Version string
	Grpc    grpcConfig
}

type grpcConfig struct {
	Addr string
}

func LoadConfig(projectDir string) *Application {
	readConfig := func(app interface{}, path string) (err error) {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			err = f.Close()
		}()

		return json.NewDecoder(f).Decode(app)
	}

	configPath := projectDir + "/configs/app.json"

	app := &Application{}
	err := readConfig(app, configPath)
	if err != nil {
		panic(err)
	}

	return app
}
