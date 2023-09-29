package config

import (
	"github.com/sayan-biswas/kubectl-tekton/internal/results/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// ViewOptions holds the command-line options for 'config View' sub command
type resultsOptions struct {
	PrintFlags  *genericclioptions.PrintFlags
	PrinterFunc printers.ResourcePrinterFunc
	IOStreams   *genericiooptions.IOStreams
	Config      config.Config

	View  bool
	Reset bool
}

var (
	resultsLong = templates.LongDesc(i18n.T(`
		Display tekton settings from kubeconfig file.

		You can use --output jsonpath={...} to extract specific values using a jsonpath expression.`))

	resultsExample = templates.Examples(`
		# Show tekton settings
		kubectl tekton config View

		# Show tekton settings, raw certificate data, and exposed secrets
		kubectl tekton config View --raw`)
)

func Results(s *genericiooptions.IOStreams) *cobra.Command {
	o := &resultsOptions{
		PrintFlags: genericclioptions.NewPrintFlags("").WithTypeSetter(scheme.Scheme).WithDefaultOutput("yaml"),
		IOStreams:  s,
	}

	c := &cobra.Command{
		Use:     "results",
		Short:   i18n.T("Display tekton settings from kubeconfig file"),
		Long:    resultsLong,
		Example: resultsExample,
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(c)
	c.Flags().BoolVarP(&o.View, "View", "v", false, "View tekton results config")
	c.Flags().BoolVarP(&o.Reset, "Reset", "r", false, "Reset tekton results config")

	return c
}

// Complete completes the required command-line options
func (o *resultsOptions) Complete(cmd *cobra.Command, args []string) error {
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

// Validate makes sure that provided values for command-line options are valid
func (o *resultsOptions) Validate() error {
	return nil
}

// Run performs the execution of 'config View' sub command
func (o *resultsOptions) Run() (err error) {
	o.Config, err = config.NewConfig()
	if err != nil {
		return err
	}

	if o.Reset {
		err := o.Config.UpdateConfig()
		if err != nil {
			return err
		}
	}
	if o.View {
		o.PrinterFunc(o.Config.RawConfig(), o.IOStreams.Out)
	}
	return nil
}
