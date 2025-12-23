package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1" // metav1
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// Define Prometheus metrics.
	apiServerUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "kubernetes_apiserver_up",
		Help: "Whether the app can contact the Kubernetes API server (1 = up, 0 = down)",
	})
	deploymentSpecReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "deployment_spec_replicas",
		Help: "Desired replicas for a deployment",
	}, []string{"deployment_namespace", "deployment"})
	deploymentAvailableReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "deployment_status_replicas_available",
		Help: "Available replicas for a deployment",
	}, []string{"deployment_namespace", "deployment"})
	customDenyPolicies = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "custom_deny_policies_total",
		Help: "Number of CustomDeny operations",
	}, []string{"namespace", "status"})
)

func main() {
	// Parse command-line flag for optional kubeconfig file.
	kubeconfig := flag.String("kubeconfig", "", "Path to kubeconfig (empty for in-cluster)")
	listenAddr := flag.String("address", ":8080", "HTTP server listen address")
	flag.Parse()

	// Build Kubernetes client configuration.
	// If a kubeconfig path is provided, use it; otherwise assume in-cluster config.
	var config *rest.Config
	var err error
	if *kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		log.Fatalf("Failed to build kube config: %v", err)
	}

	// (Optional) Set a default timeout for all requests.
	config.Timeout = 5 * time.Second

	// Create the Kubernetes clientset.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Create dynamic client for CRD and Calico resources
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %v", err)
	}

	version, err := getKubernetesVersion(clientset)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Connected to Kubernetes %s\n", version)

	// Register Prometheus metrics so they will be exposed at /metrics.
	// This is the standard pattern (prometheus.MustRegister) to add custom metrics:contentReference[oaicite:3]{index=3}.
	prometheus.MustRegister(
		apiServerUp,
		deploymentSpecReplicas,
		deploymentAvailableReplicas,
		customDenyPolicies,
	)

	// Goroutine #1: Periodically check the API server health.
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if _, err := clientset.Discovery().ServerVersion(); err != nil {
				apiServerUp.Set(0)
			} else {
				apiServerUp.Set(1)
			}
		}
	}()

	// Goroutine #2: Periodically collect deployment replica metrics.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			// Clear old metrics to avoid stale entries when deployments are deleted.
			deploymentSpecReplicas.Reset() // clears all previous label values:contentReference[oaicite:4]{index=4}
			deploymentAvailableReplicas.Reset()

			// List all deployments in all namespaces, with a timeout.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			deployments, err := clientset.AppsV1().Deployments("").List(ctx, v1.ListOptions{})
			cancel()
			if err != nil {
				log.Printf("Error listing deployments: %v", err)
				continue
			}

			// Iterate over each Deployment and set the metrics.
			for _, deploy := range deployments.Items {
				ns := deploy.Namespace
				name := deploy.Name
				// Desired replicas (spec.Replicas is a *int32).
				var specCount int32
				if deploy.Spec.Replicas != nil {
					specCount = *deploy.Spec.Replicas
				}
				deploymentSpecReplicas.WithLabelValues(ns, name).Set(float64(specCount))
				// Available replicas from status.
				deploymentAvailableReplicas.WithLabelValues(ns, name).Set(float64(deploy.Status.AvailableReplicas))
			}
		}
	}()

	// --- CustomDeny watcher ---
	go watchCustomDeny(dynamicClient)

	if err := startServer(*listenAddr); err != nil {
		panic(err)
	}
}

// getKubernetesVersion returns a string GitVersion of the Kubernetes server defined by the clientset.
//
// If it can't connect an error will be returned, which makes it useful to check connectivity.
func getKubernetesVersion(clientset kubernetes.Interface) (string, error) {
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}

	return version.String(), nil
}

// startServer launches an HTTP server with defined handlers and blocks until it's terminated or fails with an error.
//
// Expects a listenAddr to bind to.
func startServer(listenAddr string) error {
	http.HandleFunc("/healthz", healthHandler)
	http.Handle("/metrics", promhttp.Handler()) // Uses default registry with registered metrics.

	fmt.Printf("Server listening on %s\n", listenAddr)

	return http.ListenAndServe(listenAddr, nil)
}

// healthHandler responds with the health status of the application.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("ok"))
	if err != nil {
		fmt.Println("failed writing to response")
	}
}

