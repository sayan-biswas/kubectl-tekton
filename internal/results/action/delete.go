package action

import (
	"context"
	"encoding/json"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/client"
	"github.com/tektoncd/results/pkg/watcher/reconciler/annotation"
	results "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

func Delete(c client.Client, o *Options) error {
	// Delete child resources recursively
	if a, ok := o.Annotations[annotation.Result]; ok {
		o := &Options{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{UID: o.UID},
				},
			},
		}
		for nextPage := true; nextPage; {
			lrr, err := c.ListRecords(context.Background(), &results.ListRecordsRequest{
				Parent:    a,
				Filter:    o.filter(),
				PageToken: o.ListOptions.Continue,
			})
			if err != nil {
				return err
			}
			for _, record := range lrr.Records {
				m := new(struct {
					metav1.ObjectMeta `json:"metadata,omitempty"`
				})
				err := json.Unmarshal(record.Data.Value, m)
				if err != nil {
					return err
				}
				if err := Delete(c, &Options{
					ObjectMeta: metav1.ObjectMeta{
						UID:         m.UID,
						Annotations: m.Annotations,
					},
				}); err != nil {
					return err
				}
			}
			if nextPage = lrr.NextPageToken != ""; nextPage {
				o.ListOptions.Continue = lrr.NextPageToken
			}
		}
	}

	//Delete record entries
	if a, ok := o.Annotations[annotation.Record]; ok {
		if _, err := c.DeleteRecord(context.Background(), &results.DeleteRecordRequest{
			Name: a,
		}); err != nil && client.Status(err) != http.StatusNotFound {
			return err
		}
	}

	//Delete result entries if all child records are deleted
	if a, ok := o.Annotations[annotation.Result]; ok {
		if lrr, err := c.ListRecords(context.Background(), &results.ListRecordsRequest{
			Parent: a,
		}); err != nil {
			return err
		} else if len(lrr.Records) != 0 {
			return nil
		}
		if _, err := c.DeleteResult(context.Background(), &results.DeleteResultRequest{
			Name: a,
		}); err != nil && client.Status(err) != http.StatusNotFound {
			return err
		}
	}

	return nil
}
