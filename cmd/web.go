package main

import (
	"github.com/starudream/go-lib/cobra/v2"
	"github.com/starudream/go-lib/core/v2/config"

	"github.com/starudream/clash-speedtest/web"
)

var webCmd = cobra.NewCommand(func(c *cobra.Command) {
	c.Use = "web"
	c.Short = "Start web UI server"

	c.PersistentFlags().Int("port", 8080, "web server port")

	c.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		config.LoadFlags(c.PersistentFlags())
	}

	c.RunE = func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		server := web.NewServer(port)
		return server.Start()
	}
})

func init() {
	rootCmd.AddCommand(webCmd)
}
