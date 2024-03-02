package main

import (
	"context"
	"net/http"

	"github.com/lichuan0620/taliban/pkg/config"
	"github.com/lichuan0620/taliban/pkg/metrics"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configPath    string
	listenAddress string
)

var command = cobra.Command{
	Use:   "taliban",
	Short: "An invasive telemetry benchmark tool",
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := config.LoadFile(configPath)
		if err != nil {
			return errors.Wrap(err, "load config")
		}
		background := context.Background()
		mux := http.NewServeMux()
		for i := range cfg.Factories {
			logEntry := log.WithFields(log.Fields{
				"path":    cfg.Factories[i].ExpositionPath,
				"format":  cfg.Factories[i].ExpositionFormat,
				"vectors": len(cfg.Factories[i].Vectors),
			})
			logEntry.Infof("constructing metric factory (%d/%d)", i, len(cfg.Factories))
			factory, err := metrics.NewFactory(&cfg.Factories[i])
			if err != nil {
				return errors.Wrap(err, "build metrics factory")
			}
			mux.Handle(cfg.Factories[i].ExpositionPath, factory.Handler())
			go factory.Run(background.Done())
			logEntry.Info("metric factory running")
		}
		log.WithField("address", listenAddress).Info("start listening for requests")
		return http.ListenAndServe(listenAddress, mux)
	},
}

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetReportCaller(true)
	flags := command.PersistentFlags()
	flags.StringVarP(&configPath,
		"config_path", "p", "taliban.yaml",
		"Path to the config file",
	)
	flags.StringVar(&listenAddress,
		"listen_address", "0.0.0.0:8080",
		"Address on which to listen for HTTP requests",
	)
}

func main() {
	if err := command.Execute(); err != nil {
		log.WithError(err).Fatal("execution error")
	}
}
