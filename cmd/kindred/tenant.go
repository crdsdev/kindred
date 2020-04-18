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

package kindred

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"text/tabwriter"
	"time"

	"github.com/crdsdev/kindred/pkg/tenant"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	tenantAPIServerPrefix         = "tenant-kube-apiserver"
	tenantControllerManagerPrefix = "tenant-kube-controller-manager"
	tenantKubeconfigSecretPrefix  = "tenant-kubeconfig"

	tenantNamespacePrefix = "kube-tenant"

	defaultSysNamespace = "kube-system"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	rootCmd.AddCommand(tenantCmd)
	tenantCmd.AddCommand(createCmd)
	tenantCmd.AddCommand(listCmd)
	tenantCmd.AddCommand(configCmd)
}

var tenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Commands for configuring tenant Kubernetes insides of a cluster.",
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create and configure a new Kube API Server and Controller Manager.",
	Long:  `Create and configure a new Kube API Server and Controller Manager in the currently configured cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		identifier := fmt.Sprint(rand.Intn(50000))

		// TODO(hasheddan): select ports in a sensible manner
		serverPort := rand.Intn(8000-7000) + 7000
		managerPort := rand.Intn(12000-11000) + 11000

		cfg, err := ctrl.GetConfig()
		if err != nil {
			panic(err)
		}

		c, err := client.New(cfg, client.Options{})
		if err != nil {
			panic(err)
		}

		tenantNS := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s", tenantNamespacePrefix, identifier),
			},
		}

		fmt.Printf("Creating tenant namespace: %s\n", tenantNS.Name)
		if err := c.Create(context.TODO(), tenantNS); err != nil {
			panic(err)
		}

		kubeAPI := &corev1.Pod{}
		if err := c.Get(context.TODO(), types.NamespacedName{Name: "kube-apiserver-kind-control-plane", Namespace: defaultSysNamespace}, kubeAPI); err != nil {
			panic(err)
		}

		fmt.Printf("Creating tenant kubeconfig secret: %s/%s-%s\n", tenantNS.Name, tenantKubeconfigSecretPrefix, identifier)
		s, err := secretForTenantKubeConfig(cfg, kubeAPI.Status.HostIP, identifier, tenantNS.Name, serverPort)
		if err != nil {
			panic(err)
		}

		if err := c.Create(context.TODO(), s); err != nil {
			panic(err)
		}

		fmt.Printf("Creating tenant API server: %s/%s-%s\n", tenantNS.Name, tenantAPIServerPrefix, identifier)
		server := tenant.NewAPIServer(identifier, tenantNS.Name, serverPort)
		if err := c.Create(context.TODO(), server); err != nil {
			panic(err)
		}

		fmt.Printf("Creating tenant controller manager: %s/%s-%s\n", tenantNS.Name, tenantAPIServerPrefix, identifier)
		controller := tenant.NewControllerManager(identifier, tenantNS.Name, managerPort)
		if err := c.Create(context.TODO(), controller); err != nil {
			panic(err)
		}

		fmt.Printf("Done! Run...\n\nkindred tenant config %s\n\n...to access your tenant cluster!\n", identifier)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tenant Kubernetes instances in cluster.",
	Long:  `Returns a list of all tenant Kubernetes instances in the currently configured cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)

		cfg, err := ctrl.GetConfig()
		if err != nil {
			panic(err)
		}

		c, err := client.New(cfg, client.Options{})
		if err != nil {
			panic(err)
		}

		apiServers := &corev1.PodList{}
		if err := c.List(context.TODO(), apiServers, client.HasLabels([]string{"kindred.crds.dev/tenant-api-server"})); err != nil {
			panic(err)
		}

		if len(apiServers.Items) == 0 {
			fmt.Println("No tenant instances found.")
		} else {
			fmt.Fprint(w, "NAMESPACE\tAPI SERVER\tCONTROLLER MANAGER\tKUBECONFIG\n")
		}

		for _, s := range apiServers.Items {
			identifier := s.Labels["kindred.crds.dev/tenant-api-server"]
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Namespace, s.Name, fmt.Sprintf("%s-%s", "tenant-kube-controller-manager", identifier), fmt.Sprintf("%s-%s", "tenant-kubeconfig", identifier))
		}

		if err := w.Flush(); err != nil {
			panic(err)
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Get kubeconfig for tenant Kubernetes cluster.",
	Long:  `Get the kubeconfig for a tenant Kubernetes cluster by identifier.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		identifier := args[0]

		cfg, err := ctrl.GetConfig()
		if err != nil {
			panic(err)
		}

		c, err := client.New(cfg, client.Options{})
		if err != nil {
			panic(err)
		}

		secretName := fmt.Sprintf("%s-%s", tenantKubeconfigSecretPrefix, identifier)
		secretNS := fmt.Sprintf("%s-%s", tenantNamespacePrefix, identifier)

		tenantKubeconfig := &corev1.Secret{}
		if err := c.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: secretNS}, tenantKubeconfig); err != nil {
			panic(err)
		}

		fmt.Print(string(tenantKubeconfig.Data["kubeconfig"]))
	},
}

func secretForTenantKubeConfig(cfg *rest.Config, host, identifier, namespace string, port int) (*corev1.Secret, error) {
	kube := clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			cfg.ServerName: {
				Server:                   fmt.Sprintf("https://%s:%d", host, port),
				CertificateAuthorityData: cfg.CAData,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			cfg.ServerName: {
				Cluster: cfg.ServerName,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			cfg.ServerName: {
				ClientCertificateData: cfg.CertData,
				ClientKeyData:         cfg.KeyData,
			},
		},
		CurrentContext: cfg.ServerName,
	}

	b, err := clientcmd.Write(kube)
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", tenantKubeconfigSecretPrefix, identifier),
			Namespace: namespace,
			Labels: map[string]string{
				"kindred.crds.dev/tenant-api-server": identifier,
			},
		},
		Data: map[string][]byte{
			"kubeconfig": b,
		},
	}, nil
}
