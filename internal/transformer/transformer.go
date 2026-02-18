package transformer

import (
	"errors"
	"log"
)

var ErrNilTransformer = errors.New("transformer is nil")

type Transformer interface {
	Transform(input string) (string, error)
	Name() string
}

func DefaultTransformers() []Transformer {
	return []Transformer{
		NewImageURLPruner(false),
		NewClassPruner(),
		NewStylePruner(),
	}
}

func ApplyAll(input string, transformers ...Transformer) (string, error) {
	output := input
	for _, t := range transformers {
		if t == nil {
			return "", ErrNilTransformer
		}
		var err error
		output, err = t.Transform(output)
		if err != nil {
			return "", err
		}
	}
	return output, nil
}

func FromNames(names ...string) []Transformer {
	var transformers []Transformer
	for _, name := range names {
		switch name {
		case "image-url-pruner":
			transformers = append(transformers, NewImageURLPruner(false))
		case "ImageURLPruner":
			transformers = append(transformers, NewImageURLPruner(false))
		case "class-pruner":
			transformers = append(transformers, NewClassPruner())
		case "ClassPruner":
			transformers = append(transformers, NewClassPruner())
		case "style-pruner":
			transformers = append(transformers, NewStylePruner())
		case "StylePruner":
			transformers = append(transformers, NewStylePruner())
		default:
			log.Printf("warning: unknown transformer name '%s'", name)
		}
	}
	return transformers
}
