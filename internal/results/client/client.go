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

type Interface interface {
	resultsv1alpha2.LogsClient
	resultsv1alpha2.ResultsClient
}

type Options struct {
	ClientType          Type
	TLSConfig           *transport.TLSConfig
	ImpersonationConfig *transport.ImpersonationConfig
	Timeout             time.Duration
	Token               string
	Host                string
}

func NewClient(opts *Options) (Interface, error) {
	switch opts.ClientType {
	case GRPC:
		gc := &grpcClient{}
		return gc.create(opts)
	case REST:
		rc := &restClient{}
		return rc.create(opts)
	default:
		return nil, errors.New("invalid client type")
	}
}
