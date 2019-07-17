package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/spaceuptech/space-cloud/config"
	"github.com/spaceuptech/space-cloud/utils"
	"github.com/spaceuptech/space-cloud/utils/server"
)

var essentialFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "config",
		Value: "config.yaml",
		Usage: "Load space cloud config from `FILE`",
	},
	cli.BoolFlag{
		Name:   "prod",
		Usage:  "Run space-cloud in production mode",
		EnvVar: "PROD",
	},
	cli.BoolFlag{
		Name:   "disable-metrics",
		Usage:  "Disable anonymous metric collection",
		EnvVar: "DISABLE_METRICS",
	},
	cli.BoolFlag{
		Name:   "disable-nats",
		Usage:  "Disable embedded nats server",
		EnvVar: "DISABLE_NATS",
	},
	cli.StringFlag{
		Name:   "seeds",
		Value:  "",
		Usage:  "Seed nodes to cluster with",
		EnvVar: "SEEDS",
	},
	cli.BoolFlag{
		Name:   "profiler",
		Usage:  "Enable profiler endpoints for profiling",
		EnvVar: "PROFILER",
	},
	cli.StringFlag{
		Name:   "admin-user",
		Usage:  "Set the admin user name",
		EnvVar: "ADMIN_USER",
		Value:  "",
	},
	cli.StringFlag{
		Name:   "admin-pass",
		Usage:  "Set the admin password",
		EnvVar: "ADMIN_PASS",
		Value:  "",
	},
	cli.StringFlag{
		Name:   "admin-sercret",
		Usage:  "Set the admin secret",
		EnvVar: "ADMIN_SECRET",
		Value:  "",
	},
}

func main() {
	app := cli.NewApp()
	app.Version = utils.BuildVersion
	app.Name = "space-cloud-ee"
	app.Usage = "core binary to run space cloud"

	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "runs the space cloud instance",
			Action: actionRun,
			Flags:  essentialFlags,
		},
		{
			Name:   "init",
			Usage:  "creates a config file with sensible defaults",
			Action: actionInit,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func actionRun(c *cli.Context) error {
	// Load cli flags
	configPath := c.String("config")
	isProd := c.Bool("prod")
	disableMetrics := c.Bool("disable-metrics")
	disableNats := c.Bool("disable-nats")
	seeds := c.String("seeds")
	profiler := c.Bool("profiler")

	// Flags related to the admin details
	adminUser := c.String("admin-user")
	adminPass := c.String("admin-pass")
	adminSecret := c.String("admin-secret")

	// Project and env cannot be changed once space cloud has started
	s := server.New(isProd)

	// Load the configFile from path if provided
	conf, err := config.LoadConfigFromFile(configPath)
	if err != nil {
		conf = config.GenerateEmptyConfig()
	}

	// Save the config file path for future use
	s.SetConfigFilePath(configPath)

	// Override the admin config if provided
	if adminUser != "" {
		conf.Admin.Users[0].User = adminUser
	}
	if adminPass != "" {
		conf.Admin.Users[0].Pass = adminPass
	}
	if adminSecret != "" {
		conf.Admin.Secret = adminSecret
	}

	// Configure all modules
	s.SetConfig(conf)

	// Anonymously collect usage metrics if not explicitly disabled
	if !disableMetrics {
		go s.RoutineMetrics()
	}

	// Download and host mission control
	staticPath, err := initMissionContol(utils.BuildVersion)
	if err != nil {
		return err
	}

	s.Routes(profiler, staticPath)

	// Start nats if not disabled
	if !disableNats {
		err := s.RunNatsServer(seeds, utils.PortNatsServer, utils.PortNatsCluster)
		if err != nil {
			return err
		}
	}

	return s.Start(seeds)
}

func actionInit(*cli.Context) error {
	return config.GenerateConfig("none")
}

func initMissionContol(version string) (string, error) {
	homeDir := utils.UserHomeDir()
	uiPath := homeDir + "/.space-cloud/mission-control-v" + version
	if _, err := os.Stat(uiPath); os.IsNotExist(err) {
		fmt.Println("Could not find mission control")
		if _, err := os.Stat(homeDir + "/space-cloud"); os.IsNotExist(err) {
			os.Mkdir(homeDir+"/.space-cloud", os.ModePerm)
		}
		fmt.Println("Downloading...")
		err := utils.DownloadFileFromURL("https://spaceuptech.com/downloads/mission-control/mission-control-v"+version+".zip", uiPath+".zip")
		if err != nil {
			return "", err
		}
		fmt.Println("Extracting...")
		err = utils.Unzip(uiPath+".zip", uiPath)
		if err != nil {
			return "", err
		}
		fmt.Println("Done...")
		err = os.Remove(uiPath + ".zip")
		if err != nil {
			return "", err
		}
	}
	return uiPath + "/build", nil
}
