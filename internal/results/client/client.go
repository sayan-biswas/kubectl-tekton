package client

import (
	"errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	resultsv1alpha2 "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/transport"
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
	ClientType          string
	Host                string
	ImpersonationConfig *transport.ImpersonationConfig
	Timeout             time.Duration
	TLSConfig           *transport.TLSConfig
	Token               string
}

func NewClient(c *Config) (Client, error) {
	c.SetDefault()

	switch c.ClientType {
	case GRPC:
		return NewGRPCClient(c)
	case REST:
		return NewRESTClient(c)
	default:
		return nil, errors.New("invalid client type")
	}
}

func (c *Config) SetDefault() {
	if c.ClientType == "" {
		c.ClientType = REST
	}

	if c.Timeout == 0 {
		c.Timeout = time.Minute
	}
}

func Status(err error) int {
	var HTTPStatusError *runtime.HTTPStatusError
	if errors.As(err, &HTTPStatusError) {
		return HTTPStatusError.HTTPStatus
	}
	return runtime.HTTPStatusFromCode(status.Code(err))
}
