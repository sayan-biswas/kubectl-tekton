package config

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/AlecAivazis/survey/v2"
	v1 "github.com/openshift/api/route/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/transport"
	"k8s.io/kubectl/pkg/cmd/util"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Config interface {
	Get() *client.Config
	RawConfig() runtime.Object
	Set(data map[string]*string, prompt bool) error
	Reset() error
}

type config struct {
	ConfigAccess clientcmd.ConfigAccess
	APIConfig    *api.Config
	RESTConfig   *rest.Config
	ClientConfig *client.Config
	Extension    *Extension
}

const (
	ServiceLabel  string = "app.kubernetes.io/name=tekton-results-api"
	ExtensionName string = "tekton-results"
)

func NewConfig(factory util.Factory) (Config, error) {
	cc := factory.ToRawKubeConfigLoader()

	ca := cc.ConfigAccess()

	ac, err := cc.RawConfig()
	if err != nil {
		return nil, err
	}

	rc, err := cc.ClientConfig()
	if err != nil {
		return nil, err
	}

	c := &config{
		ConfigAccess: ca,
		APIConfig:    &ac,
		RESTConfig:   rc,
	}

	if err := c.LoadExtension(); err != nil {
		return nil, err
	}

	if c.Extension == nil {
		c.SetVersion()
		if err := c.Set(nil, true); err != nil {
			return nil, err
		}
	}

	c.LoadClientConfig()

	return c, nil
}

func (c *config) Get() *client.Config {
	return c.ClientConfig
}

func (c *config) RawConfig() runtime.Object {
	return c.Extension
}

func (c *config) Reset() error {
	c.Extension = new(Extension)
	c.SetVersion()
	return c.Persist()
}

func (c *config) LoadExtension() error {
	cc := c.APIConfig.Contexts[c.APIConfig.CurrentContext]
	if cc == nil {
		return errors.New("current context is not set in kubeconfig")
	}
	c.Extension = new(Extension)
	e := cc.Extensions[ExtensionName]
	if e != nil {
		return json.Unmarshal(e.(*runtime.Unknown).Raw, c.Extension)
	}
	c.SetVersion()
	return c.Set(nil, true)
}

func (c *config) SetVersion() {
	c.Extension.TypeMeta = runtime.TypeMeta{
		APIVersion: "v1alpha1",
		Kind:       "Client",
	}
}

func (c *config) Set(data map[string]*string, prompt bool) error {
	t := reflect.TypeOf(c.Extension).Elem()
	v := reflect.ValueOf(c.Extension).Elem()
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		name, _, _ := strings.Cut(tf.Tag.Get("json"), ",")
		if name == "" {
			continue
		}

		vf := v.Field(i)
		if _, ok := data[name]; !ok && data != nil {
			if group, ok := tf.Tag.Lookup("group"); ok {
				if _, ok := data[group]; !ok {
					continue
				}
			} else {
				continue
			}
		}

		// get data from prompt in enabled
		if prompt {
			if err := c.Prompt(name, vf, c.CallMethod(tf.Name)); err != nil {
				return err
			}
			continue
		}

		// get data from user input if provided
		if value, ok := data[name]; ok && value != nil {
			vf.SetString(*value)
			continue
		}

		// get data from suggestion methods if exists
		switch value := c.CallMethod(tf.Name).(type) {
		case string:
			vf.SetString(value)
		case []string:
			if len(value) > 0 {
				vf.SetString(value[0])
			}
		default:
			vf.SetZero()
		}
	}

	return c.Persist()
}

