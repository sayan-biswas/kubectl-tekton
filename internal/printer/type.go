package printer

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1 "knative.dev/pkg/apis/duck/v1"
)

type List struct {
	runtime.TypeMeta `json:",inline"`
	NextPageToken    string `json:"nextPageToken,omitempty" yaml:"nextPageToken,omitempty"`
	Items            []struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`
		Status            struct {
			v1.Status      `json:",inline"`
			StartTime      *metav1.Time `json:"startTime,omitempty"`
			CompletionTime *metav1.Time `json:"completionTime,omitempty"`
		} `json:"status,omitempty"`
	} `json:"items"`
}
