// internal/controllers/resources/controller.go
package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	client          client.Client
	scheme          *runtime.Scheme
	restConfig      *rest.Config
	discoveryClient *discovery.DiscoveryClient
}

type APIResource struct {
	Group      string
	Version    string
	Resource   string
	Kind       string
	Namespaced bool
}

func NewController(client client.Client, scheme *runtime.Scheme, config *rest.Config) (*Controller, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %v", err)
	}

	return &Controller{
		client:          client,
		scheme:          scheme,
		restConfig:      config,
		discoveryClient: discoveryClient,
	}, nil
}

func (c *Controller) GetNamespacedResource(ctx context.Context, namespace, group, version, resourceType, resourceName string) (map[string]interface{}, error) {
	if group == "core" {
		group = ""
	}

	config := rest.CopyConfig(c.restConfig)

	if group == "" {
		config.APIPath = "api"
	} else {
		config.APIPath = "apis"
	}

	config.GroupVersion = &schema.GroupVersion{
		Group:   group,
		Version: version,
	}
	config.ContentType = runtime.ContentTypeJSON
	config.AcceptContentTypes = runtime.ContentTypeJSON
	config.NegotiatedSerializer = runtime.NewSimpleNegotiatedSerializer(runtime.SerializerInfo{})

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %v", err)
	}

	var path string
	if group == "" {
		path = fmt.Sprintf("/api/%s/namespaces/%s/%s/%s",
			version, namespace, resourceType, resourceName)
	} else {
		path = fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s/%s",
			group, version, namespace, resourceType, resourceName)
	}

	result := restClient.Get().AbsPath(path).Do(ctx)
	if err := result.Error(); err != nil {
		return nil, fmt.Errorf("failed to get resource: %v", err)
	}

	raw, err := result.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return response, nil
}

func (c *Controller) GetResource(ctx context.Context, namespace, group, version, resourceType, resourceName string) (map[string]interface{}, error) {
	if group == "core" {
		group = ""
	}

	resources, err := c.ListAPIResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check resource type: %v", err)
	}

	var isNamespaced bool
	resourceFound := false
	for _, r := range resources {
		if r.Group == group && r.Version == version && r.Resource == resourceType {
			isNamespaced = r.Namespaced
			resourceFound = true
			break
		}
	}

	if !resourceFound {
		return nil, fmt.Errorf("resource type not found: %s", resourceType)
	}

	config := rest.CopyConfig(c.restConfig)
	if group == "" {
		config.APIPath = "api"
	} else {
		config.APIPath = "apis"
	}

	config.GroupVersion = &schema.GroupVersion{
		Group:   group,
		Version: version,
	}
	config.ContentType = runtime.ContentTypeJSON
	config.AcceptContentTypes = runtime.ContentTypeJSON
	config.NegotiatedSerializer = runtime.NewSimpleNegotiatedSerializer(runtime.SerializerInfo{})

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %v", err)
	}

	var path string
	if isNamespaced && namespace != "" {
		if group == "" {
			path = fmt.Sprintf("/api/%s/namespaces/%s/%s/%s",
				version, namespace, resourceType, resourceName)
		} else {
			path = fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s/%s",
				group, version, namespace, resourceType, resourceName)
		}
	} else {
		if group == "" {
			path = fmt.Sprintf("/api/%s/%s/%s",
				version, resourceType, resourceName)
		} else {
			path = fmt.Sprintf("/apis/%s/%s/%s/%s",
				group, version, resourceType, resourceName)
		}
	}

	result := restClient.Get().AbsPath(path).Do(ctx)
	if err := result.Error(); err != nil {
		return nil, fmt.Errorf("failed to get resource: %v", err)
	}

	raw, err := result.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return response, nil
}

// ListAPIResources returns all available API resources in the cluster
func (c *Controller) ListAPIResources(ctx context.Context) ([]APIResource, error) {
	_, apiResourceLists, err := c.discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get server resources: %v", err)
	}

	var resources []APIResource
	for _, apiResourceList := range apiResourceLists {
		gv := strings.Split(apiResourceList.GroupVersion, "/")
		var group, version string
		if len(gv) == 1 {
			group = ""
			version = gv[0]
		} else {
			group = gv[0]
			version = gv[1]
		}

		for _, resource := range apiResourceList.APIResources {
			// Skip subresources
			if strings.Contains(resource.Name, "/") {
				continue
			}

			resources = append(resources, APIResource{
				Group:      group,
				Version:    version,
				Resource:   resource.Name,
				Kind:       resource.Kind,
				Namespaced: resource.Namespaced,
			})
		}
	}

	return resources, nil
}

func (c *Controller) ApplyResource(ctx context.Context, namespace, group, version, resourceType, resourceName string, content map[string]interface{}) error {
	if group == "core" {
		group = ""
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(c.restConfig)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Create GroupVersionResource
	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resourceType,
	}

	// Convert content to unstructured
	unstructuredObj := &unstructured.Unstructured{
		Object: content,
	}

	// Update the resource
	if namespace != "" {
		// For namespaced resources
		_, err = dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, unstructuredObj, metav1.UpdateOptions{})
	} else {
		// For cluster-scoped resources
		_, err = dynamicClient.Resource(gvr).Update(ctx, unstructuredObj, metav1.UpdateOptions{})
	}

	if err != nil {
		return fmt.Errorf("failed to update resource: %v", err)
	}

	return nil
}

func (c *Controller) DeleteResource(
	ctx context.Context,
	namespace string,
	group string,
	version string,
	resourceType string,
	resourceName string,
) error {
	// Get the GVK for the resource
	gvk := schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    resourceType,
	}

	// Create a new unstructured object
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	obj.SetName(resourceName)
	if namespace != "" {
		obj.SetNamespace(namespace)
	}

	// Delete the resource
	err := c.client.Delete(ctx, obj)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}
