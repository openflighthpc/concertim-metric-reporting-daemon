package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
}

func Test_MinimumDocumentIsValid(t *testing.T) {
	tests := []struct {
		name string
		doc  string
	}{
		{
			name: "valid string metric",
			doc:  `{"type": "string", "name": "foo", "value": "bar", "slope": "both", "ttl": 60}`,
		},
		{
			name: "valid empty string metric",
			doc:  `{"type": "string", "name": "foo", "value": "", "slope": "both", "ttl": 60}`,
		},
		{
			name: "valid positive int32 metric",
			doc:  `{"type": "int32", "name": "foo", "value": 20, "slope": "both", "ttl": 60}`,
		},
		{
			name: "valid zero int32 metric",
			doc:  `{"type": "int32", "name": "foo", "value": 0, "slope": "both", "ttl": 60}`,
		},
		{
			name: "valid negative int32 metric",
			doc:  `{"type": "int32", "name": "foo", "value": -20, "slope": "both", "ttl": 60}`,
		},
		{
			name: "valid positive double metric",
			doc:  `{"type": "double", "name": "foo", "value": 20.1234, "slope": "both", "ttl": 60}`,
		},
		{
			name: "valid zero double metric",
			doc:  `{"type": "double", "name": "foo", "value": 0, "slope": "both", "ttl": 60}`,
		},
		{
			name: "valid negative double metric",
			doc:  `{"type": "double", "name": "foo", "value": -20.1234, "slope": "both", "ttl": 60}`,
		},
		{
			name: "valid positive uint32 metric",
			doc:  `{"type": "uint32", "name": "foo", "value": 20, "slope": "both", "ttl": 60}`,
		},
		{
			name: "valid zero uint32 metric",
			doc:  `{"type": "uint32", "name": "foo", "value": 0, "slope": "both", "ttl": 60}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			putMetric := &putMetricRequest{}
			err := parseAndValidate(putMetric, bytes.NewBufferString(tt.doc))
			assert.NoError(t, err)
		})
	}
}

func Test_SimpleValidations(t *testing.T) {
	tests := []struct {
		name   string
		tags   []string
		field  string
		detail string
		doc    string
	}{
		{
			name:   "name is required",
			tags:   []string{"required"},
			field:  "name",
			detail: "name is a required field",
			doc:    `{"type": "int32", "name": "", "value": 1, "units": " ", "slope": "both", "ttl": 60}`,
		},
		{
			name:   "name cannot be blank",
			tags:   []string{"notblank"},
			field:  "name",
			detail: "name cannot be blank",
			doc:    `{"type": "int32", "name": " ", "value": 1, "units": " ", "slope": "both", "ttl": 60}`,
		},
		{
			name:   "type is required",
			tags:   []string{"required"},
			field:  "type",
			detail: "type is a required field",
			doc:    `{"type": "", "name": "foo", "value": 1, "units": " ", "slope": "both", "ttl": 60}`,
		},
		{
			name:   "type must be oneof ...",
			tags:   []string{"oneof"},
			field:  "type",
			detail: "type must be one of [string int8 uint8 int16 uint16 int32 uint32 float double]",
			doc:    `{"type": "invalid", "name": "foo", "value": 1, "units": " ", "slope": "both", "ttl": 60}`,
		},
		{
			name:   "slope is required",
			tags:   []string{"required"},
			field:  "slope",
			detail: "slope is a required field",
			doc:    `{"type": "int32", "name": "foo", "value": 1, "units": " ", "slope": "", "ttl": 60}`,
		},
		{
			name:   "slope must be oneof ...",
			tags:   []string{"oneof"},
			field:  "slope",
			detail: "slope must be one of [zero positive negative both derivative]",
			doc:    `{"type": "int32", "name": "foo", "value": 1, "units": " ", "slope": "invalid", "ttl": 60}`,
		},
		{
			name:   "ttl is required",
			tags:   []string{"required"},
			field:  "ttl",
			detail: "ttl is a required field",
			doc:    `{"type": "int32", "name": "foo", "value": 1, "units": " ", "slope": "both", "ttl": 0}`,
		},
		{
			name:   "ttl must be 1 or greater",
			tags:   []string{"min"},
			field:  "ttl",
			detail: "ttl must be 1 or greater",
			doc:    `{"type": "int32", "name": "foo", "value": 1, "units": " ", "slope": "both", "ttl": -1}`,
		},
		{
			name:   "units cannot contain ...",
			tags:   []string{"min"},
			field:  "units",
			detail: "units cannot contain any of the following characters '<>'\"&'",
			doc:    `{"type": "int32", "name": "foo", "value": 1, "units": "<", "slope": "both", "ttl": 1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			putMetric := &putMetricRequest{}
			var validationErrs validator.ValidationErrors

			// Action
			err := parseAndValidate(putMetric, bytes.NewBufferString(tt.doc))

			// Assertions
			assert.Error(t, err)
			assert.ErrorAs(t, err, &validationErrs)
			assert.Len(t, validationErrs, len(tt.tags))
			for i, fieldErr := range validationErrs {
				if i < len(tt.tags)-1 {
					assert.Equal(t, tt.tags[i], fieldErr.Tag())
				}
				assert.Equal(t, tt.field, fieldErr.Field())
				assert.Equal(t, tt.detail, fieldErr.Translate(trans))
			}
		})
	}
}

func Test_InvalidValueTypesAreInvalid(t *testing.T) {
	tests := []struct {
		name string
		typ  string
		doc  string
	}{
		{
			name: "strings are not int32",
			typ:  "int32",
			doc:  `{"type": "int32", "name": "foo", "value": "a string", "units": "", "slope": "both", "ttl": 60}`,
		},
		{
			name: "floats are not int32",
			typ:  "int32",
			doc:  `{"type": "int32", "name": "foo", "value": 3.14, "units": "", "slope": "both", "ttl": 60}`,
		},
		{
			name: "strings are not doubles",
			typ:  "double",
			doc:  `{"type": "double", "name": "foo", "value": "a string", "units": "", "slope": "both", "ttl": 60}`,
		},
		{
			name: "ints are not strings",
			typ:  "string",
			doc:  `{"type": "string", "name": "foo", "value": 10, "units": "", "slope": "both", "ttl": 60}`,
		},
		{
			name: "floats are not strings",
			typ:  "string",
			doc:  `{"type": "string", "name": "foo", "value": 3.14, "units": "", "slope": "both", "ttl": 60}`,
		},
		{
			name: "negative numbers are not uint32",
			typ:  "uint32",
			doc:  `{"type": "uint32", "name": "foo", "value": -1, "units": "", "slope": "both", "ttl": 60}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			putMetric := &putMetricRequest{}
			var validationErrs validator.ValidationErrors

			// Action
			err := parseAndValidate(putMetric, bytes.NewBufferString(tt.doc))

			// Assertions
			assert.Error(t, err)
			assert.ErrorAs(t, err, &validationErrs)
			assert.Len(t, validationErrs, 1)
			for _, fieldErr := range validationErrs {
				assert.Equal(t, "validtype", fieldErr.Tag())
				assert.Equal(t, "value", fieldErr.Field())
				assert.Equal(t, fmt.Sprintf("value is not valid for type %s", tt.typ), fieldErr.Translate(trans))
			}
		})
	}
}

func parseAndValidate(params any, body io.Reader) error {
	err := json.NewDecoder(body).Decode(params)
	if err != nil {
		return err
	}
	err = validate.Struct(params)
	return err
}
