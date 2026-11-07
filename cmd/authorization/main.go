package main

import "github.com/notopia-uit/notopia/pkg/common/cli"

func main() {
	cli.Execute(cli.NewRoot(ServiceName, "NOTOPIA_AUTHORIZATION_LOG_LEVEL", cli.NewServeCommand(InitializeServer), cli.NewSeedCommand(InitializeSeeder)))
}
