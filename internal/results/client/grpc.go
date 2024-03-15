package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	resultsv1alpha2 "github.com/tektoncd/results/proto/v1alpha2/results_go_proto"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/transport"
	"net/url"
	"os"
)

type GRPCClient struct {
	resultsv1alpha2.LogsClient
	resultsv1alpha2.ResultsClient
}

// NewGRPCClient creates a new gRPC client.
func NewGRPCClient(c *Config) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	if c.Timeout == 0 {
		ctx = context.Background()
	}

	if c.URL.Port() == "" {
		switch c.URL.Scheme {
		case "https":
			c.URL.Host = c.URL.Host + ":443"
		case "http":
			c.URL.Host = c.URL.Host + ":80"
		default:
			return nil, errors.New("port or scheme missing in host")
		}
	}

	tc := insecure.NewCredentials()
	if c.URL.Scheme == "https" {
		tls, err := c.ClientTLSConfig()
		if err != nil {
			return nil, err
		}
		tc = credentials.NewTLS(tls)
	}

	cos := []grpc.CallOption{
		grpc.PerRPCCredentials(&Credentials{
			TokenSource: transport.NewCachedTokenSource(oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: c.Transport.BearerToken,
			})),
			ImpersonationConfig:   &c.Transport.Impersonate,
			SkipTransportSecurity: c.URL.Scheme != "https",
		}),
	}

	dos := []grpc.DialOption{
		grpc.WithDefaultCallOptions(cos...),
		grpc.WithTransportCredentials(tc),
	}

	clientConn, err := grpc.DialContext(ctx, c.URL.Host, dos...)
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		resultsv1alpha2.NewLogsClient(clientConn),
		resultsv1alpha2.NewResultsClient(clientConn),
	}, nil
}

func (c *Config) ClientTLSConfig() (*tls.Config, error) {
	tc := &tls.Config{
		InsecureSkipVerify: c.Transport.TLS.Insecure,
	}

	if c.Transport.TLS.CertFile != "" && c.Transport.TLS.KeyFile != "" {
		keyPair, err := tls.LoadX509KeyPair(c.Transport.TLS.CertFile, c.Transport.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not load client key pair: %v", err)
		}
		tc.Certificates = []tls.Certificate{keyPair}
	} else if c.Transport.TLS.CAFile != "" {
		cp := x509.NewCertPool()
		ca, err := os.ReadFile(c.Transport.TLS.CAFile)
		if err != nil {
			return nil, fmt.Errorf("could not read CA certificate: %v", err)
		}
		if ok := cp.AppendCertsFromPEM(ca); !ok {
			return nil, errors.New("failed to append ca certs")
		}
		tc.RootCAs = cp
	}
	return tc, nil
}

// Credentials supplies PerRPCCredentials from a Token Source and Impersonation config.
type Credentials struct {
	oauth2.TokenSource
	*transport.ImpersonationConfig
	SkipTransportSecurity bool
}

// GetRequestMetadata gets the request metadata as a map from a Custom.
func (c *Credentials) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) { //nolint:revive
	sl := credentials.PrivacyAndIntegrity
	if c.SkipTransportSecurity {
		sl = credentials.NoSecurity
	}
	ri, _ := credentials.RequestInfoFromContext(ctx)
	if err := credentials.CheckSecurityLevel(ri.AuthInfo, sl); err != nil {
		return nil, fmt.Errorf("unable to transfer TokenSource PerRPCCredentials: %v", err)
	}

	token, err := c.Token()
	if err != nil {
		return nil, err
	}

	m := map[string]string{
		"authorization": token.Type() + " " + token.AccessToken,
	}
	if c.UserName != "" {
		m[transport.ImpersonateUserHeader] = c.UserName
	}
	if c.UID != "" {
		m[transport.ImpersonateUIDHeader] = c.UID
	}
	for _, group := range c.Groups {
		m[transport.ImpersonateUIDHeader] = group
	}
	for ek, ev := range c.Extra {
		for _, v := range ev {
			m[transport.ImpersonateUserExtraHeaderPrefix+unescapeExtraKey(ek)] = v
		}
	}

	return m, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security.
func (c *Credentials) RequireTransportSecurity() bool {
	return !c.SkipTransportSecurity
}

func unescapeExtraKey(encodedKey string) string {
	key, err := url.PathUnescape(encodedKey) // Decode %-encoded bytes.
	if err != nil {
		return encodedKey // Always record extra strings, even if malformed/encoded.
	}
	return key
}
