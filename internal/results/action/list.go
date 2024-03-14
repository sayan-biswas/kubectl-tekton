package action

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/client"
	results "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func List(c client.Client, o *Options) (*unstructured.UnstructuredList, error) {
	err := o.validate()
	if err != nil {
		return nil, err
	}

	lrr, err := c.ListRecords(context.Background(), &results.ListRecordsRequest{
		Parent:    fmt.Sprintf("%s/results/-", o.Namespace),
		Filter:    o.filter(),
		OrderBy:   "update_time desc",
		PageSize:  int32(o.ListOptions.Limit),
		PageToken: o.ListOptions.Continue,
	})
	if err != nil {
		return nil, err
	}

	ul := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"kind":          o.Kind,
			"apiVersion":    o.APIVersion,
			"nextPageToken": lrr.NextPageToken,
		},
	}

	for _, r := range lrr.Records {
		d := r.GetData().GetValue()
		u := new(unstructured.Unstructured)
		err = json.Unmarshal(d, u)
		if err != nil {
			return nil, err
		}
		ul.Items = append(ul.Items, *u)
	}

	return ul, nil
}
