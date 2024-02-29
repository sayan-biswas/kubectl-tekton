package action

import (
	"context"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/client"
	results "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
)

func Log(c client.Client, o *Options) ([]byte, error) {
	glc, err := c.GetLog(context.Background(), &results.GetLogRequest{
		Name: o.Name,
	})
	if err != nil {
		return nil, err
	}
	l, err := glc.Recv()
	if err != nil {
		return nil, err
	}
	return l.GetData(), nil
}
