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
