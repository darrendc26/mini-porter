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
		fmt.Println("Ingress 					Exists, updating...")

		existing, getErr := ingressesClient.Get(context.TODO(), cfg.Name, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}

		ingress.ResourceVersion = existing.ResourceVersion

		_, err = ingressesClient.Update(context.TODO(), ingress, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	} else {
		fmt.Println("[6/6] Ingress 					Creating...")
	}

	fmt.Println("[6/6] Ingress 					Completed")

	fmt.Println("\nApp URLs:")
	for _, svc := range deployedServices {
		fmt.Printf("%s is live: http://%s.miniporter/%s\n", svc.Name, cfg.Name, svc.Name)
	}

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

func DeleteIngressRule(client kubernetes.Interface, cfg *config.Config, serviceName string) error {
	ctx := context.Background()

	ingressClient := client.NetworkingV1().Ingresses("default")

	ingress, err := ingressClient.Get(ctx, cfg.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ingress: %w", err)
	}

	// Filter rules
	for i, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}

		var newPaths []networkingv1.HTTPIngressPath

		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service.Name != serviceName {
				newPaths = append(newPaths, path)
			}
		}

		ingress.Spec.Rules[i].HTTP.Paths = newPaths
	}

	// 🔥 If no paths left → delete ingress
	empty := true
	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP != nil && len(rule.HTTP.Paths) > 0 {
			empty = false
			break
		}
	}

	if empty {
		fmt.Println("No routes left, deleting ingress...")
		return ingressClient.Delete(ctx, cfg.Name, metav1.DeleteOptions{})
	}

	// Update ingress
	_, err = ingressClient.Update(ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ingress: %w", err)
	}

	fmt.Printf("Ingress updated: removed %s\n", serviceName)
	return nil
}
