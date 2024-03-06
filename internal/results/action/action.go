package action

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/client"
	results "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Delete(c client.Client, recordID, logID, resultID string) error {
	_, rerr := c.DeleteRecord(context.Background(), &results.DeleteRecordRequest{Name: recordID})
	if rerr != nil {
		return rerr
	}
	_, lerr := c.DeleteLog(context.Background(), &results.DeleteLogRequest{Name: logID})
	if lerr != nil {
		return lerr
	}
	_, aerr := c.DeleteResult(context.Background(), &results.DeleteResultRequest{Name: resultID})
	if aerr != nil {
		return aerr
	}
	return nil
}

func List(c client.Client, o *Options) (*unstructured.UnstructuredList, error) {
	err := o.validate()
	if err != nil {
		return nil, err
	}

	rl, err := c.ListRecords(context.Background(), &results.ListRecordsRequest{
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
			"nextPageToken": rl.NextPageToken,
		},
	}

	for _, r := range rl.Records {
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

func Log(c client.Client, o *Options) ([]byte, error) {
	lc, err := c.GetLog(context.Background(), &results.GetLogRequest{
		Name: o.Name,
	})
	if err != nil {
		return nil, err
	}
	l, err := lc.Recv()
	if err != nil {
		return nil, err
	}
	return l.GetData(), nil
}
