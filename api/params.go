package api

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

func ParseParams(r *http.Request, in any) error {
	inElem := reflect.ValueOf(in).Elem()
	inType := inElem.Type()

	for i := 0; i < inType.NumField(); i++ {
		field := ParamsStructField(inType.Field(i))
		options, err := field.Options()

		if err != nil {
			return err
		}

		var value string

		if options.Source.Is(Path) {
			value = r.PathValue(options.Key)
		} else if options.Source.Is(Header) {
			value = r.Header.Get(options.Key)
		} else if options.Source.Is(Form) {
			value = r.PostFormValue(options.Key)
		} else {
			value = r.URL.Query().Get(options.Key)
		}

		if len(value) == 0 {
			continue
		}

		fieldValue := inElem.Field(i)

		switch kind := fieldValue.Kind(); kind {
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Int, reflect.Int64:
			intValue, err := strconv.ParseInt(value, 10, 64)

			if err != nil {
				return fmt.Errorf("invalid field value format: %w", err)
			}

			fieldValue.SetInt(intValue)
		case reflect.Float32:
			floatValue, err := strconv.ParseFloat(value, 32)

			if err != nil {
				return fmt.Errorf("invalid float value: %w", err)
			}

			fieldValue.SetFloat(floatValue)
		case reflect.Float64:
			floatValue, err := strconv.ParseFloat(value, 64)

			if err != nil {
				return fmt.Errorf("invalid float value: %w", err)
			}

			fieldValue.SetFloat(floatValue)
		case reflect.Bool:
			var boolValue bool

			switch value {
			case "true":
				boolValue = true
			case "false":
				boolValue = false
			default:
				return fmt.Errorf("invalid bool value: %s", value)
			}

			fieldValue.SetBool(boolValue)
		default:
			return fmt.Errorf("unsupported field kind %s", kind)
		}
	}

	return ApplyTransforms(in)
}

// Options contains options parsed from tag value
type Options struct {
	Key    string
	Source Source
}

type Source string

func (s Source) Is(sources ...Source) bool {
	return slices.Contains(sources, s)
}

func (s Source) String() string {
	return string(s)
}

func (s Source) Validate() error {
	if s.Is(Path, Query, Header, Form) {
		return nil
	}

	return fmt.Errorf("source %s should be one of %s", s.String(), strings.Join(AvailableParamsSources, ","))
}

const (
	Path   Source = "path"
	Query  Source = "query"
	Header Source = "header"
	Form   Source = "form"
)

var (
	AvailableParamsSources = []string{
		Path.String(),
		Query.String(),
		Header.String(),
		Form.String(),
	}
)

// ApplyTransforms normalizes the string fields of a struct according to their
// `transform` tag (a comma-separated list of transforms applied in order),
// e.g. `transform:"lowercase"`. It runs after a field has been populated, so it
// works for values bound from request params as well as from a decoded JSON
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
	if slices.Contains(AvailableParamsTransforms, t.String()) {
		return nil
	}

	return fmt.Errorf("transform %s should be one of %s", t.String(), strings.Join(AvailableParamsTransforms, ","))
}

const (
	Lowercase Transform = "lowercase"
	Uppercase Transform = "uppercase"
)

var (
	AvailableParamsTransforms = []string{
		Lowercase.String(),
		Uppercase.String(),
	}
)

// ParamsStructField wraps reflect.StructField
type ParamsStructField reflect.StructField

// Key returns parameter key for the specific struct field
func (f ParamsStructField) Options() (Options, error) {
	value, ok := f.Tag.Lookup("params")

	if !ok {
		return Options{
			Key:    strings.ToLower(f.Name),
			Source: Query,
		}, nil
	}

	parts := strings.Split(value, ",")

	if len(parts) == 1 {
		return Options{
			Key:    parts[0],
			Source: Query,
		}, nil
	}

	source := Source(parts[1])

	if err := source.Validate(); err != nil {
		return Options{}, fmt.Errorf("failed to parse tag value for the field %s: %w", f.Name, err)
	}

	return Options{
		Key:    parts[0],
		Source: source,
	}, nil
}
