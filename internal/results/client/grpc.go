package client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	resultsv1alpha2 "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/client-go/transport"
	"net/url"
)

type GRPCClient struct {
	resultsv1alpha2.LogsClient
	resultsv1alpha2.ResultsClient
}

// NewGRPCClient creates a new gRPC client.
func NewGRPCClient(o *Options) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), o.Timeout)
	defer cancel()

	callOptions := []grpc.CallOption{
		grpc.PerRPCCredentials(&customCredentials{
			TokenSource: transport.NewCachedTokenSource(oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: o.Token,
			})),
			ImpersonationConfig: o.ImpersonationConfig,
		}),
	}

	dialOptions := []grpc.DialOption{
		//grpc.WithBlock(),
		grpc.WithDefaultCallOptions(callOptions...),
	}

	if o.TLSConfig.Insecure {
		dialOptions = append(dialOptions,
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})),
		)
	}

	u, err := url.Parse(o.Host)
	if err != nil {
		return nil, err
	}

	if u.Port() == "" {
		switch u.Scheme {
		case "https":
			u.Host = u.Host + ":443"
		case "http":
			u.Host = u.Host + ":80"
		default:
			return nil, errors.New("port or scheme missing in host")
		}
	}

	clientConn, err := grpc.DialContext(ctx, u.Host, dialOptions...)
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		resultsv1alpha2.NewLogsClient(clientConn),
		resultsv1alpha2.NewResultsClient(clientConn),
	}, nil
}

// customCredentials supplies PerRPCCredentials from a Token Source and Impersonation config.
type customCredentials struct {
	oauth2.TokenSource
	*transport.ImpersonationConfig
}

// GetRequestMetadata gets the request metadata as a map from a Custom.
func (cc *customCredentials) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) { //nolint:revive
	ri, _ := credentials.RequestInfoFromContext(ctx)
	if err := credentials.CheckSecurityLevel(ri.AuthInfo, credentials.PrivacyAndIntegrity); err != nil {
		return nil, fmt.Errorf("unable to transfer TokenSource PerRPCCredentials: %v", err)
	}

	token, err := cc.Token()
	if err != nil {
		return nil, err
	}

	m := map[string]string{
		"authorization": token.Type() + " " + token.AccessToken,
	}
	if cc.UserName != "" {
		m[transport.ImpersonateUserHeader] = cc.UserName
	}
	if cc.UID != "" {
		m[transport.ImpersonateUIDHeader] = cc.UID
	}
	for _, group := range cc.Groups {
		m[transport.ImpersonateUIDHeader] = group
	}
	for ek, ev := range cc.Extra {
		for _, v := range ev {
			m[transport.ImpersonateUserExtraHeaderPrefix+unescapeExtraKey(ek)] = v
		}
	}

	return m, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security.
func (cc *customCredentials) RequireTransportSecurity() bool {
	return true
}

func unescapeExtraKey(encodedKey string) string {
	key, err := url.PathUnescape(encodedKey) // Decode %-encoded bytes.
	if err != nil {
		return encodedKey // Always record extra strings, even if malformed/encoded.
	}
	return key
}
