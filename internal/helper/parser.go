package helper

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"regexp"
	"strings"
)

func ParseSelector(s string) map[string]string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	s = strings.ReplaceAll(s, " ", "")
	selectors := strings.Split(s, ",")
	m := map[string]string{}
	r := regexp.MustCompile(`={1,2}`)
	for _, selector := range selectors {
		parts := r.Split(selector, 2)
		switch len(parts) {
		case 1:
			m[parts[0]] = ""
		case 2:
			m[parts[0]] = parts[1]
		}
	}
	return m
}

func ParseLabels(s string) map[string]string {
	return ParseSelector(s)
}

func ParseAnnotations(s string) map[string]string {
	return ParseSelector(s)
}

func ParseFinalizers(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	s = strings.ReplaceAll(s, " ", "")
	finalizers := strings.Split(s, ",")
	return finalizers
}

func ParseOwnerReferences(s string) []v1.OwnerReference {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	substrings := strings.Split(s, ",")
	var ownerReferences []v1.OwnerReference
	for _, str := range substrings {
		ownerReference := v1.OwnerReference{}
		fm := map[string]string{}
		fields := strings.Fields(str)
		for _, field := range fields {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) == 2 {
				fm[parts[0]] = parts[1]
			}
		}
		v := reflect.ValueOf(&ownerReference).Elem()
		for i := 0; i < v.NumField(); i++ {
			name, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("json"), ",")
			if fv, ok := fm[name]; ok && v.Field(i).Kind() == reflect.String {
				v.Field(i).SetString(fv)
				break
			}
		}
		ownerReferences = append(ownerReferences, ownerReference)
	}
	return ownerReferences
}

func ParseArgs(args []string) map[string]string {
	if len(args) == 0 {
		return nil
	}
	m := make(map[string]string)
	for _, arg := range args {
		arg = strings.ReplaceAll(arg, "\"", "")
		kv := strings.SplitN(arg, "=", 2)
		switch len(kv) {
		case 1:
			m[kv[0]] = ""
		case 2:
			m[kv[0]] = kv[1]
		}
	}
	return m
}
