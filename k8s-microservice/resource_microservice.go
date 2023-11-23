package k8smicroservice

import (
	"fmt"
	"os"

	"github.com/ErwinSalas/terraform-provider-k8s-microservice/k8s"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMicroservice() *schema.Resource {
	return &schema.Resource{
		Create: resourceMicroserviceCreate,
		Read:   resourceMicroserviceRead,
		Update: resourceMicroserviceUpdate,
		Delete: resourceMicroserviceDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the microservice, will be used as selector label to bind service/deployment among other kinds.",
			},
			"image": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Docker image name and tag (e.g., nginx:latest).",
			},
			"replicas": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Number of replicas to have in the deployment",
			},
			"ports": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Port protocol for service spec.",
						},
						"container_port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The internal port inside the container.",
						},
						"service_port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The external port on the host.",
						},
					},
				},
				Required:    true,
				Description: "Port mappings from container to host.",
			},
			"expose_type": {
				Type:        schema.TypeString,
				Description: "Determines how the service is exposed. Defaults to `ClusterIP`. Valid options are `ExternalName`, `ClusterIP`, `NodePort`, and `LoadBalancer`. `ExternalName` maps to the specified `external_name`. More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types",
				Optional:    true,
				Default:     "ClusterIP",
				ValidateFunc: validation.StringInSlice([]string{
					"ClusterIP",
					"ExternalName",
					"NodePort",
					"LoadBalancer",
				}, false),
			},
			// Add more schema attributes as needed for your Docker container configuration.
		},
	}
}

func resourceMicroserviceCreate(d *schema.ResourceData, m interface{}) error {
	providerConfig := m.(*K8sMicroserviceProvider)
	clientset := providerConfig.client

	err := k8s.DeploymentCreate(d, providerConfig.projectNamespace, clientset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create deployment - %s\n", err)
		return err
	}
	err = k8s.ServiceCreate(d, providerConfig.projectNamespace, clientset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to to create service - %s\n", err)
		return err
	}

	return nil
}

func resourceMicroserviceRead(d *schema.ResourceData, m interface{}) error {
	providerConfig := m.(*K8sMicroserviceProvider)
	clientset := providerConfig.client
	deployment, err := k8s.DeploymentRead(d, providerConfig.projectNamespace, clientset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create deployment - %s\n", err)
		return err
	}
	d.Set("name", deployment.Spec.Selector.MatchLabels[k8s.MICROSERVICE_MATCH_LABEL_KEY])
	d.Set("image", deployment.Spec.Template.Spec.Containers[0].Image)
	d.Set("replicas", deployment.Spec.Replicas)

	service, err := k8s.ServiceRead(d, providerConfig.projectNamespace, clientset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create deployment - %s\n", err)
		return err
	}
	d.Set("expose_type", service.Kind)

	return nil
}

func resourceMicroserviceUpdate(d *schema.ResourceData, m interface{}) error {
	providerConfig := m.(*K8sMicroserviceProvider)
	clientset := providerConfig.client
	err := k8s.DeploymentUpdate(d, providerConfig.projectNamespace, clientset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create deployment - %s\n", err)
		return err
	}

	err = k8s.ServiceUpdate(d, providerConfig.projectNamespace, clientset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create deployment - %s\n", err)
		return err
	}
	return resourceMicroserviceRead(d, m)
}

func resourceMicroserviceDelete(d *schema.ResourceData, m interface{}) error {
	providerConfig := m.(*K8sMicroserviceProvider)
	clientset := providerConfig.client

	err := k8s.DeploymentDelete(d, providerConfig.projectNamespace, clientset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create deployment - %s\n", err)
		return err
	}

	err = k8s.ServiceDelete(d, providerConfig.projectNamespace, clientset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create deployment - %s\n", err)
		return err
	}
	return nil
}
