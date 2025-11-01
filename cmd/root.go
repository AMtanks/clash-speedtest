package main

import (
	"path/filepath"
	"strings"

	"github.com/starudream/go-lib/cobra/v2"
	"github.com/starudream/go-lib/core/v2/config"
	"github.com/starudream/go-lib/core/v2/utils/osutil"

	"github.com/starudream/clash-speedtest/api/cloudflare"
	"github.com/starudream/clash-speedtest/api/fast"
	"github.com/starudream/clash-speedtest/api/speedtest"
	"github.com/starudream/clash-speedtest/job"
)

var rootCmd = cobra.NewRootCommand(func(c *cobra.Command) {
	c.Use = "clash-speedtest"

	c.PersistentFlags().String("clash-addr", "http://127.0.0.1:9090", "clash external controller address")
	c.PersistentFlags().String("clash-secret", "", "clash external controller secret")
	c.PersistentFlags().String("clash-proxy", "", "clash proxy url, http or socks5")
	c.PersistentFlags().Uint16P("size", "r", 10, "download size for each thread")
	c.PersistentFlags().Uint16P("threads", "t", 1, "download threads for each type, each thread will took 2MB traffic")
	c.PersistentFlags().StringP("download", "d", cloudflare.Name, "download type: "+strings.Join(availDownloads, ", "))
	c.PersistentFlags().StringSliceP("includes", "i", []string{}, "include proxy names")
	c.PersistentFlags().StringSliceP("excludes", "e", []string{}, "exclude proxy names, after filter by include")
	c.PersistentFlags().BoolP("confirm", "y", false, "confirm to start speedtest")
	c.PersistentFlags().StringP("output", "o", filepath.Join(osutil.ExeDir(), "output"), "output file path")
	c.PersistentFlags().BoolP("ping", "p", false, "enable periodic ping test during speedtest")
	c.PersistentFlags().Int("ping-interval", 60, "ping interval in seconds")
	c.PersistentFlags().Int("ping-timeout", 5000, "ping timeout in milliseconds")
	c.PersistentFlags().StringP("format", "f", "txt", "output format: txt or png")
	c.PersistentFlags().IntP("concurrent", "n", 1, "number of nodes to test concurrently (1-5)")

	c.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		config.LoadFlags(c.PersistentFlags())
	}
	c.RunE = func(cmd *cobra.Command, args []string) error {
		return job.Run()
	}
})

var availDownloads = []string{cloudflare.Name, speedtest.Name, fast.Name}
