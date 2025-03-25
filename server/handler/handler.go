package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/a-h/templ"
	"github.com/gopl-dev/server/app"
)

var (
	ErrIDParamMissingFromRequest = app.ErrBadRequest("ID param missing from request")
	ErrIDParamMustBeInt64        = app.ErrBadRequest("ID param must be positive int64")
)

func StatusHandler(w http.ResponseWriter, req *http.Request) {
	conf := app.Config()
	jsonOK(w, map[string]any{
		"env":     conf.App.Env,
		"version": conf.App.Version,
		"time":    time.Now(),
	})
}

type Request struct {
	Request  *http.Request
	Response http.ResponseWriter
	aborted  bool
}

func NewRequest(r *http.Request, w http.ResponseWriter) *Request {
	return &Request{
		Request:  r,
		Response: w,
	}
}

func handleJSON(w http.ResponseWriter, r *http.Request, body any) *Request {
	h := NewRequest(r, w)

	err := bindJSON(r, body)
	if err != nil {
		h.Abort(err)
	}

	if v, ok := body.(Sanitizer); ok {
		v.Sanitize()
	}

	if v, ok := body.(Validator); ok {
		err := app.NewInputError()
		v.Validate(&err)
		if err.Has() {
			h.Abort(err)
			return h
		}
	}

	if v, ok := body.(ValidateSchemaProvider); ok {
		err = app.Validate(v.ValidationSchema(), body)
		if err != nil {
			h.Abort(err)
			return h
		}
	}

	return h
}

func (h *Request) bindJSON(body any) *Request {
	return handleJSON(h.Response, h.Request, body)
}

func handleQueryRequest(w http.ResponseWriter, r *http.Request, body any) *Request {
	h := &Request{
		Request:  r,
		Response: w,
	}

	err := bindQuery(r, body)
	if err != nil {
		h.Abort(err)
	}

	if v, ok := body.(Sanitizer); ok {
		v.Sanitize()
	}

	if v, ok := body.(Validator); ok {
		err := app.NewInputError()
		v.Validate(&err)
		if err.Has() {
			h.Abort(err)
		}
	}

	return h
}

func (h *Request) MapHeaders(to any) {
	val := reflect.ValueOf(to).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("h")
		if tag == "" {
			continue
		}
		headerName := tag
		if commaIndex := strings.Index(tag, ","); commaIndex != -1 {
			headerName = tag[:commaIndex]
		}

		if value, ok := h.Request.Header[http.CanonicalHeaderKey(headerName)]; ok && len(value) > 0 {
			fieldVal := val.Field(i)
			switch fieldVal.Kind() {
			// TODO handle other types (when needed)
			case reflect.Int:
				intVal, err := strconv.Atoi(value[0])
				if err == nil {
					fieldVal.SetInt(int64(intVal))
				}
			case reflect.String:
				fieldVal.SetString(value[0])
			}
		}
	}
}

func (h *Request) jsonOK(body any) {
	jsonOK(h.Response, body)
}

func (h *Request) Abort(err error) {
	h.aborted = true
	abort(h.Response, err)
}

func (h *Request) Aborted() bool {
	return h.aborted
}

func bindJSON(r *http.Request, obj any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()

	err := decoder.Decode(obj)
	if err != nil {
		err = fmt.Errorf("decode JSON: %s", err.Error())
	}

	return err
}

func writeJSON(w http.ResponseWriter, body any, status int) (err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err = json.NewEncoder(w).Encode(body)
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

func jsonOK(w http.ResponseWriter, body any) {
	err := writeJSON(w, body, http.StatusOK)
	if err != nil {
		log.Println(err)
	}
}

func render(ctx context.Context, w http.ResponseWriter, t templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	err := t.Render(ctx, w)
	if err != nil {
		log.Println(err)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Println(err)
		}
	}
}

type Validator interface {
	Validate(err *app.InputError)
}

type ValidateSchemaProvider interface {
	ValidationSchema() z.Schema
}

type Sanitizer interface {
	Sanitize()
}

//func bindID(c *gin.Context, id *int64, paramNameOpt ...string) (ok bool) {
//	name := "id"
//	if len(paramNameOpt) == 1 {
//		name = paramNameOpt[0]
//	}
//
//	val := c.Param(name)
//	if val == "" {
//		abort(c, ErrIDParamMissingFromRequest)
//		return
//	}
//
//	intVal, err := strconv.ParseInt(val, 10, 64)
//	if err != nil {
//		abort(c, ErrIDParamMustBeInt64)
//		return
//	}
//
//	if intVal < 1 {
//		abort(c, ErrIDParamMustBeInt64)
//		return
//	}
//
//	*id = intVal
//	return true
//}

func copyRequestBody(r *http.Request) (body []byte, err error) {
	body, err = io.ReadAll(r.Body)
	if err != nil {
		return
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return
}

func abort(w http.ResponseWriter, err error) {
	resp := Error{
		Code: app.CodeInternal,
	}

	switch e := err.(type) {
	case app.Error:
		resp.Code = e.Code
		resp.Error = e.Error()
	case app.InputError:
		resp.Code = app.CodeUnprocessable
		resp.InputErrors = e
	default:
		resp.Error = err.Error()
	}

	log.Println(resp.Error)

	if resp.Code >= app.CodeInternal {
		log.Println(string(debug.Stack()))
	}

	err = writeJSON(w, resp, resp.Code)
	if err != nil {
		log.Println(err)
	}
}

type Error struct {
	Code        int               `json:"code"`
	Error       string            `json:"error,omitempty"`
	Errors      []string          `json:"errors,omitempty"`
	InputErrors map[string]string `json:"input_errors,omitempty"`
}

func bindQuery(r *http.Request, to any) (err error) {
	query := r.URL.Query()

	val := reflect.ValueOf(to).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("q")
		if tag == "" {
			continue
		}

		if value, ok := query[tag]; ok && len(value) > 0 {
			fieldVal := val.Field(i)
			// TODO handle more types (when needed)
			switch fieldVal.Kind() {
			case reflect.Int:
				intVal, err := strconv.Atoi(value[0])
				if err == nil {
					fieldVal.SetInt(int64(intVal))
				}
			case reflect.String:
				fieldVal.SetString(value[0])
			}
		}
	}

	return nil
}
