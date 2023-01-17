package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	validate *validator.Validate
	trans    ut.Translator
)

func init() {
	en := en.New()
	uni := ut.New(en, en)
	validate = validator.New()
	trans, _ = uni.GetTranslator("en")
	err := en_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		panic(err)
	}

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		// skip if tag key says it should be ignored
		if name == "-" {
			return ""
		}
		return name
	})

	err = validate.RegisterValidation("notblank", validators.NotBlank)
	if err != nil {
		panic(err)
	}

	err = validate.RegisterTranslation("notblank", trans, func(ut ut.Translator) error {
		return ut.Add("notblank", "{0} cannot be blank", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("notblank", fe.Field())
		return t
	})
	if err != nil {
		panic(err)
	}
}

// parseJSONBody reads a single JSON-encoded value from the request body and
// stores it in the value pointed to by params.  Any remaining content in the
// request body is discarded.
//
// If an error is encountered either a 400 bad request or a 422 unprocessable
// entity response is written and an error returned.
func parseJSONBody(params any, rw http.ResponseWriter, r *http.Request) error {
	defer io.Copy(ioutil.Discard, r.Body) //nolint:errcheck
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		BadRequest(rw, r, err, "error parsing JSON body")
		return err
	}
	err = validate.Struct(params)
	if err != nil {
		errs := make([]*ErrorObject, 0)
		for _, err := range err.(validator.ValidationErrors) {
			eo := ErrorObject{
				Status: http.StatusUnprocessableEntity,
				Title:  err.Tag(),
				Detail: err.Translate(trans),
				Source: err.Field(),
			}
			errs = append(errs, &eo)
		}

		resp := ErrorsPayload{
			Status: http.StatusUnprocessableEntity,
			Errors: errs,
		}
		renderJSON(resp, http.StatusUnprocessableEntity, rw)
		return err

	}
	return nil
}

// renderJSON writes a JSON representation of `body` to `rw`, sets the content
// type header to application/json and sets the status header to `status`.
//
// If an error is encountered a 500 internal server error is written.
func renderJSON(body any, status int, rw http.ResponseWriter) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(body); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	rw.Write(buf.Bytes()) //nolint:errcheck
}
