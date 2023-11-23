package k8s

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

func DeploymentCreate(d *schema.ResourceData, namespace string, clientset *kubernetes.Clientset) error {

	name := d.Get("name").(string)

	imageName := d.Get("image").(string)
	replicas := d.Get("replicas").(int32)

	containerPort := d.Get("ports.0.container_port").(int)
	// Define your Kubernetes Deployment here.
	deployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &v1.LabelSelector{MatchLabels: map[string]string{MICROSERVICE_MATCH_LABEL_KEY: name}},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: imageName,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: int32(containerPort),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create the Deployment.
	_, err := clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, v1.CreateOptions{})
	if err != nil {
		return err
	}

	// Set the ID for Terraform to track.
	d.SetId(name)

	return nil
}

func DeploymentRead(d *schema.ResourceData, namespace string, clientset *kubernetes.Clientset) (*appsv1.Deployment, error) {

	name := d.Id()

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return deployment, err
	}

	return deployment, nil
}

func DeploymentUpdate(d *schema.ResourceData, namespace string, clientset *kubernetes.Clientset) error {
	name := d.Id()

	// Define your updated Kubernetes Deployment here.
	updatedDeployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			// Define your updated deployment spec here.
		},
	}

	// Update the Deployment using retry.
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the current Deployment for conflict resolution.
		currentDeployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return err
		}

		// Update the current Deployment with the updatedDeployment.
		// You should handle any specific updates you need here.
		currentDeployment.Spec = updatedDeployment.Spec

		_, updateErr := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), currentDeployment, v1.UpdateOptions{})
		return updateErr
	})

	if err != nil {
		return err
	}

	return nil
}

func DeploymentDelete(d *schema.ResourceData, namespace string, clientset *kubernetes.Clientset) error {
	name := d.Id()

	// Delete the Deployment.
	err := clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	// Wait for the Deployment to be deleted.
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		_, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, v1.GetOptions{})
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
