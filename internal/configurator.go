package internal

import (
	"encoding/json"
	"os"
)

type application struct {
	Version string
}

func LoadConfig(configDir string) {
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

	configPath := configDir + "/configs/app.json"

	for i := 0; i < 5; i++ {
		app := &application{}
		err := readConfig(app, configPath)
		if err != nil {
			panic(err)
		}
	}
}
