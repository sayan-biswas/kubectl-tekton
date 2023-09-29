package client

import (
	"errors"
	resultsv1alpha2 "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"k8s.io/client-go/transport"
	"time"
)

type Type string

const (
	GRPC Type = "GRPC"
	REST Type = "REST"
)

type Client interface {
	resultsv1alpha2.LogsClient
	resultsv1alpha2.ResultsClient
}

type Options struct {
	Type                Type
	TLSConfig           *transport.TLSConfig
	ImpersonationConfig *transport.ImpersonationConfig
	Timeout             time.Duration
	Token               string
	Host                string
}

func NewClient(o *Options) (Client, error) {
	switch o.Type {
	case GRPC:
		return NewGRPCClient(o)
	case REST:
		return NewRESTClient(o)
	default:
		return nil, errors.New("invalid client type")
	}
}
