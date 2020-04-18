/*
Copyright 2020 The CRDS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tenant

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func toInt32Ptr(i int) *int32 {
	i32 := int32(i)
	return &i32
}

func toInt64Ptr(i int) *int64 {
	i64 := int64(i)
	return &i64
}

func toBoolPtr(b bool) *bool {
	return &b
}

func toStringPtr(s string) *string {
	return &s
}

func toHostPathTypePtr(h corev1.HostPathType) *corev1.HostPathType {
	return &h
}

// NewAPIServer is a kube-apiserver instance configured against kind defaults.
func NewAPIServer(identifier, namespace string, securePort int) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("tenant-kube-apiserver-%s", identifier),
			Namespace: namespace,
			Labels: map[string]string{
				"kindred.crds.dev/tenant-api-server": identifier,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "kube-apiserver",
					Image: "k8s.gcr.io/kube-apiserver:v1.17.0",
					Command: []string{
						"kube-apiserver",
						"--advertise-address=172.17.0.2",
						"--allow-privileged=true",
						"--authorization-mode=Node,RBAC",
						"--client-ca-file=/etc/kubernetes/pki/ca.crt",
						"--enable-admission-plugins=NodeRestriction",
						"--enable-bootstrap-token-auth=true",
						"--etcd-cafile=/etc/kubernetes/pki/etcd/ca.crt",
						"--etcd-certfile=/etc/kubernetes/pki/apiserver-etcd-client.crt",
						"--etcd-keyfile=/etc/kubernetes/pki/apiserver-etcd-client.key",
						"--etcd-servers=https://127.0.0.1:2379",
						fmt.Sprintf("%s-%s/", "--etcd-prefix=/tenant", identifier),
						"--insecure-port=0",
						"--kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt",
						"--kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key",
						"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
						"--proxy-client-cert-file=/etc/kubernetes/pki/front-proxy-client.crt",
						"--proxy-client-key-file=/etc/kubernetes/pki/front-proxy-client.key",
						"--requestheader-allowed-names=front-proxy-client",
						"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
						"--requestheader-extra-headers-prefix=X-Remote-Extra-",
						"--requestheader-group-headers=X-Remote-Group",
						"--requestheader-username-headers=X-Remote-User",
						fmt.Sprintf("%s=%d", "--secure-port", securePort),
						"--service-account-key-file=/etc/kubernetes/pki/sa.pub",
						"--service-cluster-ip-range=10.96.0.0/12",
						"--tls-cert-file=/etc/kubernetes/pki/apiserver.crt",
						"--tls-private-key-file=/etc/kubernetes/pki/apiserver.key",
					},
					ImagePullPolicy: corev1.PullIfNotPresent,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("250m"),
						},
					},
					LivenessProbe: &corev1.Probe{
						FailureThreshold:    8,
						InitialDelaySeconds: 15,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						TimeoutSeconds:      15,
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "172.17.0.2",
								Path:   "/healthz",
								Port:   intstr.FromInt(securePort), // must match secure-port above
								Scheme: corev1.URISchemeHTTPS,
							},
						},
					},
					TerminationMessagePath:   "/dev/termination-log",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/etc/ssl/certs",
							Name:      "ca-certs",
							ReadOnly:  true,
						},
						{
							MountPath: "/etc/ca-certificates",
							Name:      "etc-ca-certificates",
							ReadOnly:  true,
						},
						{
							MountPath: "/etc/kubernetes/pki",
							Name:      "k8s-certs",
							ReadOnly:  true,
						},
						{
							MountPath: "/usr/local/share/ca-certificates",
							Name:      "usr-local-share-ca-certificates",
							ReadOnly:  true,
						},
						{
							MountPath: "/usr/share/ca-certificates",
							Name:      "usr-share-ca-certificates",
							ReadOnly:  true,
						},
					},
				},
			},
			DNSPolicy:                     corev1.DNSClusterFirst,
			EnableServiceLinks:            toBoolPtr(true),
			HostNetwork:                   true,
			NodeName:                      "kind-control-plane",
			Priority:                      toInt32Ptr(2000000000),
			PriorityClassName:             "system-cluster-critical",
			RestartPolicy:                 corev1.RestartPolicyAlways,
			SchedulerName:                 "default-scheduler",
			TerminationGracePeriodSeconds: toInt64Ptr(30),
			Tolerations: []corev1.Toleration{
				{
					Effect:   corev1.TaintEffectNoExecute,
					Operator: corev1.TolerationOpExists,
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "ca-certs",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/ssl/certs",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "etc-ca-certificates",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/ca-certificates",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "k8s-certs",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/kubernetes/pki",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "usr-local-share-ca-certificates",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/usr/local/share/ca-certificates",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "usr-share-ca-certificates",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/usr/share/ca-certificates",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
			},
		},
	}
}

// NewControllerManager is a kube-controller-manager instance configured against kind defaults.
func NewControllerManager(identifier, namespace string, securePort int) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("tenant-kube-controller-manager-%s", identifier),
			Namespace: namespace,
			Labels: map[string]string{
				"kindred.crds.dev/tenant-controller-manager": identifier,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "kube-controller-manager",
					Image: "k8s.gcr.io/kube-controller-manager:v1.17.0",
					Command: []string{
						"kube-controller-manager",
						// "--cluster-name=tenant-kubernetes", // TODO(hasheddan): create api server service for connection on non-linux machines
						"--allocate-node-cidrs=true",
						"--authentication-kubeconfig=/etc/kubernetes/tenant-controller-manager.conf", // modifiable
						"--authorization-kubeconfig=/etc/kubernetes/tenant-controller-manager.conf",  // modifiable
						"--bind-address=127.0.0.1",
						fmt.Sprintf("%s=%d", "--secure-port", securePort),
						"--port=0", // needed to override default address and port
						"--client-ca-file=/etc/kubernetes/pki/ca.crt",
						"--cluster-cidr=10.244.0.0/16",
						"--cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt",
						"--cluster-signing-key-file=/etc/kubernetes/pki/ca.key",
						"--controllers=clusterrole-aggregation,garbagecollector,serviceaccount,serviceaccount-token,namespace", // modifiable
						"--enable-hostpath-provisioner=true",
						"--kubeconfig=/etc/kubernetes/tenant-controller-manager.conf", // modifiable
						"--leader-elect=true",
						"--node-cidr-mask-size=24",
						"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
						"--root-ca-file=/etc/kubernetes/pki/ca.crt",
						"--service-account-private-key-file=/etc/kubernetes/pki/sa.key",
						"--service-cluster-ip-range=10.96.0.0/12",
						"--use-service-account-credentials=true",
					},
					ImagePullPolicy: corev1.PullIfNotPresent,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("200m"),
						},
					},
					LivenessProbe: &corev1.Probe{
						FailureThreshold:    8,
						InitialDelaySeconds: 15,
						PeriodSeconds:       10,
						SuccessThreshold:    1,
						TimeoutSeconds:      15,
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Host:   "127.0.0.1",
								Path:   "/healthz",
								Port:   intstr.FromInt(securePort), // must match secure-port above
								Scheme: corev1.URISchemeHTTPS,
							},
						},
					},
					TerminationMessagePath:   "/dev/termination-log",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/etc/ssl/certs",
							Name:      "ca-certs",
							ReadOnly:  true,
						},
						{
							MountPath: "/etc/ca-certificates",
							Name:      "etc-ca-certificates",
							ReadOnly:  true,
						},
						{
							MountPath: "/usr/libexec/kubernetes/kubelet-plugins/volume/exec",
							Name:      "flexvolume-dir",
							ReadOnly:  true,
						},
						{
							MountPath: "/etc/kubernetes/pki",
							Name:      "k8s-certs",
							ReadOnly:  true,
						},
						{
							MountPath: "/etc/kubernetes",
							Name:      "kubeconfig",
							ReadOnly:  true,
						},
						{
							MountPath: "/usr/local/share/ca-certificates",
							Name:      "usr-local-share-ca-certificates",
							ReadOnly:  true,
						},
						{
							MountPath: "/usr/share/ca-certificates",
							Name:      "usr-share-ca-certificates",
							ReadOnly:  true,
						},
					},
				},
			},
			DNSPolicy:                     corev1.DNSClusterFirst,
			EnableServiceLinks:            toBoolPtr(true),
			HostNetwork:                   true,
			NodeName:                      "kind-control-plane",
			Priority:                      toInt32Ptr(2000000000),
			PriorityClassName:             "system-cluster-critical",
			RestartPolicy:                 corev1.RestartPolicyAlways,
			SchedulerName:                 "default-scheduler",
			TerminationGracePeriodSeconds: toInt64Ptr(30),
			Tolerations: []corev1.Toleration{
				{
					Effect:   corev1.TaintEffectNoExecute,
					Operator: corev1.TolerationOpExists,
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "ca-certs",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/ssl/certs",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "etc-ca-certificates",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/ca-certificates",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "flexvolume-dir",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/usr/libexec/kubernetes/kubelet-plugins/volume/exec",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "k8s-certs",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/kubernetes/pki",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "kubeconfig",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: fmt.Sprintf("%s-%s", "tenant-kubeconfig", identifier),
							Items: []corev1.KeyToPath{
								{
									Key:  "kubeconfig",
									Path: "tenant-controller-manager.conf",
								},
							},
						},
					},
				},
				{
					Name: "usr-local-share-ca-certificates",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/usr/local/share/ca-certificates",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
				{
					Name: "usr-share-ca-certificates",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/usr/share/ca-certificates",
							Type: toHostPathTypePtr(corev1.HostPathDirectoryOrCreate),
						},
					},
				},
			},
		},
	}
}
