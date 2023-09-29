package main

import (
	"github.com/sayan-biswas/kubectl-tekton/internal/cmd"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-tekton", pflag.ExitOnError)
	pflag.CommandLine = flags

	c := cmd.Command()
	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
