package datasource

import (
	"github.com/jmoiron/sqlx"
	"time"
	"wizapp/sdk/app"
)

type (
	Config struct {
		ConnectionString      string        `mapstructure:"connection_string"`
		DriverName            string        `mapstructure:"driver_name"`
		MaxOpenConnections    int           `mapstructure:"max_open_connections"`
		MaxIdleConnections    int           `mapstructure:"max_idle_connections"`
		MaxConnectionLifeTime time.Duration `mapstructure:"max_connection_life_time"`
		MaxConnectionIdleTime time.Duration `mapstructure:"max_connection_idle_time"`
	}

	Datasource struct {
		config map[string]Config
	}

	ErrDataSource string
)

func (e ErrDataSource) Error() string {
	return string(e)
}

const (
	DefaultDatasourceName = "default"

	ErrDataSourceNotConfigured ErrDataSource = "Datasource not configured"
)

// LoadFromConfig
//
// Creates a Datasource instance from the configuration
// It expects the following yaml values:
//
// datasource:
//
//	  default:
//	    connection_string: root:password@tcp(localhost:3306)/localdb?multiStatements=true&parseTime=true
//		driver_name: mysql # Check here the list of available drivers https://zchee.github.io/golang-wiki/SQLDrivers/
//	    max_open_connections: 0
//	    max_idle_connections: 0
//	    max_connection_life_time: 0s
//	    max_connection_idle_time: 0s
//
//	  my-second-datasource:
//	    connection_string: postgresql://root:password@localhost/p_localdb?sslmode=disable
//		driver_name: postgres # Check here the list of available drivers https://zchee.github.io/golang-wiki/SQLDrivers/
//	    max_open_connections: 0
//	    max_idle_connections: 0
//	    max_connection_life_time: 0s
//	    max_connection_idle_time: 0s
func LoadFromConfig(cfg *app.ApplicationConfig) (*Datasource, error) {
	var dsCfg map[string]Config
	if err := cfg.UnmarshalKey("datasource", &dsCfg); err != nil {
		return nil, err
	}
	return Load(dsCfg)
}

// Load
//
// Creates a Datasource instance from the provided configuration map where
//
//	key: datasource name
//	value: datasource configuration
func Load(config map[string]Config) (*Datasource, error) {
	return &Datasource{config: config}, nil
}

// LoadWithDefault
//
// Creates a Datasource instance with  the DefaultDatasourceName
func LoadWithDefault(config Config) (*Datasource, error) {
	return Load(map[string]Config{DefaultDatasourceName: config})
}

// GetConnection
//
// Creates a sqlx.DB instance for the provided datasource name
// It can throw the error ErrDataSourceNotConfigured if the provided datasource name is not registered
func (ds *Datasource) GetConnection(name string) (*sqlx.DB, error) {
	c, ok := ds.config[name]
	if !ok {
		return nil, ErrDataSourceNotConfigured
	}

	db, err := sqlx.Open(c.DriverName, c.ConnectionString)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(c.MaxOpenConnections)
	db.SetConnMaxLifetime(c.MaxConnectionLifeTime)
	db.SetMaxIdleConns(c.MaxIdleConnections)
	db.SetConnMaxIdleTime(c.MaxConnectionIdleTime)

	return db, nil
}

// GetDefaultConnection
//
// Creates a sqlx.DB instance for the DefaultDatasourceName
// It can throw the error ErrDataSourceNotConfigured if the DefaultDatasourceName is not defined
func (ds *Datasource) GetDefaultConnection() (*sqlx.DB, error) {
	return ds.GetConnection(DefaultDatasourceName)
}
