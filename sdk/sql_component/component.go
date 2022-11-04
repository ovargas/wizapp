package sql_component

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/urfave/cli/v2"
	"wizapp/sdk/app"
	"wizapp/sdk/datasource"
	"wizapp/sdk/logger"
)

const (
	FlagDatasource = "datasource"
)

var (
	log = logger.Log()
)

type (
	Config struct {
		datasource.Config `mapstructure:",squash"`
		MigrationPath     string `mapstructure:"migration_path"`
	}

	component struct {
		config map[string]Config
	}
)

func init() {
	app.RegisterComponent("sql", createComponent)
}

func createComponent(cfg *app.ApplicationConfig) (app.Component, error) {
	var dsCfg map[string]Config
	if err := cfg.UnmarshalKey("datasource", &dsCfg); err != nil {
		return nil, err
	}
	return &component{config: dsCfg}, nil
}

func (c *component) Command() *app.Command {
	return &app.Command{
		Name:  "sql",
		Usage: "Sql database operations",
		Subcommands: []*app.Command{
			{
				Name:    "migrate",
				Aliases: []string{"m"},
				Usage:   "Apply migration scripts",
				Action:  c.migrate,
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Fetch database schema version",
				Action:  c.version,
			},
		},
		Flags: []app.Flag{
			&cli.StringFlag{
				Name:     "datasource",
				Aliases:  []string{"ds"},
				Usage:    "Datasource name",
				Required: true,
			},
		},
	}
}

func (c *component) migrate(ctx *cli.Context) error {
	name := ctx.String(FlagDatasource)
	m, err := c.getMigration(ctx)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		log.Errorf("error applying datasource %s migration: %v", name, err)
		return err
	}
	return nil
}

func (c *component) version(ctx *cli.Context) error {
	name := ctx.String(FlagDatasource)

	m, err := c.getMigration(ctx)
	if err != nil {
		return err
	}
	version, dirty, err := m.Version()
	if err != nil {
		log.Errorf("error fetching datasource %s schema version: %v", name, err)
		return err
	}

	fmt.Printf("Datasource %s version %d, dirty %t\n", ctx.String(FlagDatasource), version, dirty)
	return nil
}

func (c *component) getMigration(ctx *cli.Context) (*migrate.Migrate, error) {
	name := ctx.String(FlagDatasource)

	config, ok := c.config[name]

	if !ok {
		log.Errorf("datasource %s no configured", name)
		return nil, datasource.ErrDataSourceNotConfigured
	}

	log.Infof("migrating with %v", config)

	m, err := migrate.New(config.MigrationPath, fmt.Sprintf("%s://%s", config.DriverName, config.ConnectionString))
	if err != nil {
		log.Errorf("error creating migration %s: %v", name, err)
		return nil, err
	}

	return m, nil
}
