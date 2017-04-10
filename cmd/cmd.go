package main

import (
	"fmt"

	"github.com/dutchsec/identify"
	"github.com/fatih/color"
	"github.com/minio/cli"
	"github.com/op/go-logging"
)

var Version = "0.1"
var helpTemplate = `NAME:
{{.Name}} - {{.Usage}}

DESCRIPTION:
{{.Description}}

USAGE:
{{.Name}} {{if .Flags}}[flags] {{end}}command{{if .Flags}}{{end}} [arguments...]

COMMANDS:
{{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
{{end}}{{if .Flags}}
FLAGS:
{{range .Flags}}{{.}}
{{end}}{{end}}
VERSION:
` + Version +
	`{{ "\n"}}`

var log = logging.MustGetLogger("identify/cmd")

var globalFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "application",
		Usage: "the application to identify",
		Value: "",
	},
	cli.StringFlag{
		Name:  "proxy",
		Usage: "socks5://127.0.0.1:9050",
		// Usage: "the proxy to use",
		Value: "",
	},
	cli.BoolFlag{
		Name:  "debug",
		Usage: "enable debug mode",
	},
	cli.BoolFlag{
		Name:  "no-branches",
		Usage: "don't identify branches",
	},
	cli.BoolFlag{
		Name:  "no-tags",
		Usage: "don't identify tags",
	},
	cli.BoolFlag{
		Name:  "json",
		Usage: "output json",
	},
}

type Cmd struct {
	*cli.App
}

func VersionAction(c *cli.Context) {
	fmt.Println(color.YellowString(fmt.Sprintf("identify")))
}

func main() {
	app := cli.NewApp()
	app.Name = "identify"
	app.Author = ""
	app.Usage = "DutchSec"
	app.Description = ``
	app.Flags = globalFlags
	app.CustomAppHelpTemplate = helpTemplate
	app.Commands = []cli.Command{
		{
			Name:   "version",
			Action: VersionAction,
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.Action = func(c *cli.Context) {
		fmt.Println("Identify - Identify application versions")
		fmt.Println("http://github.com/dutchsec/identify")
		fmt.Println("DutchSec [https://dutchsec.com/]")
		fmt.Println("========================================")

		options := []identify.OptionFn{}

		if args := c.Args(); len(args) == 0 {
			fmt.Println(color.RedString("[*] No target url set"))
			return
		} else if fn, err := identify.TargetURL(args[0]); err != nil {
			fmt.Println(color.RedString("[*] Could not parse target url: %s", err.Error()))
			return
		} else {
			options = append(options, fn)
		}

		if proxy := c.GlobalString("proxy"); proxy == "" {
		} else if fn, err := identify.ProxyURL(proxy); err != nil {
			fmt.Println(color.RedString("[*] Could find set proxy: %s", err.Error()))
			return
		} else {
			options = append(options, fn)
		}

		if application := c.GlobalString("application"); application == "" {
			fmt.Println(color.RedString("[*] No application set"))
			return
		} else if fn, err := identify.TargetApplication(application); err != nil {
			fmt.Println(color.RedString("[*] Could find target application: %s", err.Error()))
			return
		} else {
			options = append(options, fn)
		}

		if !c.Bool("no-branches") {
		} else if fn, err := identify.NoBranches(); err != nil {
		} else {
			options = append(options, fn)
		}

		if !c.Bool("no-tags") {
		} else if fn, err := identify.NoTags(); err != nil {
		} else {
			options = append(options, fn)
		}

		if !c.Bool("debug") {
		} else if fn, err := identify.Debug(); err != nil {
		} else {
			options = append(options, fn)
		}

		b, err := identify.New(options...)
		if err != nil {
			fmt.Println(color.RedString("[*] Could not parse target url: %s", err.Error()))
			return
		}

		if err := b.Identify(""); err != nil {
			fmt.Println(color.RedString("[*] Error identifying application: %s", err.Error()))
			return
		}
	}

	app.RunAndExitOnError()
	/*
		return &Cmd{
			App: app,
		}
	*/
}
