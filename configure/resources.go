package configure

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

const (
	NAMESPACE_PRODUCE   = "produce"
	NAMESPACE_CONSUME   = "consume"
	NAMESPACE_TRANSFORM = "transform"
	NAMESPACE_VALUE     = "value"
)

func name(namespace string, resource *Resource) string {
	return strings.Join([]string{namespace, resource.Kind, resource.Name}, ".")
}

func loadResourceSlice(namespace string, resources []*Resource) (cty.Value, error) {
	kinds := make(map[string]map[string]cty.Value, 0)
	for _, resource := range resources {
		if _, ok := kinds[resource.Kind]; !ok {
			kinds[resource.Kind] = make(map[string]cty.Value, 0)
		}

		kinds[resource.Kind][resource.Name] = cty.StringVal(name(namespace, resource))
	}

	refs := make(map[string]cty.Value, len(kinds))
	for name, kindMap := range kinds {
		refs[name] = cty.ObjectVal(kindMap)
	}

	return cty.ObjectVal(refs), nil
}

func loadResources(filename string, literal []byte, context *hcl.EvalContext) (*Resources, error) {
	file, diags := hclparse.NewParser().ParseHCL(literal, filename)
	if diags != nil {
		return nil, diags
	}

	resources := new(Resources)
	gohcl.DecodeBody(file.Body, context, resources)
	return resources, nil
}

func loadResourcesContext(filename string, literal []byte) (*hcl.EvalContext, error) {
	if resources, err := loadResources(filename, literal, nil); err != nil {
		return nil, err
	} else {
		produce, err := loadResourceSlice(NAMESPACE_PRODUCE, resources.Producers)
		if err != nil {
			return nil, err
		}

		consume, err := loadResourceSlice(NAMESPACE_CONSUME, resources.Consumers)
		if err != nil {
			return nil, err
		}

		transform, err := loadResourceSlice(NAMESPACE_TRANSFORM, resources.Transformers)
		if err != nil {
			return nil, err
		}

		return &hcl.EvalContext{
			Variables: map[string]cty.Value{
				NAMESPACE_PRODUCE:   produce,
				NAMESPACE_CONSUME:   consume,
				NAMESPACE_TRANSFORM: transform,
			},
		}, nil
	}
}

func loadResorceLookup(filename string, literal []byte, context *hcl.EvalContext) (map[string]*Resource, error) {
	if resources, err := loadResources(filename, literal, context); err != nil {
		return nil, err
	} else {
		lookupSize := len(resources.Producers) + len(resources.Consumers) + len(resources.Transformers)
		lookup := make(map[string]*Resource, lookupSize)

		for _, each := range resources.Producers {
			lookup[name(NAMESPACE_PRODUCE, each)] = each
		}

		for _, each := range resources.Consumers {
			lookup[name(NAMESPACE_CONSUME, each)] = each
		}

		for _, each := range resources.Transformers {
			lookup[name(NAMESPACE_TRANSFORM, each)] = each
		}

		return lookup, nil
	}
}
