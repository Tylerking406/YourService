package web

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dimfeld/httptreemux"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
	translations "gopkg.in/go-playground/validator.v9/translations/en"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

// validate holds the settings and caches for validating request struct values.
var validate *validator.Validate

// translator is a cache of locale and translation information.
var translator *ut.UniversalTranslator

func init() {

	// Instantiate the validator for use.
	validate = validator.New()

	// Instantiate the english locale for the validator library.
	enLocale := en.New()

	// Create a value using English as the fallback locale (first argument).
	// Provide one or more arguments for additional supported locales.
	translator = ut.New(enLocale, enLocale)

	// Register the english error messages for validation errors.
	lang, _ := translator.GetTranslator("en")
	_ = translations.RegisterDefaultTranslations(validate, lang)

	// Use JSON tag names for errors instead of Go struct names.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Params returns the web call parameters from the request.
func Params(r *http.Request) map[string]string {
	return httptreemux.ContextParams(r.Context())
}

// Decode reads the body of an HTTP request looking for a JSON document. The
// body is decoded into the provided value.
//
// If the provided value is a struct then it is checked for validation tags.
func Decode(r *http.Request, dst interface{}) error {

	// Decode body into struct interface{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return NewRequestError(err, http.StatusBadRequest)
	}

	// Sanitize all string fields dynamically
	v := reflect.ValueOf(dst).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i).Interface()
		switch field.(type) {
		case string:
			// Sanitize field
			fieldStr := field.(string)
			p := bluemonday.UGCPolicy()
			fieldStr = p.Sanitize(fieldStr)
			v.Field(i).SetString(fieldStr)
		}
	}

	// Validate decoded struct
	if err := validate.Struct(dst); err != nil {

		// Use a type assertion to get the real error value.
		verrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		// lang controls the language of the error messages. You could look at
		// the Accept-Language header if you intend to support multiple
		// languages.
		lang, _ := translator.GetTranslator("en")

		var fields []FieldError
		for _, verror := range verrors {
			field := FieldError{
				Field: verror.Field(),
				Error: verror.Translate(lang),
			}
			fields = append(fields, field)
		}

		return &Error{
			Err:        errors.New("field validation error"),
			StatusCode: http.StatusBadRequest,
			Fields:     fields,
		}
	}

	return nil
}

// DoRequest handles sending a basic HTTP request to any URL
// and get a response as []byte
func DoRequest(url string, headers map[string]string, httpMethod string, data interface{}) ([]byte, error) {

	// Create the http request
	// Encode the data and set its content type in the case of an http POST
	var req *http.Request
	var err error
	if httpMethod == http.MethodPost {
		req, err = http.NewRequest(httpMethod, url, Encode(data))
	} else if httpMethod == http.MethodGet {
		req, err = http.NewRequest(httpMethod, url, nil)
	} else {
		err = errors.Errorf("unrecognized httpMethod [%v]", httpMethod)
	}
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Attempt to do http request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// If not StatusOK, return error
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.Errorf("failed to send [%v] request to [%v]", httpMethod, url)
	}

	// Return success
	bytes, err := ioutil.ReadAll(resp.Body)
	return bytes, nil

}

// Encode is a convenience function for testing http requests.
// WARNING, there is no error handling for Marshaling
func Encode(data interface{}) *bytes.Reader {
	jsonData, _ := json.Marshal(data)
	return bytes.NewReader(jsonData)
}

// GetParam gets a specified query param from the specified request
func GetParam(r *http.Request, param string) string {
	_ = r.ParseForm()
	return r.Form.Get(param)
}

// GetPathParam gets a specified path param using the treemux Context
func GetPathParam(ctx context.Context, param string) (string, error) {
	// Get params from context as map
	params := httptreemux.ContextParams(ctx)

	// Get specified param from map
	val, ok := params[param]
	if !ok {
		// Path param not found
		return "", errors.Errorf("/%v/:%v not provided!", param, param)
	}

	// Support the default namespace
	if val == "default" {
		val = ""
	}

	// Return success
	return val, nil
}
