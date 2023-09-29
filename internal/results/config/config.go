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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/transport"
	"time"
)

type Config interface {
	ClientConfig() *client.Options
	RawConfig() runtime.Object
	UpdateConfig() error
}

type config struct {
	configAccess     clientcmd.ConfigAccess
	kubeConfig       *api.Config
	kubeClientConfig *rest.Config
	resultsConfig    *client.Options
	resultsExtension *Extension
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
	c.configAccess = clientConfig.ConfigAccess()
	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return nil, err
	}
	c.kubeConfig = &rawConfig

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	c.kubeClientConfig = restConfig

	c.resultsConfig = new(client.Options)
	c.resultsExtension = new(Extension)

	ctx := c.kubeConfig.Contexts[c.kubeConfig.CurrentContext]
	if ctx == nil {
		return nil, errors.New("current context not set in kubeconfig")
	}

	ext := ctx.Extensions[ExtensionName]
	if ext != nil {
		obj := ext.(*runtime.Unknown)
		err = json.Unmarshal(obj.Raw, c.resultsExtension)
		if err != nil {
			return nil, err
		}
	}

	err = c.populateConfig()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *config) ClientConfig() *client.Options {
	return c.resultsConfig
}

func (c *config) RawConfig() runtime.Object {
	return c.resultsExtension
}

func (c *config) UpdateConfig() error {
	c.resultsExtension = new(Extension)
	return c.populateConfig()
}

func (c *config) populateConfig() error {
	// Host
	if c.resultsExtension.Host == "" {
		host, err := c.getHost()
		if err != nil {
			return err
		}
		c.resultsExtension.Host = host
	}
	c.resultsConfig.Host = c.resultsExtension.Host

	// Client type
	if c.resultsExtension.Client == "" {
		ct, err := c.getClient()
		if err != nil {
			return err
		}
		c.resultsExtension.Client = ct
	}
	c.resultsConfig.ClientType = client.Type(c.resultsExtension.Client)

	// Token
	if c.resultsExtension.Token == "" {
		c.resultsExtension.Token = c.kubeClientConfig.BearerToken
		t, err := c.getToken()
		if err != nil {
			return err
		}
		if t != "" {
			c.resultsExtension.Token = t
		}
	}
	c.resultsConfig.Token = c.resultsExtension.Token

	// Default config
	c.resultsConfig.Timeout = time.Second * 10
	c.resultsConfig.TLSConfig = &transport.TLSConfig{
		Insecure: true,
	}
	ic := transport.ImpersonationConfig(c.kubeClientConfig.Impersonate)
	c.resultsConfig.ImpersonationConfig = &ic

	// Update kubeconfig
	c.kubeConfig.Contexts[c.kubeConfig.CurrentContext].Extensions[ExtensionName] = c.resultsExtension
	err := clientcmd.ModifyConfig(c.configAccess, *c.kubeConfig, false)
	if err != nil {
		return err
	}

	return nil
}

func (c *config) getRoutes() ([]*v1.Route, error) {
	kubeClient, err := kubernetes.NewForConfig(c.kubeClientConfig)
	if err != nil {
		return nil, err
	}

	routeClient, err := routev1.NewForConfig(c.kubeClientConfig)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	serviceList, err := kubeClient.
		CoreV1().
		Services("").
		List(ctx, metav1.ListOptions{
			LabelSelector: ServiceLabel,
		})
	if err != nil {
		return nil, err
	}
	if len(serviceList.Items) == 0 {
		return nil, errors.New("tekton results service not found, try manual configuration")
	}

	var routes []*v1.Route
	for _, service := range serviceList.Items {
		routeList, err := routeClient.Routes(service.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		if len(routeList.Items) == 0 {
			return nil, errors.New("tekton results routes not found, try manual configuration")
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

func (c *config) getHost() (string, error) {
	var host string
	routes, err := c.getRoutes()
	if err != nil {
		return host, survey.AskOne(&survey.Input{
			Message: "Host : ",
		}, &host, survey.WithValidator(survey.Required))
	}

	var options []string
	for _, r := range routes {
		host = "http://" + r.Spec.Host
		if r.Spec.TLS != nil {
			host = "https://" + r.Spec.Host
		}
		options = append(options, host)
	}
	return host, survey.AskOne(&survey.Select{
		Message: "Tekton Results Routes:",
		Options: options,
		Default: options[0],
		Description: func(value string, index int) string {
			return fmt.Sprintf("[%s]", routes[index].Namespace)
		},
	}, &host, survey.WithValidator(survey.Required))
}

func (c *config) getClient() (string, error) {
	var client string
	return client, survey.AskOne(&survey.Select{
		Message: "Client Type :",
		Options: []string{"GRPC", "REST"},
	}, &client)
}

func (c *config) getToken() (string, error) {
	var token string
	return token, survey.AskOne(&survey.Input{
		Message: "Token :",
		Default: c.kubeClientConfig.BearerToken,
	}, &token, survey.WithValidator(survey.Required))
}