func watchCustomDeny(client dynamic.Interface) {
	customDenyGVR := schema.GroupVersionResource{
		Group:    "security.internal.io",
		Version:  "v1",
		Resource: "customdenies",
	}

	for {
		watcher, err := client.Resource(customDenyGVR).
			Namespace(v1.NamespaceAll).
			Watch(context.Background(), v1.ListOptions{})
		if err != nil {
			log.Printf("CustomDeny watch failed: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		for event := range watcher.ResultChan() {
			obj, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}

			switch event.Type {
			case watch.Added, watch.Modified:
				if err := handleCustomDeny(client, obj); err != nil {
					log.Printf("Apply failed %s/%s: %v", obj.GetNamespace(), obj.GetName(), err)
				}

			case watch.Deleted:
				if err := deleteCalicoPolicy(client, obj); err != nil {
					log.Printf("Delete failed %s/%s: %v", obj.GetNamespace(), obj.GetName(), err)
					customDenyPolicies.WithLabelValues(obj.GetNamespace(), "delete_failed").Inc()
				} else {
					customDenyPolicies.WithLabelValues(obj.GetNamespace(), "deleted").Inc()
				}
			}
		}
		watcher.Stop()
	}
}

// handleCustomDeny processes a CustomDeny resource and creates corresponding Calico NetworkPolicy
func handleCustomDeny(client dynamic.Interface, obj *unstructured.Unstructured) error {
	spec, found, err := unstructured.NestedMap(obj.Object, "spec")
	if err != nil || !found {
		return fmt.Errorf("spec not found in CustomDeny")
	}

	sourceNamespace, _, _ := unstructured.NestedString(spec, "sourceNamespace")
	sourceLabels, _, _ := unstructured.NestedStringMap(spec, "sourceLabels")
	targetLabels, _, _ := unstructured.NestedStringMap(spec, "targetLabels")

	policyName := fmt.Sprintf("deny-%s-%s", obj.GetNamespace(), obj.GetName())
	
	policy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "projectcalico.org/v3",
			"kind":       "NetworkPolicy",
			"metadata": map[string]interface{}{
				"name": policyName,
				"namespace": obj.GetNamespace(),
			},
			"spec": map[string]interface{}{
				"order": 100,
				"selector": buildSelector(targetLabels),
				"types": []string {"Ingress"},
				"ingress": []interface{}{
					map[string]interface{}{
						"action": "Deny",
						"source": map[string]interface{}{
							"selector": buildSelector(sourceLabels),
      						"namespaceSelector": "projectcalico.org/name == '" + sourceNamespace + "'",
						},
					},
				},
			},
		},
	}

	gvr := schema.GroupVersionResource{
		Group:    "projectcalico.org",
		Version:  "v3",
		Resource: "networkpolicies",
	}

	// Try to create new policy first
	_, err = client.Resource(gvr).Namespace(obj.GetNamespace()).Create(context.Background(), policy, v1.CreateOptions{})
	if err != nil {
		// If create fails, try patch instead of update
		patchData := map[string]interface{}{
			"spec": policy.Object["spec"],
		}
		patchBytes, _ := json.Marshal(patchData)
		_, err = client.Resource(gvr).Namespace(obj.GetNamespace()).Patch(context.Background(), policyName, types.MergePatchType, patchBytes, v1.PatchOptions{})
	}

	return err
}

func deleteCalicoPolicy(client dynamic.Interface, obj *unstructured.Unstructured) error {
	policyName := fmt.Sprintf("deny-%s-%s", obj.GetNamespace(), obj.GetName())

	gvr := schema.GroupVersionResource{
		Group:    "projectcalico.org",
		Version:  "v3",
		Resource: "networkpolicies",
	}

	return client.Resource(gvr).Namespace(obj.GetNamespace()).Delete(context.Background(), policyName, v1.DeleteOptions{})
}

// buildSelector creates a Calico selector from namespaces and labels
func buildSelector(labels map[string]string) string {
	var selectors []string

	for k, v := range labels {
		selectors = append(selectors, fmt.Sprintf("%s == '%s'", k, v))
	}

	if len(selectors) == 0 {
		return "all()"
	}

	result := selectors[0]
	for _, sel := range selectors[1:] {
		result += " && " + sel
	}

	return result
}
