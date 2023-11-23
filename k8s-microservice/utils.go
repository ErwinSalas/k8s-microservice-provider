package k8smicroservice

import (
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getKubeConfigPath(kubeconfigPath string) string {
	var kubeconfig string

	if kubeconfigPath == "" {
		// If kubeconfigPath is not provided, use the default location (~/.kube/config).
		home := homedir.HomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		// Use the provided kubeconfigPath.
		kubeconfig = kubeconfigPath
	}

	return kubeconfig
}

func createKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	kubeconfig := getKubeConfigPath(kubeconfigPath)

	// Load the kubeconfig file.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// Create a Kubernetes clientset.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func createKubeConfigForHelm(d *schema.ResourceData, namespace string) clientcmd.ClientConfig {
	kubeConfigPath := d.Get("kubeconfig_path").(string)
	context := d.Get("kube_context").(string)
	kubeconfig := getKubeConfigPath(kubeConfigPath)
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig
	overrides := &clientcmd.ConfigOverrides{}
	overrides.Context.Namespace = namespace
	overrides.CurrentContext = context
	// Create a client config using the loading rules.
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
}
