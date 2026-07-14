package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func reflectFields(t *testing.T, in any) map[string]reflect.StructField {
	t.Helper()

	inType := reflect.TypeOf(in)
	fields := make(map[string]reflect.StructField, inType.NumField())

	for i := range inType.NumField() {
		field := inType.Field(i)
		fields[field.Name] = field
	}

	return fields
}

func TestParseParamsSources(t *testing.T) {
	type params struct {
		ID    string `params:"id,path"`
		Token string `params:"authorization,header"`
		Limit int    `params:"limit,query"`
		Name  string // no tag: lowercased field name, query source
	}

	r := httptest.NewRequest(http.MethodGet, "/items/abc123?limit=10&name=widget", nil)
	r.SetPathValue("id", "abc123")
	r.Header.Set("Authorization", "Bearer secret")

	var p params
	require.NoError(t, ParseParams(r, &p))

	assert.Equal(t, "abc123", p.ID)
	assert.Equal(t, "Bearer secret", p.Token)
	assert.Equal(t, 10, p.Limit)
	assert.Equal(t, "widget", p.Name)
}

func TestParseParamsKinds(t *testing.T) {
	type params struct {
		Str     string  `params:"str"`
		Int     int     `params:"int"`
		Int64   int64   `params:"int64"`
		Float32 float32 `params:"float32"`
		Float64 float64 `params:"float64"`
		Bool    bool    `params:"bool"`
	}

	r := httptest.NewRequest(http.MethodGet,
		"/?str=hello&int=42&int64=9000000000&float32=1.5&float64=2.25&bool=true", nil)

	var p params
	require.NoError(t, ParseParams(r, &p))

	assert.Equal(t, "hello", p.Str)
	assert.Equal(t, 42, p.Int)
	assert.Equal(t, int64(9000000000), p.Int64)
	assert.Equal(t, float32(1.5), p.Float32)
	assert.Equal(t, 2.25, p.Float64)
	assert.True(t, p.Bool)
}

// An absent value must leave the field as-is rather than erroring or zeroing it,
// which is what lets readers/binders be layered over an already-populated struct.
func TestParseParamsSkipsEmptyValues(t *testing.T) {
	type params struct {
		Name string `params:"name"`
		Age  int    `params:"age"`
	}

	r := httptest.NewRequest(http.MethodGet, "/?name=set", nil)

	p := params{Age: 7}
	require.NoError(t, ParseParams(r, &p))

	assert.Equal(t, "set", p.Name)
	assert.Equal(t, 7, p.Age) // untouched, not reset to 0
}

func TestParseParamsInvalidValues(t *testing.T) {
	t.Run("rejects a non-numeric int", func(t *testing.T) {
		var p struct {
			Limit int `params:"limit"`
		}

		r := httptest.NewRequest(http.MethodGet, "/?limit=abc", nil)
		assert.Error(t, ParseParams(r, &p))
	})

	t.Run("rejects a non-numeric float", func(t *testing.T) {
		var p struct {
			Ratio float64 `params:"ratio"`
		}

		r := httptest.NewRequest(http.MethodGet, "/?ratio=abc", nil)
		assert.Error(t, ParseParams(r, &p))
	})

	t.Run("rejects a bool that is not true/false", func(t *testing.T) {
		var p struct {
			Flag bool `params:"flag"`
		}

		r := httptest.NewRequest(http.MethodGet, "/?flag=yes", nil)

		err := ParseParams(r, &p)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid bool value")
	})

	t.Run("rejects an unsupported field kind", func(t *testing.T) {
		var p struct {
			Tags []string `params:"tags"`
		}

		r := httptest.NewRequest(http.MethodGet, "/?tags=a", nil)

		err := ParseParams(r, &p)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported field kind")
	})

	t.Run("rejects an unknown source in the tag", func(t *testing.T) {
		var p struct {
			Name string `params:"name,body"`
		}

		r := httptest.NewRequest(http.MethodGet, "/?name=x", nil)
		assert.Error(t, ParseParams(r, &p))
	})
}

func TestSourceValidate(t *testing.T) {
	for _, source := range []Source{Path, Query, Header} {
		assert.NoError(t, source.Validate(), "%s should be valid", source)
	}

	assert.Error(t, Source("body").Validate())
}

func TestSourceIs(t *testing.T) {
	assert.True(t, Path.Is(Path, Query))
	assert.False(t, Header.Is(Path, Query))
}

func TestParamsStructFieldOptions(t *testing.T) {
	type params struct {
		Untagged  string
		KeyOnly   string `params:"key"`
		KeyAndSrc string `params:"vin,path"`
		BadSource string `params:"x,body"`
	}

	fields := reflectFields(t, params{})

	t.Run("defaults to lowercased name and query source", func(t *testing.T) {
		options, err := ParamsStructField(fields["Untagged"]).Options()
		require.NoError(t, err)
		assert.Equal(t, "untagged", options.Key)
		assert.Equal(t, Query, options.Source)
	})

	t.Run("defaults to query source when only a key is given", func(t *testing.T) {
		options, err := ParamsStructField(fields["KeyOnly"]).Options()
		require.NoError(t, err)
		assert.Equal(t, "key", options.Key)
		assert.Equal(t, Query, options.Source)
	})

	t.Run("parses key and source", func(t *testing.T) {
		options, err := ParamsStructField(fields["KeyAndSrc"]).Options()
		require.NoError(t, err)
		assert.Equal(t, "vin", options.Key)
		assert.Equal(t, Path, options.Source)
	})

	t.Run("fails on an unknown source", func(t *testing.T) {
		_, err := ParamsStructField(fields["BadSource"]).Options()
		assert.Error(t, err)
	})
}
