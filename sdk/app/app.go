package app

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	mu sync.RWMutex
)

type (
	Application interface {
		Config() *ApplicationConfig
	}

	Command = cli.Command
	Flag    = cli.Flag

	Setup func(config *ApplicationConfig) error
)

func Run(args []string, setup Setup) error {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config-path",
			EnvVars: []string{"CONFIG_PATH"},
			Value:   "./resources",
		},
		&cli.StringFlag{
			Name:    "active-profiles",
			EnvVars: []string{"ACTIVE_PROFILE"},
			Value:   "",
		},
	}

	cfg := Config()

	if err := setup(cfg); err != nil {
		return err
	}

	servers := make(map[string]Server)

	components := make(map[string]Component)
	for k, v := range componentFactories {
		c, err := v(cfg)
		if err != nil {
			log.Fatalf("unable to create component %s: %v", k, err)
		}
		components[k] = c
		app.Commands = append(app.Commands, c.Command())
	}

	app.Commands = append(app.Commands, &cli.Command{
		Name:   "start",
		Usage:  "Start registered servers",
		Flags:  disableServerFlags(),
		Before: prepare(cfg, servers),
		Action: start(servers),
	})

	return app.Run(args)
}

func disableServerFlags() []cli.Flag {
	var disableServerFlags []cli.Flag
	for k := range serverFactories {
		disableServerFlags = append(disableServerFlags,
			&cli.BoolFlag{
				Name:  fmt.Sprintf("disable-%s", k),
				Usage: fmt.Sprintf("Disable %s server", k),
				Value: false,
			})
	}
	return disableServerFlags
}

func prepare(cfg *ApplicationConfig, servers map[string]Server) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		for k, factory := range serverFactories {
			if ctx.Bool(fmt.Sprintf("disable-%s", k)) {
				continue
			}

			srv, err := factory(cfg)
			if err != nil {
				log.Fatalf("unable to start \"%s\" server: %v", k, err)
			}
			servers[k] = srv
		}

		return nil
	}
}

func start(servers map[string]Server) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, s := range servers {
			go func(srv Server) {
				if err := srv.Start(); err != nil {
					log.Fatalf("unable to start server: %v", err)
				}
			}(s)
		}
		wait(interruptCh(), servers)
		return nil
	}
}

func wait(interruptCh <-chan interface{}, servers map[string]Server) {
	select {
	case s := <-interruptCh:
		log.Printf("application stopped. Signal %v", s)
		for k, s := range servers {
			if err := s.Stop(); err != nil {
				log.Printf("error stopping %s server: %v", k, err)
			}
		}
	}
}

func interruptCh() <-chan interface{} {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ret := make(chan interface{}, 1)
	go func() {
		s := <-c
		ret <- s
		close(ret)
	}()

	return ret
}
