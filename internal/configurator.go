package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Application struct {
	Version     string
	ShutdownSec int

	Tracing    tracingConfig
	Monitoring monitoringConfig
	Http       httpConfig
	Grpc       grpcConfig
	Kafka      kafkaConfig
	Database   databaseConfig
}

func (app *Application) GetShutdownTime() time.Duration {
	return time.Duration(app.ShutdownSec) * time.Second
}

type tracingConfig struct {
	Disabled    bool
	ServiceName string

	GrpcEndpoints map[string]string
}

type monitoringConfig struct {
	HttpRoute      string
	StatusCounters []monitoringCounterConfig
}

type monitoringCounterConfig struct {
	Name string
	Desc string

	GrpcStatus    string
	GrpcEndpoints []string
}

type httpConfig struct {
	Addr string
}

type grpcConfig struct {
	Addr string
}

type kafkaConfig struct {
	Brokers []string
}

type databaseConfig struct {
	Driver string
	Host   string
	Port   string
	User   string
	Pass   string
	Db     string

	MaxOpenConns       int
	MaxIdleConns       int
	ConnTimeoutMs      int
	ConnMaxLifetimeSec int
}

func (dc *databaseConfig) GetConnTimeout() time.Duration {
	return time.Duration(dc.ConnTimeoutMs) * time.Millisecond
}

func (dc *databaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(dc.ConnMaxLifetimeSec) * time.Second
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
