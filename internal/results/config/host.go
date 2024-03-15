package config

import (
	"context"
	"errors"
	v1 "github.com/openshift/api/route/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func getRoutes(c *rest.Config) ([]*v1.Route, error) {
	coreV1Client, err := corev1.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	routeV1Client, err := routev1.NewForConfig(c)
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
