package k8s

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// var ingress *networkingv1.Ingress

func CreateIngress(client *kubernetes.Clientset, cfg *config.Config, deployedServices []ServiceInfo) error {
	fmt.Println("[5/5] Creating ingress...")

	pathType := networkingv1.PathTypePrefix

	var paths []networkingv1.HTTPIngressPath

	for _, svc := range deployedServices {
		paths = append(paths, networkingv1.HTTPIngressPath{
			Path:     "/" + svc.Name,
			PathType: &pathType,
			Backend: networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: svc.Name,
					Port: networkingv1.ServiceBackendPort{
						Number: int32(svc.Port),
					},
				},
			},
		})
	}

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.Name,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: cfg.Name + ".miniporter",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: paths,
						},
					},
				},
			},
		},
	}

	ingressesClient := client.NetworkingV1().Ingresses("default")

	_, err := ingressesClient.Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Ingress exists, updating...")

		existing, getErr := ingressesClient.Get(context.TODO(), cfg.Name, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}

		ingress.ResourceVersion = existing.ResourceVersion

		_, err = ingressesClient.Update(context.TODO(), ingress, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	fmt.Println("\nApp URLs:")
	for _, svc := range deployedServices {
		fmt.Printf("http://%s.miniporter/%s\n", cfg.Name, svc.Name)
	}

	return nil
}

func DeleteIngress(client *kubernetes.Clientset, cfg *config.Config, svc ServiceInfo) error {
	fmt.Println("Deleting ingress...")
	ingressesClient := client.NetworkingV1().Ingresses("default")
	err := ingressesClient.Delete(context.TODO(), svc.Name, metav1.DeleteOptions{})
	if err != nil {

		return err
	}
	fmt.Println("Ingress deleted successfully")
	return nil
}
