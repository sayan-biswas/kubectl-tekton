package version

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// TODO: remove this hard coding.
const clientVersion = "v0.1.2"
const serverVersion = "v0.13.2"

var (
	short = i18n.T("Print the client and server version information")

	long = templates.LongDesc(i18n.T(`
		Print the client and server version information for the current context.`))

	example = templates.Examples(i18n.T(`
		# Print the client and server versions for the current context
		kubectl version`))
)

// Options is a struct to support version command
type Options struct {
	ClientOnly bool
	Output     string
	IOStreams  *genericiooptions.IOStreams
}

// Command returns a cobra command for fetching versions
func Command(s *genericiooptions.IOStreams) *cobra.Command {
	o := &Options{
		IOStreams: s,
	}
	cmd := &cobra.Command{
		Use:     "version",
		Short:   short,
		Long:    long,
		Example: example,
		Args:    cobra.NoArgs,
		PreRunE: o.PreRun,
		RunE:    o.Run,
	}
	//cmd.Flags().BoolVar(&o.ClientOnly, "client", o.ClientOnly, "If true, shows client version only (no server required).")
	//cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "One of 'yaml' or 'json'.")
	return cmd
}

// PreRun completes all the required options
func (o *Options) PreRun(_ *cobra.Command, _ []string) error {
	return nil
}

// Run executes version command
func (o *Options) Run(_ *cobra.Command, _ []string) error {
	cv := ClientVersion()
	sv := ServerVersion()

	fmt.Fprintf(o.IOStreams.Out, "Client Version: %s\n", cv.GitVersion)
	fmt.Fprintf(o.IOStreams.Out, "Server Version: %s\n", sv.GitVersion)

	return nil
}

func ClientVersion() *version.Info {
	return &version.Info{
		GitVersion: clientVersion,
	}
}

func ServerVersion() *version.Info {
	return &version.Info{
		GitVersion: serverVersion,
	}
}
