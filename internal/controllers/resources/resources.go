// internal/controllers/resources/controller.go
package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
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

	// Build the request path
	var path string
	if group == "" {
		path = fmt.Sprintf("/api/%s/namespaces/%s/%s/%s",
			version, namespace, resourceType, resourceName)
	} else {
		path = fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s/%s",
			group, version, namespace, resourceType, resourceName)
	}

	// Make the request
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
