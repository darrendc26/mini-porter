package k8s

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/darrendc26/mini-porter/internal/config"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// var ingress *networkingv1.Ingress

func CreateIngress(client *kubernetes.Clientset, cfg *config.Config, deployedServices []ServiceInfo) error {
	ctx := context.TODO()
	namespace := cfg.Name

	pathType := networkingv1.PathTypePrefix
	var paths []networkingv1.HTTPIngressPath

	seen := make(map[string]bool)

	for _, svc := range deployedServices {
		path := "/" + svc.Name

		if svc.Name == "frontend" {
			path = "/"
		}

		if seen[path] {
			continue
		}
		seen[path] = true

		paths = append(paths, networkingv1.HTTPIngressPath{
			Path:     path,
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

	if len(paths) == 0 {
		return fmt.Errorf("no services to expose via ingress")
	}

	classname := "nginx"

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
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &classname,
			Rules:            []networkingv1.IngressRule{rule},
		},
	}

	ingressesClient := client.NetworkingV1().Ingresses(namespace)

	_, err := ingressesClient.Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Println("[6/6] Ingress Exists, updating...")

			existing, err := ingressesClient.Get(ctx, cfg.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			existing.Spec = ingress.Spec
			existing.Annotations = ingress.Annotations

			_, err = ingressesClient.Update(ctx, existing, metav1.UpdateOptions{})
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

	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		ing := svc.Status.LoadBalancer.Ingress[0]

		if ing.IP != "" {
			baseURL = fmt.Sprintf("http://%s", ing.IP)
		} else if ing.Hostname != "" {
			baseURL = fmt.Sprintf("http://%s", ing.Hostname)
		}
	}

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

	if baseURL == "" {
		return nil, fmt.Errorf("could not determine ingress URL")
	}

	if len(deployedServices) == 1 {
		urls = append(urls, baseURL+"/")
	} else {
		for _, serv := range deployedServices {
			urls = append(urls, fmt.Sprintf("%s/%s", baseURL, serv.Name))
		}
	}

	return urls, nil
}

// func GetIngressBaseURL(client *kubernetes.Clientset, ctx context.Context) (string, error) {
// 	// ctx := context.TODO()
// 	svc, err := client.CoreV1().
// 		Services("ingress-nginx").
// 		Get(ctx, "ingress-nginx-controller", metav1.GetOptions{})
// 	if err != nil {
// 		return "", err
// 	}

// 	var baseURL string

// 	if len(svc.Status.LoadBalancer.Ingress) > 0 {
// 		ing := svc.Status.LoadBalancer.Ingress[0]

// 		if ing.IP != "" {
// 			baseURL = fmt.Sprintf("http://%s", ing.IP)
// 		} else if ing.Hostname != "" {
// 			baseURL = fmt.Sprintf("http://%s", ing.Hostname)
// 		}
// 	}

// 	if baseURL == "" && svc.Spec.Type == corev1.ServiceTypeNodePort {
// 		nodePort := svc.Spec.Ports[0].NodePort

// 		nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
// 		if err != nil {
// 			return "", err
// 		}

// 		for _, addr := range nodes.Items[0].Status.Addresses {
// 			if addr.Type == corev1.NodeInternalIP {
// 				baseURL = fmt.Sprintf("http://%s:%d", addr.Address, nodePort)
// 				break
// 			}
// 		}
// 	}

// 	if baseURL == "" {
// 		WaitForURL(client, ctx)
// 	}

// 	return baseURL, nil
// }

func WaitForURL(client *kubernetes.Clientset, ctx context.Context) (string, error) {
	for i := 0; i < 30; i++ {
		svc, err := client.CoreV1().
			Services("ingress-nginx").
			Get(ctx, "ingress-nginx-controller", metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			ing := svc.Status.LoadBalancer.Ingress[0]

			if ing.IP != "" {
				fmt.Println("✅ Ingress ready at:", ing.IP)
				return fmt.Sprintf("http://%s", ing.IP), nil
			}
			if ing.Hostname != "" {
				fmt.Println("Ingress ready at:", ing.Hostname)
				return fmt.Sprintf("http://%s", ing.Hostname), nil
			}
		}

		if svc.Spec.Type == corev1.ServiceTypeNodePort {
			nodePort := svc.Spec.Ports[0].NodePort

			nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			if err != nil {
				return "", err
			}

			for _, addr := range nodes.Items[0].Status.Addresses {
				if addr.Type == corev1.NodeExternalIP {
					fmt.Println("Using NodePort fallback:", addr.Address)
					return fmt.Sprintf("http://%s:%d", addr.Address, nodePort), nil
				}
			}
		}

		fmt.Println("Waiting for ingress external IP...")
		time.Sleep(10 * time.Second)
	}

	return "", fmt.Errorf("timeout waiting for ingress external IP")
}

func IngressExists(client *kubernetes.Clientset, cfg *config.Config) bool {
	_, err := client.CoreV1().Services("ingress-nginx").Get(context.TODO(), "ingress-nginx-controller", metav1.GetOptions{})

	return err == nil
}

func InstallIngressMinikube() error {
	cmd := exec.Command(
		"minikube",
		"addons",
		"enable",
		"ingress",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func InstallIngressNginx() error {
	cmd := exec.Command(
		"kubectl",
		"apply",
		"-f",
		"https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func WaitForIngress(client *kubernetes.Clientset, cfg *config.Config) error {
	for {
		svc, err := client.CoreV1().Services("ingress-nginx").Get(context.TODO(), "ingress-nginx-controller", metav1.GetOptions{})
		if err == nil && svc.Status.LoadBalancer.Ingress != nil {
			fmt.Println("Ingress Ready")
			return nil
		}

		fmt.Println("Waiting for Ingress...")
		time.Sleep(5 * time.Second)
	}
}
