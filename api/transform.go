package api

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// ApplyTransforms normalizes the string fields of a struct according to their
// `transform` tag (a comma-separated list of transforms applied in order),
// e.g. `transform:"lowercase"`. It runs after a field has been populated, so it
// is reused for values bound from request params as well as from a decoded JSON
// body. Fields without a transform tag are left untouched; a transform tag on a
// non-string field is an error.
func ApplyTransforms(in any) error {
	inElem := reflect.ValueOf(in).Elem()
	inType := inElem.Type()

	for i := 0; i < inType.NumField(); i++ {
		field := inType.Field(i)

		tag, ok := field.Tag.Lookup("transform")

		if !ok {
			continue
		}

		transforms, err := ParseTransforms(field.Name, tag)

		if err != nil {
			return err
		}

		fieldValue := inElem.Field(i)

		if fieldValue.Kind() != reflect.String {
			return fmt.Errorf("transform tag on field %s requires a string field, got %s", field.Name, fieldValue.Kind())
		}

		value := fieldValue.String()

		for _, transform := range transforms {
			value = transform.Apply(value)
		}

		fieldValue.SetString(value)
	}

	return nil
}

// ParseTransforms parses a `transform` tag value into a validated slice of
// transforms, preserving their declared order.
func ParseTransforms(fieldName string, tag string) ([]Transform, error) {
	parts := strings.Split(tag, ",")
	transforms := make([]Transform, 0, len(parts))

	for _, part := range parts {
		transform := Transform(strings.TrimSpace(part))

		if err := transform.Validate(); err != nil {
			return nil, fmt.Errorf("failed to parse transform tag for the field %s: %w", fieldName, err)
		}

		transforms = append(transforms, transform)
	}

	return transforms, nil
}

// Transform is a normalization applied to a string field's value after it is
// populated, whether bound from a request param or decoded from a JSON body.
type Transform string

func (t Transform) String() string {
	return string(t)
}

func (t Transform) Apply(value string) string {
	switch t {
	case Lowercase:
		return strings.ToLower(value)
	case Uppercase:
		return strings.ToUpper(value)
	default:
		return value
	}
}

func (t Transform) Validate() error {
	if slices.Contains(AvailableTransforms, t.String()) {
		return nil
	}

	return fmt.Errorf("transform %s should be one of %s", t.String(), strings.Join(AvailableTransforms, ","))
}

const (
	Lowercase Transform = "lowercase"
	Uppercase Transform = "uppercase"
)

var (
	AvailableTransforms = []string{
		Lowercase.String(),
		Uppercase.String(),
	}
)
