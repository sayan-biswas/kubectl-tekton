package action

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strings"
)

type Options struct {
	metav1.ListOptions
	metav1.ObjectMeta
	Filter string
}

func (o *Options) validate() error {
	switch {
	case o.Namespace == "":
		o.Namespace = "-"
	}
	return nil
}

func (o *Options) filter() string {
	const (
		contains = "data.metadata.%s.contains(\"%s\")"
		equal    = "data.metadata.%s[\"%s\"]==\"%s\""
		dataType = "data_type==\"%s.%s\""
	)

	var filters []string

	if strings.TrimSpace(o.Filter) != "" {
		filters = append(filters, o.Filter)
	}

	if o.Kind != "" && o.APIVersion != "" {
		filters = append(filters, fmt.Sprintf(dataType, o.APIVersion, o.Kind))
	}

	// TODO: add support for other types
	v := reflect.ValueOf(o.ObjectMeta)
	for i := 0; i < v.NumField(); i++ {
		name, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("json"), ",")
		if len(name) == 0 {
			name = v.Type().Field(i).Name
		}
		value := v.Field(i).Interface()
		k := v.Type().Field(i).Type.Kind()
		switch k {
		case reflect.String:
			if fmt.Sprintf("%v", value) != "" {
				filters = append(filters, fmt.Sprintf(contains, name, value))
			}
		case reflect.Map:
			if m := value.(map[string]string); len(m) > 0 {
				for k, v := range m {
					if v == "" {
						filters = append(filters, fmt.Sprintf(contains, name, k))
					} else {
						filters = append(filters, fmt.Sprintf(equal, name, k, v))
					}
				}
			}
		case reflect.Slice:
			s := reflect.ValueOf(value)
			k := s.Type().Elem().Kind()
			for i := 0; i < s.Len(); i++ {
				v := s.Index(i)
				switch k {
				case reflect.Struct:
					for i := 0; i < v.NumField(); i++ {
						if v := v.Field(i); !v.IsZero() {
							value := v.Interface()
							filters = append(filters, fmt.Sprintf(contains, name, value))
						}
					}
				case reflect.String, reflect.Int:
					value := v.Interface()
					filters = append(filters, fmt.Sprintf(contains, name, value))
				}
			}
		}
	}
	return strings.Join(filters, " && ")
}
