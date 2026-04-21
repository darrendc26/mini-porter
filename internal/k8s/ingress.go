package k8s

import (
	"context"
	"fmt"

	"github.com/darrendc26/mini-porter/internal/config"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// var ingress *networkingv1.Ingress

func CreateIngress(client *kubernetes.Clientset, cfg *config.Config, deployedServices []ServiceInfo) error {
	pathType := networkingv1.PathTypePrefix
	var paths []networkingv1.HTTPIngressPath

	// 1. Add non-root paths first
	for _, svc := range deployedServices {
		if svc.Name != "frontend" {
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

		if svc.Name == "frontend" {
			paths = append(paths, networkingv1.HTTPIngressPath{
				Path:     "/",
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
	}

	rule := networkingv1.IngressRule{
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: paths,
			},
		},
	}

	if cfg.Domain != "" {
		rule.Host = cfg.Domain
	}

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.Name,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{rule},
		},
	}

	ingressesClient := client.NetworkingV1().Ingresses("default")

	_, err := ingressesClient.Create(context.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Println("[6/6] Ingress Exists, updating...")

			existing, err := ingressesClient.Get(context.TODO(), cfg.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			ingress.ResourceVersion = existing.ResourceVersion

			_, err = ingressesClient.Update(context.TODO(), ingress, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		fmt.Println("[6/6] Ingress Created")
	}

	fmt.Println("[6/6] Ingress Completed")

	// Output (generic, works everywhere)
	fmt.Println("\nApp URLs:")
	if cfg.Domain != "" {
		for _, svc := range deployedServices {
			if svc.Name == "frontend" {
				fmt.Printf("%s: http://%s/\n", svc.Name, cfg.Domain)
			} else {
				fmt.Printf("%s: http://%s/%s\n", svc.Name, cfg.Domain, svc.Name)
			}
		}
	} else {
		fmt.Println("Use your cluster IP:")
		fmt.Println("  minikube ip   (local)")
		fmt.Println("  kubectl get svc -n ingress-nginx   (cloud)")
		fmt.Println("Then access:")
		for _, svc := range deployedServices {
			if svc.Name == "frontend" {
				fmt.Println("  http://<IP>/")
			} else {
				fmt.Printf("  http://<IP>/%s\n", svc.Name)
			}
		}
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

func GetIngressURL(client *kubernetes.Clientset, deployedServices []ServiceInfo) ([]string, error) {
	ctx := context.TODO()
	var urls []string

	svc, err := client.CoreV1().
		Services("ingress-nginx").
		Get(ctx, "ingress-nginx-controller", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var baseURL string

	// 1. LoadBalancer (cloud)
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		ing := svc.Status.LoadBalancer.Ingress[0]

		if ing.IP != "" {
			baseURL = fmt.Sprintf("http://%s", ing.IP)
		} else if ing.Hostname != "" {
			baseURL = fmt.Sprintf("http://%s", ing.Hostname)
		}
	}

	// 2. NodePort (local)
	if baseURL == "" && svc.Spec.Type == corev1.ServiceTypeNodePort {
		nodePort := svc.Spec.Ports[0].NodePort

		nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, addr := range nodes.Items[0].Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				baseURL = fmt.Sprintf("http://%s:%d", addr.Address, nodePort)
				break
			}
		}
	}

	// ❌ If still empty → error
	if baseURL == "" {
		return nil, fmt.Errorf("could not determine ingress URL")
	}

	// 3. Build service URLs
	if len(deployedServices) == 1 {
		urls = append(urls, baseURL+"/")
	} else {
		for _, serv := range deployedServices {
			urls = append(urls, fmt.Sprintf("%s/%s", baseURL, serv.Name))
		}
	}

	return urls, nil
}