func (c *config) LoadClientConfig() {
	ic := transport.ImpersonationConfig(c.RESTConfig.Impersonate)
	c.ClientConfig = &client.Config{
		ClientType:          client.REST,
		Host:                c.RESTConfig.Host,
		ImpersonationConfig: &ic,
		Timeout:             c.RESTConfig.Timeout,
		Token:               c.RESTConfig.BearerToken,
		TLSConfig: &transport.TLSConfig{
			Insecure:   c.RESTConfig.Insecure,
			CAFile:     c.RESTConfig.CAFile,
			CertFile:   c.RESTConfig.CertFile,
			KeyFile:    c.RESTConfig.KeyFile,
			ServerName: c.RESTConfig.ServerName,
		},
	}

	if c.Extension.Host != "" {
		c.ClientConfig.Host = c.Extension.Host
	}

	if c.Extension.Token != "" {
		c.ClientConfig.Token = c.Extension.Token
	}

	if d, err := time.ParseDuration(c.Extension.Timeout); err != nil {
		c.ClientConfig.Timeout = d
	}

	if c.Extension.Impersonate != "" {
		c.ClientConfig.ImpersonationConfig = &transport.ImpersonationConfig{
			UserName: c.Extension.Impersonate,
			UID:      c.Extension.ImpersonateUID,
			Groups:   strings.Split(c.Extension.ImpersonateGroups, ","),
		}
	}

	if i, err := strconv.ParseBool(c.Extension.InsecureSkipTLSVerify); err == nil {
		c.ClientConfig.TLSConfig.Insecure = i
	}

	if c.Extension.CertificateAuthority != "" || c.Extension.ClientCertificate != "" {
		c.ClientConfig.TLSConfig = &transport.TLSConfig{
			CAFile:     c.Extension.CertificateAuthority,
			CertFile:   c.Extension.ClientCertificate,
			KeyFile:    c.Extension.ClientKey,
			ServerName: c.Extension.TLSServerName,
		}
	}
}

func (c *config) CallMethod(name string) any {
	m := reflect.ValueOf(c).MethodByName(name)
	if m.IsValid() {
		v := m.Call(nil)
		if len(v) > 0 && !v[0].IsZero() {
			return v[0].Interface()
		}
	}
	return nil
}

func (c *config) Persist() error {
	c.APIConfig.Contexts[c.APIConfig.CurrentContext].Extensions[ExtensionName] = c.Extension
	return clientcmd.ModifyConfig(c.ConfigAccess, *c.APIConfig, false)
}

func (c *config) Prompt(name string, value reflect.Value, data any) error {
	var p survey.Prompt

	m := name + " : "

	switch d := data.(type) {
	case string:
		p = &survey.Input{
			Message: m,
			Default: d,
		}
	case []string:
		p = &survey.Select{
			Message: m,
			Options: d,
		}
	default:
		p = &survey.Input{
			Message: m,
		}
	}

	return survey.AskOne(p, (*string)(value.Addr().UnsafePointer()))
}

func (c *config) Token() any {
	return c.RESTConfig.BearerToken
}

func (c *config) ClientType() any {
	return []string{client.REST, client.GRPC}
}

func (c *config) InsecureSkipTLSVerify() any {
	return []string{"false", "true"}
}

func (c *config) Host() any {
	routes, err := c.routes()
	if err != nil {
		return err
	}

	var hosts []string
	for _, route := range routes {
		host := "http://" + route.Spec.Host
		if route.Spec.TLS != nil {
			host = "https://" + route.Spec.Host
		}
		hosts = append(hosts, host)
	}
	return hosts
}

func (c *config) routes() ([]*v1.Route, error) {
	coreV1Client, err := corev1.NewForConfig(c.RESTConfig)
	if err != nil {
		return nil, err
	}

	routeV1Client, err := routev1.NewForConfig(c.RESTConfig)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	serviceList, err := coreV1Client.
		Services("").
		List(ctx, metav1.ListOptions{
			LabelSelector: ServiceLabel,
		})
	if err != nil {
		return nil, err
	}
	if len(serviceList.Items) == 0 {
		return nil, errors.New("services for tekton results not found, try manual configuration")
	}

	var routes []*v1.Route
	for _, service := range serviceList.Items {
		routeList, err := routeV1Client.Routes(service.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		if len(routeList.Items) == 0 {
			return nil, errors.New("routes for tekton results not found, try manual configuration")
		}

		for _, route := range routeList.Items {
			if route.Spec.To.Name == service.Name {
				port := route.Spec.Port.TargetPort
				for _, p := range service.Spec.Ports {
					if p.Port == port.IntVal || p.Name == port.StrVal {
						routes = append(routes, &route)
					}
				}
			}
		}
	}
	return routes, nil
}
