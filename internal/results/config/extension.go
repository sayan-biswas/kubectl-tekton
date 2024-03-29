package config

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Extension stores the information about results
type Extension struct {
	runtime.TypeMeta      `json:",inline"`
	ClientType            string `json:"client-type,omitempty"  group:"client"`
	Host                  string `json:"host,omitempty"  group:"client"`
	APIPath               string `json:"api-path,omitempty"  group:"client"`
	InsecureSkipTLSVerify string `json:"insecure-skip-tls-verify,omitempty" group:"client"`
	Timeout               string `json:"timeout,omitempty" group:"client"`
	CertificateAuthority  string `json:"certificate-authority,omitempty" group:"tls"`
	ClientCertificate     string `json:"client-certificate,omitempty" group:"tls"`
	ClientKey             string `json:"client-key,omitempty" group:"tls"`
	TLSServerName         string `json:"tls-server-name,omitempty" group:"tls"`
	Impersonate           string `json:"act-as,omitempty" group:"auth"`
	ImpersonateUID        string `json:"act-as-uid,omitempty" group:"auth"`
	ImpersonateGroups     string `json:"act-as-groups,omitempty" group:"auth"`
	Token                 string `json:"token,omitempty" group:"auth"`
}

// DeepCopy is an autogenerated deep copy function, copying the receiver, creating a new Extension.
func (in *Extension) DeepCopy() *Extension {
	if in == nil {
		return nil
	}
	out := new(Extension)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deep copy function, copying the receiver, creating a new runtime.Object.
func (in *Extension) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

// DeepCopyInto is an autogenerated deep copy function, copying the receiver, writing into out. in must be non-nil.
func (in *Extension) DeepCopyInto(out *Extension) {
	*out = *in
	out.TypeMeta = in.TypeMeta
}
