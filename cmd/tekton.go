package main

import (
	"github.com/sayan-biswas/kubectl-tekton/internal/cmd"
	"github.com/spf13/pflag"
	"k8s.io/kubectl/pkg/cmd/util"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-tekton", pflag.ExitOnError)
	pflag.CommandLine = flags

	c := cmd.Command()
	if err := c.Execute(); err != nil {
		util.CheckErr(err)
	}
}
