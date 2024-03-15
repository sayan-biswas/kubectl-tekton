package client

import (
	"errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	resultsv1alpha2 "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/transport"
	"net/url"
	"time"
)

const (
	GRPC = "GRPC"
	REST = "REST"
)

type Client interface {
	resultsv1alpha2.LogsClient
	resultsv1alpha2.ResultsClient
}

type Config struct {
	ClientType string
	URL        *url.URL
	Timeout    time.Duration
	Transport  *transport.Config
}

func NewClient(config *Config) (Client, error) {
	switch config.ClientType {
	case GRPC:
		return NewGRPCClient(config)
	case REST:
		return NewRESTClient(config)
	default:
		return NewRESTClient(config)
	}
}

func Status(err error) int {
	var HTTPStatusError *runtime.HTTPStatusError
	if errors.As(err, &HTTPStatusError) {
		return HTTPStatusError.HTTPStatus
	}
	return runtime.HTTPStatusFromCode(status.Code(err))
}
