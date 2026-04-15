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

func CreateIngress(client *kubernetes.Clientset, cfg *config.Config) error {
	fmt.Println("[5/5] Creating ingress...")

	pathType := networkingv1.PathTypePrefix

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
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: cfg.Name,
											Port: networkingv1.ServiceBackendPort{
												// TODO: Handle multiple services and ports
												Number: int32(cfg.Services[0].Port),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ingressesClient := client.NetworkingV1().Ingresses("default")
	_, err := ingressesClient.Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Ingress already exists...")

		_, err := ingressesClient.Update(context.TODO(), ingress, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		fmt.Println("App URL: http://" + cfg.Name + ".miniporter")
		return err
	}

	fmt.Println("App URL: http://" + cfg.Name + ".miniporter")
	return nil
}

func DeleteIngress(client *kubernetes.Clientset, cfg *config.Config) error {
	fmt.Println("Deleting ingress...")
	ingressesClient := client.NetworkingV1().Ingresses("default")
	err := ingressesClient.Delete(context.TODO(), cfg.Name, metav1.DeleteOptions{})
	if err != nil {

		return err
	}
	fmt.Println("Ingress deleted successfully")
	return nil
}
