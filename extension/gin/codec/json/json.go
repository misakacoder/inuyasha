package json

import (
	"github.com/gin-gonic/gin/codec/json"
	"github.com/json-iterator/go"
	"io"
	"unicode"
)

func init() {
	lowerCamelCaseJSON := &LowerCamelCaseJSON{
		API: jsoniter.ConfigCompatibleWithStandardLibrary,
	}
	lowerCamelCaseJSON.API.RegisterExtension(lowerCamelCaseJSON)
	json.API = lowerCamelCaseJSON
}

type LowerCamelCaseJSON struct {
	jsoniter.API
	jsoniter.DummyExtension
}

func (json *LowerCamelCaseJSON) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	for _, binding := range structDescriptor.Fields {
		structField := binding.Field
		if structField.Tag().Get("json") == "" {
			name := []rune(structField.Name())
			name[0] = unicode.ToLower(name[0])
			names := []string{string(name)}
			binding.FromNames = names
			binding.ToNames = names
		}
	}
}

func (json *LowerCamelCaseJSON) Marshal(v any) ([]byte, error) {
	return json.API.Marshal(v)
}

func (json *LowerCamelCaseJSON) Unmarshal(data []byte, v any) error {
	return json.API.Unmarshal(data, v)
}

func (json *LowerCamelCaseJSON) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.API.MarshalIndent(v, prefix, indent)
}

func (json *LowerCamelCaseJSON) NewEncoder(writer io.Writer) json.Encoder {
	return json.API.NewEncoder(writer)
}

func (json *LowerCamelCaseJSON) NewDecoder(reader io.Reader) json.Decoder {
	return json.API.NewDecoder(reader)
}
