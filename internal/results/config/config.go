package config

import (
	"encoding/json"
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/sayan-biswas/kubectl-tekton/internal/results/client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubectl/pkg/cmd/util"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Config interface {
	Get() *client.Config
	GetObject() runtime.Object
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
	Group         string = "results.tekton.dev"
	Version       string = "v1alpha2"
	Kind          string = "Client"
	Path          string = "apis"
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

	return c, c.LoadClientConfig()
}

func (c *config) Get() *client.Config {
	return c.ClientConfig
}

func (c *config) GetObject() runtime.Object {
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
	if e == nil {
		c.SetVersion()
		return c.Set(nil, false)
	}
	return json.Unmarshal(e.(*runtime.Unknown).Raw, c.Extension)
}

func (c *config) SetVersion() {
	c.Extension.TypeMeta.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   Group,
		Version: Version,
		Kind:    Kind,
	})
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

func (c *config) LoadClientConfig() error {
	rc := rest.CopyConfig(c.RESTConfig)

	gv := c.Extension.TypeMeta.GroupVersionKind().GroupVersion()
	rc.GroupVersion = &gv

	if c.Extension.Host != "" {
		rc.Host = c.Extension.Host
	}

	if c.Extension.APIPath != "" {
		rc.APIPath = c.Extension.APIPath
	}

	if c.Extension.Token != "" {
		rc.BearerToken = c.Extension.Token
	}

	if c.Extension.TLSServerName != "" {
		rc.TLSClientConfig.ServerName = c.Extension.TLSServerName
	}

	if i, err := strconv.ParseBool(c.Extension.InsecureSkipTLSVerify); err == nil {
		if i {
			rc.TLSClientConfig = rest.TLSClientConfig{}
		}
		rc.Insecure = i
	}

	if d, err := time.ParseDuration(c.Extension.Timeout); err != nil {
		rc.Timeout = d
	}

	if c.Extension.Impersonate != "" {
		rc.Impersonate = rest.ImpersonationConfig{
			UserName: c.Extension.Impersonate,
			UID:      c.Extension.ImpersonateUID,
			Groups:   strings.Split(c.Extension.ImpersonateGroups, ","),
		}
	}

	if c.Extension.CertificateAuthority != "" || c.Extension.ClientCertificate != "" || c.Extension.ClientKey != "" {
		rc.TLSClientConfig = rest.TLSClientConfig{
			CAFile:   c.Extension.CertificateAuthority,
			CertFile: c.Extension.ClientCertificate,
			KeyFile:  c.Extension.ClientKey,
		}
	}

	tc, err := rc.TransportConfig()
	if err != nil {
		return err
	}

	rc.APIPath = path.Join(rc.APIPath, Path)
	u, p, err := rest.DefaultServerUrlFor(rc)
	if err != nil {
		return err
	}
	u.Path = p

	c.ClientConfig = &client.Config{
		Transport:  tc,
		URL:        u,
		Timeout:    c.RESTConfig.Timeout,
		ClientType: c.Extension.ClientType,
	}

	return nil
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
	routes, err := getRoutes(c.RESTConfig)
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
