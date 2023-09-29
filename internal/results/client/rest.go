package client

import (
	"bytes"
	"context"
	"errors"
	v1alpha2 "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"k8s.io/client-go/transport"
	"net/http"
	"net/url"
)

const (
	basePath = "/apis/results.tekton.dev/v1alpha2/parents"
)

type RESTClient struct {
	httpClient *http.Client
	url        *url.URL
}

// NewRESTClient creates a new REST client.
func NewRESTClient(o *Options) (Client, error) {
	u, err := url.Parse(o.Host)
	if err != nil {
		return nil, err
	}

	u.Path = basePath

	rt, err := transport.New(&transport.Config{
		BearerToken: o.Token,
		TLS:         *o.TLSConfig,
		Impersonate: *o.ImpersonationConfig,
	})
	if err != nil {
		return nil, err
	}

	rc := &RESTClient{
		httpClient: &http.Client{
			Transport: rt,
			Timeout:   o.Timeout,
		},
		url: u,
	}

	return rc, nil
}

// TODO: Get these methods from a generated client

// GetResult makes request to get result
func (c *RESTClient) GetResult(ctx context.Context, in *v1alpha2.GetResultRequest, _ ...grpc.CallOption) (*v1alpha2.Result, error) {
	out := &v1alpha2.Result{}
	b, err := c.send(ctx, http.MethodGet, []string{in.Name}, in)
	if err != nil {
		return nil, err
	}
	return out, protojson.Unmarshal(b, out)
}

// ListResults makes request and get result list
func (c *RESTClient) ListResults(ctx context.Context, in *v1alpha2.ListResultsRequest, _ ...grpc.CallOption) (*v1alpha2.ListResultsResponse, error) {
	out := &v1alpha2.ListResultsResponse{}
	b, err := c.send(ctx, http.MethodGet, []string{in.Parent, "results"}, in)
	if err != nil {
		return nil, err
	}
	return out, protojson.Unmarshal(b, out)
}

// DeleteResult makes request to delete result
func (c *RESTClient) DeleteResult(ctx context.Context, in *v1alpha2.DeleteResultRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	out := &emptypb.Empty{}
	b, err := c.send(ctx, http.MethodDelete, []string{in.Name}, in)
	if err != nil {
		return nil, err
	}
	return out, protojson.Unmarshal(b, out)
}

func (c *RESTClient) CreateResult(_ context.Context, _ *v1alpha2.CreateResultRequest, _ ...grpc.CallOption) (*v1alpha2.Result, error) {
	//TODO implement me
	panic("not implemented")
}

func (c *RESTClient) UpdateResult(_ context.Context, _ *v1alpha2.UpdateResultRequest, _ ...grpc.CallOption) (*v1alpha2.Result, error) {
	//TODO: implement
	panic("not implemented")
}

// GetRecord makes request to get record
func (c *RESTClient) GetRecord(ctx context.Context, in *v1alpha2.GetRecordRequest, _ ...grpc.CallOption) (*v1alpha2.Record, error) {
	out := &v1alpha2.Record{}
	b, err := c.send(ctx, http.MethodGet, []string{in.Name}, in)
	if err != nil {
		return nil, err
	}
	return out, protojson.Unmarshal(b, out)
}

// ListRecords makes request to get record list
func (c *RESTClient) ListRecords(ctx context.Context, in *v1alpha2.ListRecordsRequest, _ ...grpc.CallOption) (*v1alpha2.ListRecordsResponse, error) {
	out := &v1alpha2.ListRecordsResponse{}
	b, err := c.send(ctx, http.MethodGet, []string{in.Parent, "records"}, in)
	if err != nil {
		return nil, err
	}
	return out, protojson.Unmarshal(b, out)
}

// DeleteRecord makes request to delete record
func (c *RESTClient) DeleteRecord(ctx context.Context, in *v1alpha2.DeleteRecordRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	out := &emptypb.Empty{}
	b, err := c.send(ctx, http.MethodDelete, []string{in.Name}, in)
	if err != nil {
		return nil, err
	}
	return out, protojson.Unmarshal(b, out)
}

func (c *RESTClient) CreateRecord(_ context.Context, _ *v1alpha2.CreateRecordRequest, _ ...grpc.CallOption) (*v1alpha2.Record, error) {
	//TODO: implement
	panic("not implemented")
}

func (c *RESTClient) UpdateRecord(_ context.Context, _ *v1alpha2.UpdateRecordRequest, _ ...grpc.CallOption) (*v1alpha2.Record, error) {
	//TODO: implement
	panic("not implemented")
}

type logsGetLogClient struct {
	log *v1alpha2.Log
	grpc.ClientStream
}

func (c logsGetLogClient) Recv() (*v1alpha2.Log, error) {
	return c.log, nil
}

func (c *RESTClient) GetLog(ctx context.Context, in *v1alpha2.GetLogRequest, _ ...grpc.CallOption) (v1alpha2.Logs_GetLogClient, error) {
	b, err := c.send(ctx, http.MethodGet, []string{in.Name}, in)
	if err != nil {
		return nil, err
	}
	out := logsGetLogClient{
		log: &v1alpha2.Log{
			Data: b,
		},
	}
	return out, nil
}

func (c *RESTClient) ListLogs(ctx context.Context, in *v1alpha2.ListRecordsRequest, _ ...grpc.CallOption) (*v1alpha2.ListRecordsResponse, error) {
	out := &v1alpha2.ListRecordsResponse{}
	b, err := c.send(ctx, http.MethodGet, []string{in.Parent, "records"}, in)
	if err != nil {
		return nil, err
	}
	return out, protojson.Unmarshal(b, out)
}

func (c *RESTClient) DeleteLog(ctx context.Context, in *v1alpha2.DeleteLogRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	out := &emptypb.Empty{}
	b, err := c.send(ctx, http.MethodDelete, []string{in.Name}, in)
	if err != nil {
		return nil, err
	}
	return out, protojson.Unmarshal(b, out)
}

func (c *RESTClient) UpdateLog(_ context.Context, _ ...grpc.CallOption) (v1alpha2.Logs_UpdateLogClient, error) {
	panic("not implemented")
}

func (c *RESTClient) send(ctx context.Context, method string, values []string, in proto.Message) ([]byte, error) {
	u := c.url.JoinPath(values...)
	q := u.Query()

	in.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.JSONName() == "parent" || !fd.HasJSONName() || fd.Kind() == protoreflect.BytesKind {
			return true
		}
		q.Set(fd.JSONName(), v.String())
		return true
	})
	u.RawQuery = q.Encode()

	var body io.Reader
	if method == http.MethodPost || method == http.MethodPut {
		b, err := protojson.Marshal(in)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(http.StatusText(res.StatusCode))
	}

	defer res.Body.Close()
	return io.ReadAll(res.Body)
}
