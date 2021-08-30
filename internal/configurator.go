package internal

import (
	"encoding/json"
	"fmt"
	"os"
)

type Application struct {
	Version  string
	Grpc     grpcConfig
	Database databaseConfig
}

type grpcConfig struct {
	Addr string
}

type databaseConfig struct {
	Driver string
	Host   string
	Port   string
	User   string
	Pass   string
	Db     string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

func (dc *databaseConfig) String() string {
	switch dc.Driver {
	case "pgx":
		return fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s",
			dc.User,
			dc.Pass,
			dc.Host,
			dc.Port,
			dc.Db,
		)
	default:
		panic("driver not supported")
	}
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
