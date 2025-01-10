package logs

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
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/explain"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"strings"
)

type Options struct {
	PrintFlags  *genericclioptions.PrintFlags
	PrintObject printers.ResourcePrinterFunc
	ToPrinter   func(*meta.RESTMapping, *bool, bool, bool) (printers.ResourcePrinterFunc, error)

	Namespace string
	Resource  string
	Name      string
	UID       string
	Limit     int32

	Client     client.Client
	RESTMapper meta.RESTMapper

	IOStreams *genericiooptions.IOStreams
	Factory   util.Factory
}

var (
	short = i18n.T(`Get resource logs`)

	long = templates.LongDesc(i18n.T(`
		Display executions logs for resource from tekton results storage`))

	example = templates.Examples(i18n.T(`
		# Get logs from tekton results storage
		kubectl tekton logs tr test 
		kubectl tekton logs pr test

		# Get logs for a particular run using UID
		kubectl tekton logs tr test --uid f27a6d83-21d3-4256-a8f0-0875b123895f`))
)

func Command(s *genericiooptions.IOStreams, f util.Factory) *cobra.Command {
	o := &Options{
		PrintFlags: genericclioptions.
			NewPrintFlags("").
			WithTypeSetter(scheme.Scheme).
			WithDefaultOutput("yaml"),
		IOStreams: s,
		Factory:   f,
	}

	c := &cobra.Command{
		Use:     "logs [type] [name]",
		Aliases: []string{"log"},
		Short:   short,
		Long:    long,
		Example: example,
		Args:    cobra.RangeArgs(1, 2),
		PreRunE: o.PreRun,
		RunE:    o.Run,
	}

	o.PrintFlags.AddFlags(c)
	c.Flags().StringVarP(&o.UID, "uid", "", "", "UID")

	return c
}

// PreRun completes the required command-line options
func (o *Options) PreRun(_ *cobra.Command, args []string) (err error) {
	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.PrintObject = printer.PrintObj

	o.Namespace, _, err = o.Factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.RESTMapper, err = o.Factory.ToRESTMapper()
	if err != nil {
		return err
	}

	c, err := config.NewConfig(o.Factory)
	if err != nil {
		return err
	}

	o.Client, err = client.NewClient(c.Get())
	if err != nil {
		return err
	}

	o.Resource = args[0]
	if len(args) > 1 {
		o.Name = args[1]
	}

	if o.Namespace == "" {
		return errors.New("namespace must be specified")
	}

	return nil
}

// Run performs the execution of 'config view' sub command
func (o *Options) Run(_ *cobra.Command, _ []string) error {
	gvr, _, err := explain.SplitAndParseResourceRequest(o.Resource, o.RESTMapper)
	if err != nil {
		return err
	}

	gvk, err := o.RESTMapper.KindFor(gvr)

	// TODO: remove after tekton results migration to V1 APIs
	//gvk.Version = "v1beta1"

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

	switch len(ul.Items) {
	default:
		return printers.WriteEscaped(o.IOStreams.Out,
			fmt.Sprintf("Multiple %s found, narrow down with --uid flag.", gvk.Kind))
	case 0:
		return printers.WriteEscaped(o.IOStreams.Out, fmt.Sprintf("No %s found", gvk.Kind))
	case 1:
		break
	}

	a, ok := ul.Items[0].GetAnnotations()[annotation.Record]
	if !ok || a == "" {
		return printers.WriteEscaped(o.IOStreams.Out, "No logs found")
	}
	// with v1alpha3 API, end point has changed from records to logs
	a = strings.Replace(a, "records", "logs", -1)
	log, err := action.Log(o.Client, &action.Options{
		ObjectMeta: metav1.ObjectMeta{
			Name: a,
		},
	})
	if err != nil {
		return err
	}

	_, err = o.IOStreams.Out.Write(log)

	return err
}
