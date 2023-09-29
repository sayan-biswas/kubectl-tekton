package action

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/client"
	results "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
	"strings"
)

type Options struct {
	metav1.ListOptions
	metav1.ObjectMeta
}

func List(c client.Interface, o *Options) (*unstructured.UnstructuredList, error) {
	err := o.validate()
	if err != nil {
		return nil, err
	}

	rl, err := c.ListRecords(context.Background(), &results.ListRecordsRequest{
		Parent:   fmt.Sprintf("%s/results/-", o.Namespace),
		Filter:   o.filter(),
		PageSize: int32(o.ListOptions.Limit),
		OrderBy:  "update_time desc",
	})
	if err != nil {
		return nil, err
	}

	ul := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"kind":       o.Kind,
			"apiVersion": o.APIVersion,
			"metadata":   map[string]interface{}{},
		},
	}

	for _, r := range rl.Records {
		d := r.GetData().GetValue()
		u := new(unstructured.Unstructured)
		err = json.Unmarshal(d, u)
		if err != nil {
			return nil, err
		}
		ul.Object = u.Object
		ul.Items = append(ul.Items, *u)
	}

	return ul, nil
}

func Log(c client.Interface, o *Options) ([]byte, error) {
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

func (o *Options) validate() error {
	switch {
	case o.Namespace == "":
		o.Namespace = "-"
	}
	return nil
}

func (o *Options) filter() string {
	var filters []string

	if o.Kind != "" && o.APIVersion != "" {
		filters = append(filters, fmt.Sprintf("data_type==\"%s.%s\"", o.APIVersion, o.Kind))
	}

	// TODO: add support for other types
	v := reflect.ValueOf(o.ObjectMeta)
	for i := 0; i < v.NumField(); i++ {
		name, _, found := strings.Cut(v.Type().Field(i).Tag.Get("json"), ",")
		if !found {
			continue
		}
		value := v.Field(i).Interface()
		switch v.Type().Field(i).Type.Kind() {
		case reflect.String:
			if fmt.Sprintf("%v", value) != "" {
				filters = append(filters, fmt.Sprintf("data.metadata.%s==\"%s\"", name, value))
			}
		}
	}

	return strings.Join(filters, " && ")
}
