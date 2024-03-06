package delete

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
)

type Options struct {
	PrintFlags  *genericclioptions.PrintFlags
	PrintObject printers.ResourcePrinterFunc
	ToPrinter   func(*meta.RESTMapping, *bool, bool, bool) (printers.ResourcePrinterFunc, error)

	Namespace string
	Resource  string
	Name      string
	UID       string

	Client     client.Client
	RESTMapper meta.RESTMapper

	IOStreams *genericiooptions.IOStreams
	Factory   util.Factory
}

var (
	short = i18n.T(`Get or List resources from tekton results`)

	long = templates.LongDesc(i18n.T(`
		Delete a resource from tekton results `))

	example = templates.Examples(i18n.T(`
		# Delete a resource from a namespace
		kubectl tekton delete pr foo -n default`))
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
		Use:     "delete type name",
		Short:   short,
		Long:    long,
		Example: example,
		Args:    cobra.RangeArgs(1, 2),
		PreRunE: o.PreRun,
		RunE:    o.Run,
	}

	o.PrintFlags.AddFlags(c)

	c.Flags().StringVarP(&o.UID, "uid", "", "", "UID to select unique item")

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

	c, err := config.NewConfig()
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

	if o.Name == "" {
		return errors.New("name must be specified")
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
	if err != nil {
		return err
	}

	// TODO: Version override is not required after tekton results migration to V1 APIs
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
		return printers.WriteEscaped(o.IOStreams.Out, fmt.Sprintf("No %s with id %s:%s found", gvk.Kind, o.Namespace, o.Name))
	}

	r, ok := ul.Items[0].GetAnnotations()[annotation.Record]
	if !ok || r == "" {
		return printers.WriteEscaped(o.IOStreams.Out, "No record found")
	}
	l, ok := ul.Items[0].GetAnnotations()[annotation.Log]
	if !ok || l == "" {
		return printers.WriteEscaped(o.IOStreams.Out, "No log found")
	}
	a, ok := ul.Items[0].GetAnnotations()[annotation.Result]
	if !ok || a == "" {
		return printers.WriteEscaped(o.IOStreams.Out, "No result found")
	}

	return action.Delete(o.Client, r, l, a)
}
