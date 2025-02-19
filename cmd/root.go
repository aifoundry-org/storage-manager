package cmd

import (
	"fmt"
	"strings"

	"github.com/aifoundry-org/storage-manager/pkg/cache/ocidir"
	"github.com/aifoundry-org/storage-manager/pkg/server"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type subCommand func() (*cobra.Command, error)

var subCommands = []subCommand{}

func rootCmd() (*cobra.Command, error) {
	var (
		v      *viper.Viper
		cmd    *cobra.Command
		logger = log.New()
	)
	cmd = &cobra.Command{
		Use:   "storage-manager",
		Short: "storage-manager",
		Long: `Ainekko storage manager, receives requests to ensure aimages, models and components
are available at a local path for use by the Ainekko system. Exposed API allows for checking status, requesting
new components and deleting existing ones.

Run with --help for more information.
		`,
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			bindFlags(cmd, v)
			logLevel := v.GetInt("verbose")
			switch logLevel {
			case 0:
				logger.SetLevel(log.InfoLevel)
			case 1:
				logger.SetLevel(log.DebugLevel)
			case 2:
				logger.SetLevel(log.TraceLevel)
			}

			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			addr := v.GetString("address")
			logger.Infof("Starting server on %s", addr)
			cacheDir := v.GetString("cache-dir")
			logger.Infof("Cache directory is %s", cacheDir)

			// get a reference to the cache
			cache, err := ocidir.New(cacheDir)
			if err != nil {
				return err
			}

			// Start the server
			srv := server.New(addr, cache, logger)
			if err := srv.Start(); err != nil {
				return err
			}
			log.Info("exiting")
			return nil
		},
	}

	v = viper.New()
	v.SetEnvPrefix("storage-manager")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// server hostname via CLI or env var
	pflags := cmd.PersistentFlags()
	pflags.String("address", "localhost:8050", "address and port for listening for API requests")

	// debug via CLI or env var or default
	pflags.IntP("verbose", "v", 0, "set log level, 0 is info, 1 is debug, 2 is trace")

	// which mode we are running in
	pflags.String("cache-dir", "/var/lib/nekko/cache", "directory to store cached files")

	for _, subCmd := range subCommands {
		if sc, err := subCmd(); err != nil {
			return nil, err
		} else {
			cmd.AddCommand(sc)
		}
	}

	return cmd, nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name
		_ = v.BindPFlag(configName, f)
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

// Execute primary function for cobra
func Execute() {
	rootCmd, err := rootCmd()
	if err != nil {
		log.Fatal(err)
	}
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
