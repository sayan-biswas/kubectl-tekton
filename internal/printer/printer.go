package printer

import (
	"github.com/jonboulle/clockwork"
	"github.com/tektoncd/cli/pkg/formatted"
	"io"
	"text/tabwriter"
	"text/template"

	"k8s.io/apimachinery/pkg/runtime"
	cliopts "k8s.io/cli-runtime/pkg/genericclioptions"
)

func PrintObject(w io.Writer, o runtime.Object, f *cliopts.PrintFlags) error {
	printer, err := f.ToPrinter()
	if err != nil {
		return err
	}
	return printer.PrintObj(o, w)
}

func PrintList(w io.Writer, l *List) error {

	var data = struct {
		List          *List
		Time          clockwork.Clock
		AllNamespaces bool
		NoHeaders     bool
	}{
		List:          l,
		Time:          clockwork.NewRealClock(),
		AllNamespaces: false,
		NoHeaders:     false,
	}

	funcMap := template.FuncMap{
		"formatAge":       formatted.Age,
		"formatDuration":  formatted.Duration,
		"formatCondition": formatted.Condition,
	}

	tw := tabwriter.NewWriter(w, 0, 5, 5, ' ', tabwriter.TabIndent)
	t := template.Must(template.New("List PipelineRuns").Funcs(funcMap).Parse(listTemplate))

	err := t.Execute(tw, data)
	if err != nil {
		return err
	}

	return tw.Flush()
}
