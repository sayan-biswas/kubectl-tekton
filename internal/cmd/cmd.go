package cmd

import (
	"github.com/sayan-biswas/kubectl-tekton/internal/cmd/config"
	"github.com/sayan-biswas/kubectl-tekton/internal/cmd/get"
	"github.com/sayan-biswas/kubectl-tekton/internal/cmd/log"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"os"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:          "tekton",
		Aliases:      []string{"tkn"},
		Short:        "Kubectl plugin for tekton",
		Long:         "Kubectl plugin for tekton",
		SilenceUsage: true,
	}

	s := &genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	f := genericclioptions.NewConfigFlags(true).
		WithDeprecatedPasswordFlag().
		WithDiscoveryBurst(300).
		WithDiscoveryQPS(50.0)

	c.AddCommand(
		config.Command(s),
		get.Command(s, f),
		log.Command(s, f),
	)

	return c
}
