package main

import (
	"os"
	"fmt"
	"encoding/json"

	"github.com/bjwschaap/markdeploy/logstash"
	"github.com/urfave/cli"
	"time"
)

type jsontime time.Time

func (t jsontime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))
	return []byte(stamp), nil
}

type application struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
}

type deployment struct {
	Timestamp   jsontime      `json:"timestamp"`
	Project     string        `json:"project"`
	Application application   `json:"application"`
	Reason      string        `json:"reason"`
	Client      string        `json:"client"`
	Environment string		  `json:"environment"`
	Hosts       []string      `json:"hosts"`
}

func main() {
	app := cli.NewApp()
	app.Name = "markdeploy"
	app.Version = "0.0.3"
	app.Author = "Bastiaan Schaap"
	app.Email = "bastiaan@gntry.io"
	app.Usage = "Send deployment marker to logstash"
	app.Copyright = "\u00a92017 GNTRY"
	app.Action = run

	cli.HelpFlag = cli.BoolFlag{
		Name: "help",
		Usage: "Show usage of markdeploy",
	}
	cli.VersionFlag = cli.BoolFlag{
		Name: "V, version",
		Usage: "Only print the version and exit",
	}

	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "h, host",
			Value: "127.0.0.1",
			Usage: "IP or hostname of logstash server",
			EnvVar: "LOGSTASH_HOST",
		},
		cli.IntFlag{
			Name: "p, port",
			Value: 9000,
			Usage: "Port of the logstash tcp input",
			EnvVar: "LOGSTASH_PORT",
		},
		cli.IntFlag{
			Name: "timeout",
			Value: 5,
			Usage: "tcp timeout",
			EnvVar: "LOGSTASH_TIMEOUT",
		},
		cli.StringFlag{
			Name: "a, app",
			Usage: "name of the application/component that was deployed",
			EnvVar: "MARKDEPLOY_APPLICATION",
		},
		cli.StringFlag{
			Name: "v, appversion",
			Usage: "version of the application that was deployed",
			EnvVar: "MARKDEPLOY_APPVERSION",
		},
		cli.StringFlag{
			Name: "project",
			Usage: "optional projectname",
		},
		cli.StringFlag{
			Name: "r, reason",
			Usage: "optional reason why deployment was performed",
		},
		cli.StringFlag{
			Name: "c, client",
			Usage: "optional client/process/user that performed the deployment",
		},
		cli.BoolFlag{
			Name: "s, silent",
			Usage: "supress all output",
			EnvVar: "MARKDEPLOY_SILENT",
		},
		cli.StringFlag{
			Name: "e, env, environment",
			Usage: "the environment the application is deployed to",
			EnvVar: "MARKDEPLOY_ENV",
		},
		cli.StringSliceFlag{
			Name: "t, target",
			Usage: "one or more hosts that the application was deployed to",
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error while marking deployment: %v\n", err)
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	err := validate(c)
	if err != nil {
		return err
	}

	if !c.GlobalBool("silent") {
		fmt.Printf("Marking deployment of: %s (%s, version: %s)\n",
			c.GlobalString("app"), c.GlobalString("project"), c.GlobalString("appversion"))
		fmt.Printf("By: %s, Reason: %s\n",
			c.GlobalString("client"), c.GlobalString("reason"))
		fmt.Printf("Connecting to logstash at %s:%d (timeout: %d)\n",
			c.GlobalString("host"), c.GlobalInt("port"), c.GlobalInt("timeout"))
	}

	l := logstash.New(c.GlobalString("host"), c.GlobalInt("port"), c.GlobalInt("timeout"))
	_, err = l.Connect()
	if err != nil {
		return err
	}

	msg, err := json.Marshal(getDeployment(c))
	if err != nil {
		return err
	}

	if !c.GlobalBool("silent") {
		fmt.Printf("%s\n", string(msg))
	}

	err = l.Writeln(msg)
	if err != nil {
		return err
	}

	if !c.GlobalBool("silent") {
		fmt.Println("Deployment marked succesfully")
	}

	return nil
}

func validate(c *cli.Context) error {
	if c.GlobalString("app") == "" {
		return fmt.Errorf("can't mark a deployment without the application name")
	}
	if c.GlobalString("appversion") == "" {
		return fmt.Errorf("can't mark a deployment without the application version")
	}
	if c.GlobalString("environment") == "" {
		return fmt.Errorf("can't mark a deployment without an environment")
	}
	if len(c.GlobalStringSlice("target")) == 0 {
		return fmt.Errorf("can't mark a deployment without at least 1 host")
	}
	return nil
}

func getDeployment(c *cli.Context) deployment {
	return deployment{
		Timestamp: jsontime(time.Now()),
		Project: c.GlobalString("project"),
		Application: application{
			Name: c.GlobalString("app"),
			Version: c.GlobalString("appversion"),
		},
		Reason: c.GlobalString("reason"),
		Client: c.GlobalString("client"),
		Environment: c.GlobalString("environment"),
		Hosts: c.GlobalStringSlice("target"),
	}
}
