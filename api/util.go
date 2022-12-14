package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

// parseJSONBody reads a single JSON-encoded value from the request body and
// stores it in the value pointed to by params.  Any remaining content in the
// request body is discarded.
//
// If an error is encountered a 400 bad request response is written and an
// error returned.
func parseJSONBody(params any, rw http.ResponseWriter, r *http.Request) error {
	defer io.Copy(ioutil.Discard, r.Body) //nolint:errcheck
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		BadRequest(rw, r, err, "error parsing JSON body")
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
