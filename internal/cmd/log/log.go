package log

import (
	"errors"
	"fmt"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/action"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/client"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/config"
	"github.com/spf13/cobra"
	"github.com/tektoncd/results/pkg/watcher/reconciler/annotation"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/explain"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

type logOptions struct {
	PrintFlags  *genericclioptions.PrintFlags
	PrintObject printers.ResourcePrinterFunc
	ToPrinter   func(*meta.RESTMapping, *bool, bool, bool) (printers.ResourcePrinterFunc, error)

	Namespace string
	Resource  string
	Name      string
	UID       string
	Limit     int32

	Client     client.Interface
	RESTMapper meta.RESTMapper

	IOStreams   *genericiooptions.IOStreams
	ConfigFlags *genericclioptions.ConfigFlags
}

var (
	logLong = templates.LongDesc(i18n.T(`
		Display executions logs for resource from tekton results.

		You can use --uid to select a specific resource`))

	logExample = templates.Examples(`
		# Get logs from tekton results server
		kubectl tekton log tr testrun 

		# Get logs for a particular run using UID
		kubectl tekton config view --uid f27a6d83-21d3-4256-a8f0-0875b123895f `)
)

func Command(s *genericiooptions.IOStreams, f *genericclioptions.ConfigFlags) *cobra.Command {
	o := &logOptions{
		PrintFlags: genericclioptions.
			NewPrintFlags("").
			WithTypeSetter(scheme.Scheme).
			WithDefaultOutput("yaml"),
		IOStreams:   s,
		ConfigFlags: f,
	}

	c := &cobra.Command{
		Use:     "log",
		Aliases: []string{"logs"},
		Short:   i18n.T("Display logs for a resource from tekton results"),
		Long:    logLong,
		Example: logExample,
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(c)
	c.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "Namespace")
	c.Flags().StringVarP(&o.UID, "uid", "u", "", "UID")

	return c
}

// Complete completes the required command-line options
func (o *logOptions) Complete(args []string) (err error) {
	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.PrintObject = printer.PrintObj

	o.RESTMapper, err = cmdutil.NewFactory(cmdutil.NewMatchVersionFlags(o.ConfigFlags)).ToRESTMapper()
	if err != nil {
		return err
	}

	c, err := config.NewConfig()
	if err != nil {
		return err
	}

	o.Client, err = client.NewClient(c.ClientConfig())
	if err != nil {
		return err
	}

	switch len(args) {
	case 2:
		o.Resource = args[0]
		o.Name = args[1]
	default:
		return errors.New("invalid arguments, there should be exactly 2 arguments")
	}

	return nil
}

// Validate makes sure that provided values for command-line options are valid
func (o *logOptions) Validate() error {
	if o.Namespace == "" {
		return errors.New("namespace must be specified")
	}
	return nil
}

// Run performs the execution of 'config view' sub command
func (o *logOptions) Run() error {
	gvr, _, err := explain.SplitAndParseResourceRequest(o.Resource, o.RESTMapper)
	if err != nil {
		return err
	}

	gvk, err := o.RESTMapper.KindFor(gvr)

	// TODO: remove after tekton results migration to V1 APIs
	gvk.Version = "v1beta1"

	v, k := gvk.ToAPIVersionAndKind()

	opts := &action.Options{
		ListOptions: metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       k,
				APIVersion: v,
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Name,
			Namespace: o.Namespace,
			UID:       types.UID(o.UID),
		},
	}

	ul, err := action.List(o.Client, opts)
	if err != nil {
		return err
	}

	if len(ul.Items) == 0 {
		return printers.WriteEscaped(o.IOStreams.Out, fmt.Sprintf("No %s found", gvk.Kind))
	}
	a, exists := ul.Items[0].GetAnnotations()[annotation.Log]
	if !exists || a == "" {
		return printers.WriteEscaped(o.IOStreams.Out, "No logs found")
	}
	log, err := action.Log(o.Client, &action.Options{
		ObjectMeta: metav1.ObjectMeta{
			Name: a,
		},
	})
	if err != nil {
		return err
	}

	_, err = o.IOStreams.Out.Write(log)
	if err != nil {
		return err
	}

	return nil
}
