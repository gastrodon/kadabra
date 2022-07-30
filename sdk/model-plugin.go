package sdk

import "github.com/hashicorp/hcl/v2/hcldec"

type kind int

const (
	PRODUCER    kind = 0b0001
	CONSUMER    kind = 0b0010
	TRANSFORMER kind = 0b0100
)

type Resource struct {
	Kinds              kind
	Name               string
	Spec               hcldec.ObjectSpec
	ProvideProducer    ProducerProvider
	ProvideConsumer    ConsumerProvider
	ProvideTransformer TransformerProvider
}

type Plugin struct {
	Name      string
	Resources []*Resource
}
