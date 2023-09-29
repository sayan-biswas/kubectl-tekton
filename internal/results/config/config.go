package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Config interface {
	Get() *client.Options
	RawConfig() runtime.Object
	Persist() error
	Set(kv map[string]string) error
	Reset() error
}

type config struct {
	ConfigAccess  clientcmd.ConfigAccess
	APIConfig     *api.Config
	RESTConfig    *rest.Config
	ClientOptions *client.Options
	Extension     *Extension
}

const (
	ServiceLabel  string = "app.kubernetes.io/name=tekton-results-api"
	ExtensionName string = "tekton-results"
)

func NewConfig() (Config, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	c := new(config)
	c.ConfigAccess = clientConfig.ConfigAccess()
	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return nil, err
	}
	c.APIConfig = &rawConfig

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	c.RESTConfig = restConfig

	c.ClientOptions = new(client.Options)
	c.Extension = new(Extension)
	c.SetVersion()

	// Config defaults
	c.ClientOptions.Timeout = time.Second * 10
	c.ClientOptions.TLSConfig = &transport.TLSConfig{
		Insecure: true,
	}
	ic := transport.ImpersonationConfig(c.RESTConfig.Impersonate)
	c.ClientOptions.ImpersonationConfig = &ic

	// Check current context is set
	currentContext := c.APIConfig.Contexts[c.APIConfig.CurrentContext]
	if currentContext == nil {
		return nil, errors.New("current context not set in kubeconfig")
	}

	// Load config from extension
	extension := currentContext.Extensions[ExtensionName]
	if extension != nil {
		unknown := extension.(*runtime.Unknown)
		if err = json.Unmarshal(unknown.Raw, c.Extension); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(unknown.Raw, c.ClientOptions); err != nil {
			return nil, err
		}
	} else {
		if err = c.Set(nil); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *config) Get() *client.Options {
	return c.ClientOptions
}

func (c *config) RawConfig() runtime.Object {
	return c.Extension
}

func (c *config) Reset() error {
	c.Extension = new(Extension)
	c.SetVersion()
	return c.Set(nil)
}

func (c *config) SetVersion() {
	c.Extension.TypeMeta = runtime.TypeMeta{
		APIVersion: "v1alpha1",
		Kind:       "Client",
	}
}

func (c *config) Set(m map[string]string) error {
	t := reflect.TypeOf(c.Extension).Elem()
	v := reflect.ValueOf(c.Extension).Elem()
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if input, err := strconv.ParseBool(tf.Tag.Get("prompt")); err != nil || !input {
			continue
		}

		name, _, found := strings.Cut(tf.Tag.Get("json"), ",")
		if !found {
			continue
		}

		vf := v.Field(i)
		if value, ok := m[name]; ok {
			if value != "" {
				vf.SetString(value)
				continue
			}
			vf.SetZero()
		} else if m != nil {
			continue
		}

		if vf.IsZero() {
			if err := c.Prompt(tf, vf, c.CallMethod(tf.Name)); err != nil {
				return err
			}
		}
	}

	return c.Persist()
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

func (c *config) Prompt(field reflect.StructField, value reflect.Value, data any) error {
	if data == nil {
		if o, ok := field.Tag.Lookup("options"); ok {
			data = strings.Split(o, ",")
		}
	}

	var p survey.Prompt
	m := fmt.Sprintf("%s : ", field.Name)

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
