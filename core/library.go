package core

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/psyduck-std/sdk"
)

type Library struct {
	ProvideProducer    func(string, *hcl.EvalContext, hcl.Body) (sdk.Producer, error)
	ProvideConsumer    func(string, *hcl.EvalContext, hcl.Body) (sdk.Consumer, error)
	ProvideTransformer func(string, *hcl.EvalContext, hcl.Body) (sdk.Transformer, error)
}

func makeBodySchema(specMap sdk.SpecMap) *hcl.BodySchema {
	attributes := make([]hcl.AttributeSchema, len(specMap))

	index := 0
	for _, spec := range specMap {
		attributes[index] = hcl.AttributeSchema{
			Name:     spec.Name,
			Required: spec.Required,
		}

		index++
	}

	return &hcl.BodySchema{
		Attributes: attributes,
	}
}

func makeSpecParser(context *hcl.EvalContext, config hcl.Body) sdk.SpecParser {
	return func(spec sdk.SpecMap, target interface{}) error {
		content, _, diags := config.PartialContent(makeBodySchema(spec))
		if diags.HasErrors() {
			return diags
		}

		if diags := decodeAttributes(spec, context, content.Attributes, target); diags.HasErrors() {
			return diags
		}

		return nil
	}
}

func makeParser(providedSpecMap sdk.SpecMap, context *hcl.EvalContext, config hcl.Body) (sdk.Parser, sdk.SpecParser) {
	parser := makeSpecParser(context, config)
	return func(target interface{}) error {
		return parser(providedSpecMap, target)
	}, parser
}

func NewLibrary(plugins []*sdk.Plugin) *Library {
	size := 0
	for _, plugin := range plugins {
		size += len(plugin.Resources)
	}

	lookupResource := make(map[string]*sdk.Resource, size)
	for _, plugin := range plugins {
		for _, resource := range plugin.Resources {
			lookupResource[resource.Name] = resource
		}
	}

	return &Library{
		ProvideProducer: func(name string, context *hcl.EvalContext, config hcl.Body) (sdk.Producer, error) {
			found, ok := lookupResource[name]
			if !ok {
				return nil, fmt.Errorf("can't find resource %s", name)
			}

			if found.Kinds&sdk.PRODUCER == 0 {
				return nil, fmt.Errorf("resource %s doesn't provide a producer", name)
			}

			return found.ProvideProducer(makeParser(found.Spec, context, config))
		},
		ProvideConsumer: func(name string, context *hcl.EvalContext, config hcl.Body) (sdk.Consumer, error) {
			found, ok := lookupResource[name]
			if !ok {
				return nil, fmt.Errorf("can't find resource %s", name)
			}

			if found.Kinds&sdk.CONSUMER == 0 {
				return nil, fmt.Errorf("resource %s doesn't provide a consumer", name)
			}

			return found.ProvideConsumer(makeParser(found.Spec, context, config))
		},
		ProvideTransformer: func(name string, context *hcl.EvalContext, config hcl.Body) (sdk.Transformer, error) {
			found, ok := lookupResource[name]
			if !ok {
				return nil, fmt.Errorf("can't find resource %s", name)
			}

			if found.Kinds&sdk.TRANSFORMER == 0 {
				return nil, fmt.Errorf("resource %s doesn't provide a consumer", name)
			}

			return found.ProvideTransformer(makeParser(found.Spec, context, config))
		},
	}
}
