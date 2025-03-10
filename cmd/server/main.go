package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"gojini.dev/config"
	"gojini.dev/web"
)

const version = "unknown"

func main() {
	name := filepath.Base(os.Args[0])
	rootCmd := new(cobra.Command)

	rootCmd.Use = name
	rootCmd.Short = "zebra server"
	rootCmd.Version = version + "\n"
	rootCmd.RunE = run
	rootCmd.SilenceUsage = true
	rootCmd.SetVersionTemplate(version + "\n")
	rootCmd.Flags().StringP("config", "c", path.Join(
		func() string {
			s, _ := os.Getwd()

			return s
		}(), "server.json"),
		"config file (default: $PWD/server.json",
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Load server configuration
	cfgFile := cmd.Flag("config").Value.String()

	cfgStore := config.New()
	if err := cfgStore.LoadFromFile(context.Background(), cfgFile); err != nil {
		return err
	}

	return startServer(cfgStore)
}

func startServer(cfgStore *config.Store) error {
	appCtx := setupLogger(cfgStore)

	serverCfg := new(web.Config)
	if e := cfgStore.Get("server", serverCfg); e != nil {
		return e
	}

	setup := setupAdapter(appCtx, cfgStore)
	login := loginAdapter()
	register := registerAdapter()
	auth := authAdapter()
	refresh := refreshAdapter()
	routes := routeHandler()

	// The order of wrap matters, routes is the final handler that is being
	// wrapped. setup, login and register are unauthenticated APIs that serve
	// as a way to bootstrap authentication. auth, refresh and all endpoints
	// registered by routes must be authenticated either via a jwt in the cookie
	// or via a rsa key token in the header.
	handler := web.Wrap(routes, setup, login, register, auth, refresh)

	webServer := web.NewServer(serverCfg, handler)

	return webServer.Start(appCtx)
}

func callNext(nextHandler http.Handler, res http.ResponseWriter, req *http.Request) {
	if nextHandler != nil {
		nextHandler.ServeHTTP(res, req)
	}
}
