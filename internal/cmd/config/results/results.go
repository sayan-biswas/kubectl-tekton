package results

import (
	"github.com/sayan-biswas/kubectl-tekton/internal/helper"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// Options holds the command-line options configure tekton results
type Options struct {
	PrintFlags  *genericclioptions.PrintFlags
	PrinterFunc printers.ResourcePrinterFunc
	IOStreams   *genericiooptions.IOStreams
	Factory     util.Factory
	Config      config.Config

	NoPrompt bool
	Reset    bool
	View     bool
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

func Command(s *genericiooptions.IOStreams, f util.Factory) *cobra.Command {
	o := &Options{
		PrintFlags: genericclioptions.NewPrintFlags("").WithTypeSetter(scheme.Scheme).WithDefaultOutput("yaml"),
		IOStreams:  s,
		Factory:    f,
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
	c.Flags().BoolVarP(&o.NoPrompt, "no-prompt", "", false, "Use prompts")
	c.Flags().BoolVarP(&o.Reset, "reset", "", false, "Reset tekton results config")
	c.Flags().BoolVarP(&o.View, "view", "", false, "View tekton results config")

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
	o.Config, err = config.NewConfig(o.Factory)
	if err != nil {
		return
	}

	switch {
	case o.View:
		return o.PrinterFunc(o.Config.GetObject(), o.IOStreams.Out)
	case o.Reset:
		return o.Config.Reset()
	case len(args) == 0:
		return o.Config.Set(nil, !o.NoPrompt)
	case len(args) >= 0:
		return o.Config.Set(helper.ParseArgs(args), !o.NoPrompt)
	default:
		return nil
	}
}
