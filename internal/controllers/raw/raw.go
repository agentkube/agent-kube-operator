package raw

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	client     client.Client
	scheme     *runtime.Scheme
	restConfig *rest.Config
}

func NewController(client client.Client, scheme *runtime.Scheme, config *rest.Config) *Controller {
	return &Controller{
		client:     client,
		scheme:     scheme,
		restConfig: config,
	}
}

func (c *Controller) GetRawResource(ctx context.Context, path string) (map[string]interface{}, error) {
	// Create a copy of the config
	config := rest.CopyConfig(c.restConfig)

	// Parse the path to determine API path and GroupVersion
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path format")
	}

	// Determine if this is an API or APIs path
	if parts[0] == "apis" {
		// Format: /apis/GROUP/VERSION/...
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid apis path format")
		}
		config.APIPath = "apis"
		config.GroupVersion = &schema.GroupVersion{
			Group:   parts[1],
			Version: parts[2],
		}
	} else if parts[0] == "api" {
		// Format: /api/VERSION/...
		config.APIPath = "api"
		config.GroupVersion = &schema.GroupVersion{
			Group:   "",
			Version: parts[1],
		}
	} else {
		return nil, fmt.Errorf("path must start with /api or /apis")
	}

	// Set content type
	config.ContentType = runtime.ContentTypeJSON
	config.AcceptContentTypes = runtime.ContentTypeJSON
	config.NegotiatedSerializer = runtime.NewSimpleNegotiatedSerializer(runtime.SerializerInfo{})

	// Create REST client
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %v", err)
	}

	// Make the raw request
	result := restClient.Get().AbsPath(path).Do(ctx)
	if err := result.Error(); err != nil {
		return nil, fmt.Errorf("failed to get resource: %v", err)
	}

	// Read the raw response
	raw, err := result.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse the response into a map
	var response map[string]interface{}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return response, nil
}
