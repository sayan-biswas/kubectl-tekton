package config

import (
	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/formatted"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func Command(s *genericiooptions.IOStreams) *cobra.Command {
	pathOptions := clientcmd.NewDefaultPathOptions()
	c := &cobra.Command{
		Use:               "config",
		Aliases:           []string{"config"},
		Short:             "Configure extensions",
		Long:              "Configure extensions",
		Example:           "tekton config",
		ValidArgsFunction: formatted.ParentCompletion,
		Run:               cmdutil.DefaultSubCommandRun(s.ErrOut),
	}

	c.PersistentFlags().StringVar(&pathOptions.LoadingRules.ExplicitPath, pathOptions.ExplicitFileFlag, pathOptions.LoadingRules.ExplicitPath, "use a particular kubeconfig file")

	c.AddCommand(Results(s))

	return c
}
