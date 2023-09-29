package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	v1alpha2 "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"k8s.io/client-go/transport"
	"net/http"
	"net/url"
)

const (
	RESTContentType string = "application/json"
)

const (
	listResultsPath   = "/apis/results.tekton.dev/v1alpha2/parents/%s/results"
	getResultsPath    = "/apis/results.tekton.dev/v1alpha2/parents/%s"
	deleteResultsPath = "/apis/results.tekton.dev/v1alpha2/parents/%s"
	listRecordsPath   = "/apis/results.tekton.dev/v1alpha2/parents/%s/records"
	getRecordPath     = "/apis/results.tekton.dev/v1alpha2/parents/%s"
	deleteRecordPath  = "/apis/results.tekton.dev/v1alpha2/parents/%s"
	listLogsPath      = "/apis/results.tekton.dev/v1alpha2/parents/%s/logs"
	getLogPath        = "/apis/results.tekton.dev/v1alpha2/parents/%s"
	deleteLogPath     = "/apis/results.tekton.dev/v1alpha2/parents/%s"
)

type restClient struct {
	httpClient *http.Client
	url        *url.URL
}

// create creates a new REST client.
func (client *restClient) create(opts *Options) (Interface, error) {
	u, err := url.Parse(opts.Host)
	if err != nil {
		return nil, err
	}

	rt, err := transport.New(&transport.Config{
		BearerToken: opts.Token,
		TLS:         *opts.TLSConfig,
		Impersonate: *opts.ImpersonationConfig,
	})
	if err != nil {
		return nil, err
	}

	rc := &restClient{
		httpClient: &http.Client{
			Transport: rt,
			Timeout:   opts.Timeout,
		},
		url: u,
	}

	return rc, nil
}

// TODO: Get these methods from a generated client

// GetResult makes request to get result
func (client *restClient) GetResult(ctx context.Context, in *v1alpha2.GetResultRequest, _ ...grpc.CallOption) (*v1alpha2.Result, error) {
	out := &v1alpha2.Result{}
	return out, client.Send(ctx, http.MethodGet, fmt.Sprintf(getResultsPath, in.Name), in, out)
}

// ListResults makes request and get result list
func (client *restClient) ListResults(ctx context.Context, in *v1alpha2.ListResultsRequest, _ ...grpc.CallOption) (*v1alpha2.ListResultsResponse, error) {
	out := &v1alpha2.ListResultsResponse{}
	return out, client.Send(ctx, http.MethodGet, fmt.Sprintf(listResultsPath, in.Parent), in, out)
}

// DeleteResult makes request to delete result
func (client *restClient) DeleteResult(ctx context.Context, in *v1alpha2.DeleteResultRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	out := &emptypb.Empty{}
	return &emptypb.Empty{}, client.Send(ctx, http.MethodDelete, fmt.Sprintf(deleteResultsPath, in.Name), in, out)
}

func (client *restClient) CreateResult(_ context.Context, _ *v1alpha2.CreateResultRequest, _ ...grpc.CallOption) (*v1alpha2.Result, error) {
	//TODO implement me
	panic("not implemented")
}

func (client *restClient) UpdateResult(_ context.Context, _ *v1alpha2.UpdateResultRequest, _ ...grpc.CallOption) (*v1alpha2.Result, error) {
	//TODO implement me
	panic("not implemented")
}

// GetRecord makes request to get record
func (client *restClient) GetRecord(ctx context.Context, in *v1alpha2.GetRecordRequest, _ ...grpc.CallOption) (*v1alpha2.Record, error) {
	out := &v1alpha2.Record{}
	return out, client.Send(ctx, http.MethodGet, fmt.Sprintf(getRecordPath, in.Name), in, out)
}

// ListRecords makes request to get record list
func (client *restClient) ListRecords(ctx context.Context, in *v1alpha2.ListRecordsRequest, _ ...grpc.CallOption) (*v1alpha2.ListRecordsResponse, error) {
	out := &v1alpha2.ListRecordsResponse{}
	return out, client.Send(ctx, http.MethodGet, fmt.Sprintf(listRecordsPath, in.Parent), in, out)
}

// DeleteRecord makes request to delete record
func (client *restClient) DeleteRecord(ctx context.Context, in *v1alpha2.DeleteRecordRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	out := &emptypb.Empty{}
	return &emptypb.Empty{}, client.Send(ctx, http.MethodDelete, fmt.Sprintf(deleteRecordPath, in.Name), in, out)
}

func (client *restClient) CreateRecord(_ context.Context, _ *v1alpha2.CreateRecordRequest, _ ...grpc.CallOption) (*v1alpha2.Record, error) {
	//TODO implement me
	panic("not implemented")
}

func (client *restClient) UpdateRecord(_ context.Context, _ *v1alpha2.UpdateRecordRequest, _ ...grpc.CallOption) (*v1alpha2.Record, error) {
	//TODO implement me
	panic("not implemented")
}

func (client *restClient) GetLog(ctx context.Context, in *v1alpha2.GetLogRequest, _ ...grpc.CallOption) (v1alpha2.Logs_GetLogClient, error) {
	out := &v1alpha2.Log{}
	return nil, client.Send(ctx, http.MethodGet, fmt.Sprintf(getLogPath, in.Name), in, out)
}

func (client *restClient) ListLogs(ctx context.Context, in *v1alpha2.ListRecordsRequest, _ ...grpc.CallOption) (*v1alpha2.ListRecordsResponse, error) {
	out := &v1alpha2.ListRecordsResponse{}
	return out, client.Send(ctx, http.MethodGet, fmt.Sprintf(listLogsPath, in.Parent), in, out)
}

func (client *restClient) DeleteLog(ctx context.Context, in *v1alpha2.DeleteLogRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	out := &emptypb.Empty{}
	return &emptypb.Empty{}, client.Send(ctx, http.MethodDelete, fmt.Sprintf(deleteLogPath, in.Name), in, out)
}

func (client *restClient) UpdateLog(_ context.Context, _ ...grpc.CallOption) (v1alpha2.Logs_UpdateLogClient, error) {
	panic("not implemented")
}

func (client *restClient) Send(ctx context.Context, method, path string, in, out proto.Message) error {
	u := client.url.JoinPath(path)

	b, err := protojson.Marshal(in)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New(http.StatusText(res.StatusCode))
	}

	defer res.Body.Close()
	b, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return protojson.Unmarshal(b, out)
}
