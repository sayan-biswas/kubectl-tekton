package cmd

import (
	"github.com/sayan-biswas/kubectl-tekton/internal/cmd/config"
	"github.com/sayan-biswas/kubectl-tekton/internal/cmd/get"
	"github.com/sayan-biswas/kubectl-tekton/internal/cmd/logs"
	"github.com/sayan-biswas/kubectl-tekton/internal/cmd/version"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/completion"
	"os"
)

// Command is root/parent command to execute.
func Command() *cobra.Command {
	c := &cobra.Command{
		Use:           "tekton",
		Aliases:       []string{"tkn"},
		Short:         "Kubectl plugin for tekton",
		Long:          "Kubectl plugin for tekton",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	ios := &genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	cf := genericclioptions.NewConfigFlags(true).
		WithDeprecatedPasswordFlag().
		WithDiscoveryBurst(300).
		WithDiscoveryQPS(50.0)

	cf.AddFlags(c.PersistentFlags())

	f := util.NewFactory(util.NewMatchVersionFlags(cf))

	completion.SetFactoryForCompletion(f)
	registerFlagCompletionFunc(c, f)

	c.AddCommand(
		config.Command(ios, f),
		get.Command(ios, f),
		logs.Command(ios, f),
		version.Command(ios),
	)

	return c
}

func registerFlagCompletionFunc(cmd *cobra.Command, f util.Factory) {
	util.CheckErr(cmd.RegisterFlagCompletionFunc(
		"namespace",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completion.CompGetResource(f, "namespace", toComplete), cobra.ShellCompDirectiveNoFileComp
		}))
	util.CheckErr(cmd.RegisterFlagCompletionFunc(
		"context",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completion.ListContextsInConfig(toComplete), cobra.ShellCompDirectiveNoFileComp
		}))
	util.CheckErr(cmd.RegisterFlagCompletionFunc(
		"cluster",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completion.ListClustersInConfig(toComplete), cobra.ShellCompDirectiveNoFileComp
		}))
	util.CheckErr(cmd.RegisterFlagCompletionFunc(
		"user",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completion.ListUsersInConfig(toComplete), cobra.ShellCompDirectiveNoFileComp
		}))
}
