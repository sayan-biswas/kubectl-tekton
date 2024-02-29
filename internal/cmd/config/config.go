package config

import (
	"github.com/sayan-biswas/kubectl-tekton/internal/cmd/config/results"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/config"
	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/formatted"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/cmd/util"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// Options holds the command-line options for 'config View' sub command
type Options struct {
	PrintFlags  *genericclioptions.PrintFlags
	PrinterFunc printers.ResourcePrinterFunc
	IOStreams   *genericiooptions.IOStreams
	Config      config.Config

	View  bool
	Reset bool
}

func Command(s *genericiooptions.IOStreams, f util.Factory) *cobra.Command {
	//pathOptions := clientcmd.NewDefaultPathOptions()
	c := &cobra.Command{
		Use:               "config",
		Aliases:           []string{"config"},
		Short:             "Configure extensions",
		Long:              "Configure extensions",
		Example:           "tekton config results",
		Args:              cobra.NoArgs,
		ValidArgsFunction: formatted.ParentCompletion,
		Run:               cmdutil.DefaultSubCommandRun(s.ErrOut),
	}

	//c.PersistentFlags().StringVar(&pathOptions.LoadingRules.ExplicitPath, pathOptions.ExplicitFileFlag, pathOptions.LoadingRules.ExplicitPath, "use a particular kubeconfig file")

	c.AddCommand(results.Command(s, f))

	return c
}

// Complete completes the required command-line options
func (o *Options) Complete(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmdutil.UsageErrorf(cmd, "unexpected arguments: %v", args)
	}

	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.PrinterFunc = printer.PrintObj

	return nil
}
