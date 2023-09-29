package get

import (
	"errors"
	"github.com/sayan-biswas/kubectl-tekton/internal/printer"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/action"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/client"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/config"
	"github.com/spf13/cobra"
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

type getOptions struct {
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
	getLong = templates.LongDesc(i18n.T(`
		Get or List resources from tekton results server .`))

	getExample = templates.Examples(`
		# Get resources from tekton results server
		kubectl tekton get pr -n default

		# Get resources by specifying name
		kubectl tekton get pr test-pr -n default`)
)

func Command(s *genericiooptions.IOStreams, f *genericclioptions.ConfigFlags) *cobra.Command {
	o := &getOptions{
		PrintFlags: genericclioptions.
			NewPrintFlags("").
			WithTypeSetter(scheme.Scheme).
			WithDefaultOutput("yaml"),
		IOStreams:   s,
		ConfigFlags: f,
	}

	c := &cobra.Command{
		Use:     "get",
		Short:   i18n.T("Get/List resources from tekton results"),
		Long:    getLong,
		Example: getExample,
		Args:    cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(c)

	c.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "Namespace to use")
	c.Flags().Int32VarP(&o.Limit, "limit", "l", 10, "Limit number or resource")
	c.Flags().StringVarP(&o.UID, "uid", "u", "", "Specify UID to select unique item")

	return c
}

// Complete completes the required command-line options
func (o *getOptions) Complete(args []string) (err error) {
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
	case 1:
		o.Resource = args[0]
	case 2:
		o.Resource = args[0]
		o.Name = args[1]
	default:
		return errors.New("invalid arguments, should of type RESOURCE NAME")
	}

	return nil
}

// Validate makes sure that provided values for command-line options are valid
func (o *getOptions) Validate() error {
	if o.Namespace == "" {
		return errors.New("namespace must be specified")
	}
	if o.PrintFlags.OutputFlagSpecified() && o.Name == "" {
		return errors.New("resource name is required to print resource definition")
	}

	if o.Limit < 5 || o.Limit > 100 {
		return errors.New("limit should be between 5 and 100")
	}
	return nil
}

// Run performs the execution of 'config view' sub command
func (o *getOptions) Run() error {
	gvr, _, err := explain.SplitAndParseResourceRequest(o.Resource, o.RESTMapper)
	if err != nil {
		return err
	}

	gvk, err := o.RESTMapper.KindFor(gvr)
	if err != nil {
		return err
	}

	// TODO: remove after tekton results migration to V1 APIs
	gvk.Version = "v1beta1"

	v, k := gvk.ToAPIVersionAndKind()

	opts := &action.Options{
		ListOptions: metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       k,
				APIVersion: v,
			},
			Limit: int64(o.Limit),
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

	if o.PrintFlags.OutputFlagSpecified() && len(ul.Items) > 0 {
		return o.PrintObject(ul.Items[0].DeepCopyObject(), o.IOStreams.Out)
	}

	return printer.PrintList(o.IOStreams.Out, ul)
}
