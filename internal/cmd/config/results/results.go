package results

import (
	"github.com/sayan-biswas/kubectl-tekton/internal/helper"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// Options holds the command-line options configure tekton results
type Options struct {
	PrintFlags  *genericclioptions.PrintFlags
	PrinterFunc printers.ResourcePrinterFunc
	IOStreams   *genericiooptions.IOStreams
	Config      config.Config

	View     bool
	Reset    bool
	Defaults bool
}

var (
	short = i18n.T(`Configure tekton results client`)

	long = templates.LongDesc(i18n.T(`
		Modify or configure tekton results client configuration.
	`))

	example = templates.Examples(`
		# Configure all parameters interactively
		kubectl tekton config results

		# Configure specific parameters interactively
		kubectl tekton config results host token

		# Configure specific parameters non-interactively
		kubectl tekton config results host="https://localhost:8080" token="test-token"
	`)
)

func Command(s *genericiooptions.IOStreams) *cobra.Command {
	o := &Options{
		PrintFlags: genericclioptions.NewPrintFlags("").WithTypeSetter(scheme.Scheme).WithDefaultOutput("yaml"),
		IOStreams:  s,
	}

	c := &cobra.Command{
		Use:     "results [key=value]",
		Short:   short,
		Long:    long,
		Example: example,
		PreRunE: o.PreRun,
		RunE:    o.Run,
	}

	o.PrintFlags.AddFlags(c)
	c.Flags().BoolVarP(&o.View, "view", "", false, "View tekton results config")
	c.Flags().BoolVarP(&o.Reset, "reset", "", false, "Reset tekton results config")
	c.Flags().BoolVarP(&o.Defaults, "defaults", "", false, "Use predetermined defaults found from the login")

	return c
}

// PreRun completes the required command-line options
func (o *Options) PreRun(_ *cobra.Command, _ []string) error {
	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.PrinterFunc = printer.PrintObj

	return nil
}

// Run performs the execution of 'config View' sub command
func (o *Options) Run(_ *cobra.Command, args []string) (err error) {
	o.Config, err = config.NewConfig()
	if err != nil {
		return
	}

	if o.Defaults {
		return o.Config.Defaults()
	}

	if len(args) > 0 {
		return o.Config.Set(helper.ParseArgs(args))
	}

	if o.Reset {
		return o.Config.Reset()
	}

	if o.View {
		err = o.PrinterFunc(o.Config.RawConfig(), o.IOStreams.Out)
		if err != nil {
			return err
		}
	}
	return nil
}
