package provider

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"gopkg.in/yaml.v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	groupSchema "k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func newClient() (dynamic.Interface, error) {
	config, _ := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	return dynamic.NewForConfig(config)
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"kubernetes_crd": crdResource(),
		},
	}
}

func crdResource() *schema.Resource {
	return &schema.Resource{
		Create: onCreate,
		Read:   onRead,
		Update: onUpdate,
		Delete: onDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"api_version": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"kind": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"namespace": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"spec": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func onCreate(d *schema.ResourceData, m interface{}) error {
	crdAPIVersion := d.Get("api_version").(string)
	crdType := strings.Split(crdAPIVersion, "/")

	crdKind := d.Get("kind").(string)
	crdName := d.Get("name").(string)
	crdNamespace := d.Get("namespace").(string)

	log.Printf("Executing create for %s - %s - %s", crdAPIVersion, crdKind, crdName)
	crdResource := groupSchema.GroupVersionResource{
		Group:    crdType[0],
		Version:  crdType[1],
		Resource: strings.ToLower(crdKind) + "s",
	}

	var crdSpec map[string]interface{}

	crdSpecString := d.Get("spec").(string)
	err := yaml.Unmarshal([]byte(crdSpecString), &crdSpec)
	if err != nil {
		log.Printf("Spec string %s", crdSpecString)
		log.Print(err)
		panic("Failure decoding CRD spec")
	}

	sanitized, err := stringMapize(crdSpec)
	if err != nil {
		return err
	}

	obj := map[string]interface{}{
		"apiVersion": crdAPIVersion,
		"kind":       crdKind,
		"metadata": map[string]string{
			"name":      crdName,
			"namespace": crdNamespace,
		},
		"spec": sanitized,
	}

	unstructuredObj := unstructured.Unstructured{}

	unstructuredObj.SetUnstructuredContent(obj)

	log.Printf("Content is '%T'", unstructuredObj.Object["spec"].(map[string]interface{})["sidecarInjector"])

	// serialized, _ := json.Marshal(unstructuredObj)
	_, err = json.Marshal(unstructuredObj.Object)
	if err != nil {
		log.Fatal(err)
	}

	client, _ := newClient()

	// log.Print(obj)
	log.Printf("Calling k8s api for %s", crdName)
	_, err = client.Resource(crdResource).Namespace(crdNamespace).Create(&unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		panic(err)
		return err
	}

	log.Printf("Created %s", crdName)

	// return onRead(d, m)
	return nil
}

func stringMapize(i interface{}) (interface{}, error) {
	var err error
	switch t := i.(type) {
	case (map[interface{}]interface{}):
		m := make(map[string]interface{}, len(t))
		log.Printf("Data -> %s", t)

		for k, v := range t {
			str, ok := k.(string)
			if !ok {
				return nil, errors.New("map had non-string keys")
			}
			m[str], err = stringMapize(v)
			if err != nil {
				return nil, err
			}
		}
	case (map[string]interface{}):
		for k, v := range t {
			t[k], err = stringMapize(v)
			if err != nil {
				return nil, err
			}
		}
		// todo: more cases
	}
	return i, nil
}

func onRead(d *schema.ResourceData, m interface{}) error {
	// return do("read", d, m)
	crdAPIVersion := d.Get("api_version").(string)
	crdType := strings.Split(crdAPIVersion, "/")

	crdKind := d.Get("kind").(string)
	crdName := d.Get("name").(string)
	crdNamespace := d.Get("namespace").(string)
	// crdSpec := d.Get("spec").(string)

	crdResource := groupSchema.GroupVersionResource{
		Group:    crdType[0],
		Version:  crdType[1],
		Resource: strings.ToLower(crdKind) + "s",
	}

	client, _ := newClient()

	crdInstance, _ := client.Resource(crdResource).Namespace(crdNamespace).Get(crdName, metav1.GetOptions{})
	m = crdInstance

	return nil
}

func onUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func onDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
