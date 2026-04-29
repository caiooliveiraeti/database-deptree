package main

import (
	"github.com/caiooliveiraeti/database-deptree/internal/config"
	"github.com/caiooliveiraeti/database-deptree/internal/server"
	"github.com/caiooliveiraeti/database-deptree/internal/store"
	web "github.com/caiooliveiraeti/database-deptree/internal/web"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web UI to visualize the dependency graph",
		Example: `  deptree serve
  deptree serve --port=9090`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile)
			if err != nil {
				return err
			}

			st, err := store.NewNeo4jStore(cfg.Neo4j.URI, cfg.Neo4j.User, cfg.Neo4j.Password)
			if err != nil {
				return err
			}
			defer st.Close()

			return server.New(st, web.FS(), port).Run(cmd.Context())
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "port to listen on")
	return cmd
}
