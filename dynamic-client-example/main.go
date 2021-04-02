package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"log"
	"os"
	"path/filepath"
)

type Whole struct {
	Group string `json:"group"`
	Version string `json:"version"`
	Resource string `json:"resource"`
	Object map[string]interface{} `json:"object"`
}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	namespace := "default"

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	dClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	//file, err := ioutil.ReadFile("./artifacts/deployment.yaml")
	file, err := os.Open("./artifacts/deployment.json")
	//file, err := os.Open("./artifacts/deployment.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var whole Whole

	//err = yaml.Unmarshal(file, &whole)
	err = json.NewDecoder(file).Decode(&whole)
	//err = yaml.NewDecoder(file).Decode(&whole)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(whole.Object["apiVersion"])
	fmt.Println(whole.Object["kind"])
	fmt.Println(whole.Object["metadata"])

	deploymentRes := schema.GroupVersionResource{
		Group:    whole.Group,
		Version:  whole.Version,
		Resource: whole.Resource,
	}

	deploymentObj := &unstructured.Unstructured{
		Object: whole.Object,
	}

	// -------------- Create function --------------------
	prompt()
	fmt.Println("Creating deployment...")

	result, err := dClient.Resource(deploymentRes).Namespace(namespace).Create(context.TODO(), deploymentObj, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("error from here")
		log.Fatal(err)
	}

	fmt.Printf("Created deployment %q\n", result.GetName())

	// --------------- Update function ----------------------
	prompt()
	fmt.Println("Updating deployment...")

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		resourceName, found, err := unstructured.NestedString(whole.Object, "metadata", "name")
		if err != nil || !found {
			log.Fatal(err)
		}

		result, getErr := dClient.Resource(deploymentRes).Namespace(namespace).Get(context.TODO(), resourceName, metav1.GetOptions{})
		if getErr != nil {
			log.Fatal(getErr)
		}

		// update replicas
		if err := unstructured.SetNestedField(result.Object, int64(1), "spec", "replicas"); err != nil {
			log.Fatal(err)
		}

		// extract spec containers
		containers, found, err := unstructured.NestedSlice(result.Object, "spec", "template", "spec", "containers")
		if err != nil || !found || containers == nil {
			log.Fatal(err)
		}

		// update container[0] image
		if err := unstructured.SetNestedField(containers[0].(map[string]interface{}), "nginx:1.13", "image"); err != nil {
			log.Fatal(err)
		}
		if err := unstructured.SetNestedField(result.Object, containers, "spec", "template", "spec", "containers"); err != nil {
			log.Fatal(err)
		}

		_, updateErr := dClient.Resource(deploymentRes).Namespace(namespace).Update(context.TODO(), result, metav1.UpdateOptions{})
		return updateErr
	})

	if retryErr != nil {
		log.Fatal(retryErr)
	}
	fmt.Println("Updated deployment...")


	// ---------------------- List function ------------------------
	prompt()
	fmt.Printf("listing deployments in namespace %q\n", apiv1.NamespaceDefault)

	list, err := dClient.Resource(deploymentRes).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range list.Items {
		replicas, found, err := unstructured.NestedInt64(d.Object, "spec", "replicas")
		if err != nil || !found {
			fmt.Printf("replicas not found for deployment %s: error = %s", d.GetName(), err)
			continue
		}
		fmt.Printf("* %s (%d replicas)\n", d.GetName(), replicas)
	}


	// -------------------- Delete function ---------------------
	prompt()
	fmt.Println("Deleting deployment...")
	resourceName, found, err := unstructured.NestedString(whole.Object, "metadata", "name")
	if err != nil || !found {
		log.Fatal(err)
	}
	if err := dClient.Resource(deploymentRes).Namespace(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{}); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted deployment")
}

func prompt() {
	fmt.Printf("-> press enter key to continue...")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println()
}