package get

import (
	"errors"
	"github.com/sayan-biswas/kubectl-tekton/internal/helper"
	"github.com/sayan-biswas/kubectl-tekton/internal/printer"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/action"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/client"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/explain"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"os"
)

const (
	Escape = 27
)

type Options struct {
	PrintFlags  *genericclioptions.PrintFlags
	PrintObject printers.ResourcePrinterFunc
	ToPrinter   func(*meta.RESTMapping, *bool, bool, bool) (printers.ResourcePrinterFunc, error)

	Namespace       string
	Resource        string
	Name            string
	UID             string
	Limit           int32
	Labels          string
	Annotations     string
	Finalizers      string
	OwnerReferences string
	Filter          string

	Client     client.Client
	RESTMapper meta.RESTMapper

	IOStreams *genericiooptions.IOStreams
	Factory   util.Factory
}

var (
	short = i18n.T(`Get or List resources from tekton results`)

	long = templates.LongDesc(i18n.T(`
		Get or List resources from tekton results `))

	example = templates.Examples(i18n.T(`
		# List resources from a namespace
		kubectl tekton get pr -n default

		# List limited resources from a namespace. By default only 10 resources are listed.
		kubectl tekton get pr -n default --limit 20

		# Get resources by specifying name. Partial name can also be provided.
		kubectl tekton get pr test -n default

		# Get resources by specifying UID. Partial UID can also be provided.
		kubectl tekton get pr test -n default --uid="e0e4148c-b914"

		# List resources from a namespace with selectors. All the selectors support partial value.
		kubectl tekton get pr -n default 
			--labels="app.kubernetes.io/name=test-app, app.kubernetes.io/component=database"

		# All selectors can be used together and works as AND operator.
		kubectl tekton get pr -n default 
			--labels="app.kubernetes.io/name=test-app"
			--annotations="app.io/timeout=100"

		# All selectors except OwnerReferences can work with only key or value.
		kubectl tekton get pr -n default --annotations="test" --labels="test"

		# Check if a particular annotation exists, without knowing the value.
		kubectl tekton get pr -n default --annotations="results.tekton.dev/log"

		# OwnerReferences filter can not filter by key/value pair, but the filter should still be provided as key/value.
		kubectl tekton get pr -n default --owner-references="kind=Service name=test-service"

		# Multiple owner references can be provided, but keys of each owner references should be seperated by space.
		kubectl tekton get pr -n default 
			--owner-references="kind=Service name=test-service, kind=Deployment name=test-app"

		# OwnerReferences filter can be used to find child resources.
		kubectl tekton get pr -n default --owner-references="name=parent-name"
		
		# Filter flag can be used to pass raw filter. Invalid syntax will cause error.
		kubectl tekton get pr -n default --filter="data.status.conditions[0].reason in ['Failed']"`))
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
		Use:     "get [type] [name]",
		Short:   short,
		Long:    long,
		Example: example,
		Args:    cobra.RangeArgs(1, 2),
		PreRunE: o.PreRun,
		RunE:    o.Run,
	}

	o.PrintFlags.AddFlags(c)

	c.Flags().Int32VarP(&o.Limit, "limit", "", 10, "Limit number or resource")
	c.Flags().StringVarP(&o.UID, "uid", "", "", "UID to select unique item")
	c.Flags().StringVarP(&o.Labels, "selector", "", "", "Filter items by labels")
	c.Flags().StringVarP(&o.Labels, "labels", "", "", "Filter items by labels")
	c.Flags().StringVarP(&o.Annotations, "annotations", "", "", "Filter items by annotations")
	c.Flags().StringVarP(&o.Finalizers, "finalizers", "", "", "Filter items by finalizers")
	c.Flags().StringVarP(&o.OwnerReferences, "owner-references", "", "", "Filter items by OwnerReferences")
	c.Flags().StringVarP(&o.Filter, "filter", "", "", "Use a raw filter string")

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

	if o.Namespace == "" {
		return errors.New("namespace must be specified")
	}

	if o.Limit < 5 || o.Limit > 100 {
		return errors.New("limit should be between 5 and 100")
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
		Filter: o.Filter,
		ListOptions: metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       k,
				APIVersion: v,
			},
			Limit: int64(o.Limit),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            o.Name,
			Namespace:       o.Namespace,
			UID:             types.UID(o.UID),
			Labels:          helper.ParseLabels(o.Labels),
			Annotations:     helper.ParseAnnotations(o.Annotations),
			Finalizers:      helper.ParseFinalizers(o.Finalizers),
			OwnerReferences: helper.ParseOwnerReferences(o.OwnerReferences),
		},
	}

	for nextPage := true; nextPage; {
		ul, err := action.List(o.Client, opts)
		if err != nil {
			return err
		}

		if o.PrintFlags.OutputFlagSpecified() {
			switch len(ul.Items) {
			case 0:
				break
			case 1:
				return o.PrintObject(ul.Items[0].DeepCopyObject(), o.IOStreams.Out)
			default:
				return errors.New("multiple resources found, use --uid flag to select")
			}
		}

		l := new(printer.List)
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(ul.UnstructuredContent(), l); err != nil {
			return err
		}
		if err := printer.PrintList(o.IOStreams.Out, l); err != nil {
			return err
		}
		if nextPage = l.NextPageToken != ""; nextPage {
			opts.ListOptions.Continue = l.NextPageToken
			if err := printers.WriteEscaped(o.IOStreams.Out,
				"\nNext Page: Press any key to continue, CTRL+ESC to exit!\n\n"); err != nil {
				return err
			}
			if Key(Escape) {
				break
			}
		}
	}

	return nil
}

func Key(key byte) bool {
	s, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return false
	}
	defer term.Restore(int(os.Stdin.Fd()), s)

	b := make([]byte, 1)
	_, err = os.Stdin.Read(b)
	if err != nil {
		return false
	}
	return key == b[0]
}
