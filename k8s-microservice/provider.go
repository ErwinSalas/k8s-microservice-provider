package k8smicroservice

import (
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/kubernetes"
)

type K8sMicroserviceProvider struct {
	projectNamespace string
	client           *kubernetes.Clientset
	data             *schema.ResourceData
	// Used to lock some operations
	sync.Mutex
}

func (p *K8sMicroserviceProvider) GetHelmConfiguration(namespace string) (*action.Configuration, error) {
	p.Lock()
	defer p.Unlock()
	fmt.Print("[INFO] GetHelmConfiguration start")
	actionConfig := new(action.Configuration)

	kubeConfig := createKubeConfigForHelm(p.data, namespace)

	// Get a *rest.Config from the RESTClientGetter.
	if err := actionConfig.Init(KubeConfig{ClientConfig: kubeConfig, Burst: 1, Mutex: sync.Mutex{}}, namespace, "secrets", nil); err != nil {
		return nil, err
	}

	fmt.Print("[INFO] GetHelmConfiguration success")
	return actionConfig, nil
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"kube_config": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Kube config path",
			},
			"kube_context": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Kube config path",
			},
			"project_namespace": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Kube config path",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"microservice": resourceMicroservice(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	// Read kubeconfig_path from the provider configuration.
	kubeConfigPath := d.Get("kubeconfig_path").(string)
	namespace := d.Get("project_namespace").(string)

	// Initialize and return your Kubernetes clientset using the kubeConfigPath.
	clientset, err := createKubernetesClient(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return &K8sMicroserviceProvider{
		projectNamespace: namespace,
		client:           clientset,
		data:             d,
	}, nil
}
