package main

import "github.com/dutchcoders/identify/cmd"

func main() {
	app := cmd.New()
	app.RunAndExitOnError()
}
