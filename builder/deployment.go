package builder

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func DeployBuilder(repoName string) error {
	config, err := rest.InClusterConfig()

	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		return err
	}

	pod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind": "Pod",
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("builder-%s", repoName),
			},
			"spec": map[string]interface{}{
				"containers": map[string]interface{}{
					"name": fmt.Sprintf("builder-%s", repoName),
					"image": "gcr.io/kaniko-project/executor:latest",
					"args": map[string]interface{}{
					},
				},
			},
		},
	}

	return nil
}
