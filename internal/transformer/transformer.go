package transformer

import "errors"

var ErrNilTransformer = errors.New("transformer is nil")

type Transformer interface {
	Transform(input string) (string, error)
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
