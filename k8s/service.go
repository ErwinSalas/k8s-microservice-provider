package k8s

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

func ServiceCreate(d *schema.ResourceData, namespace string, clientset *kubernetes.Clientset) error {
	name := d.Get("name").(string)
	serviceKind := d.Get("expose_type").(string)

	containerPort := d.Get("ports.0.container_port").(int)
	servicePort := d.Get("ports.0.service_port").(int)
	protocol := d.Get("ports.0.protocol").(string)

	// Define your Kubernetes Service here.
	service := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       SERVICE_PORT_NAME,
					Protocol:   corev1.Protocol(protocol),
					Port:       int32(servicePort),
					TargetPort: intstr.FromInt(containerPort),
				},
			},
			Selector: map[string]string{},
			Type:     corev1.ServiceType(serviceKind),
		},
	}

	// Create the Service.
	_, err := clientset.CoreV1().Services(namespace).Create(context.TODO(), service, v1.CreateOptions{})
	if err != nil {
		return err
	}

	// Set the ID for Terraform to track.
	d.SetId(name)

	return nil
}

func ServiceRead(d *schema.ResourceData, namespace string, clientset *kubernetes.Clientset) (*corev1.Service, error) {
	name := d.Get("name").(string)

	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return service, err
	}

	return service, nil
}

func ServiceUpdate(d *schema.ResourceData, namespace string, clientset *kubernetes.Clientset) error {
	name := d.Get("name").(string)

	// Define your updated Kubernetes Service here.
	updatedService := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.ServiceSpec{
			// Define your updated service spec here.
		},
	}

	// Update the Service using retry.
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the current Service for conflict resolution.
		currentService, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return err
		}

		// Update the current Service with the updatedService.
		// You should handle any specific updates you need here.
		currentService.Spec = updatedService.Spec

		_, updateErr := clientset.CoreV1().Services(namespace).Update(context.TODO(), currentService, v1.UpdateOptions{})
		return updateErr
	})

	if err != nil {
		return err
	}

	return nil
}

func ServiceDelete(d *schema.ResourceData, namespace string, clientset *kubernetes.Clientset) error {
	name := d.Get("name").(string)

	// Delete the Service.
	err := clientset.CoreV1().Services(namespace).Delete(context.TODO(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	// Wait for the Service to be deleted.
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		_, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return err
	}

	// Clear the ID since the resource is deleted.
	d.SetId("")

	return nil
}
