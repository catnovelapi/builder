package builder

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

var (
	plainTextType   = "text/plain; charset=utf-8"
	jsonContentType = "application/json"
	formContentType = "application/x-www-form-urlencoded"

	jsonCheck = regexp.MustCompile(`(?i:(application|text)/(.*json.*)(;|$))`)
	xmlCheck  = regexp.MustCompile(`(?i:(application|text)/(.*xml.*)(;|$))`)

	bufPool = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

func acquireBuffer() *bytes.Buffer {
	return bufPool.Get().(*bytes.Buffer)
}

func releaseBuffer(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset()
		bufPool.Put(buf)
	}
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package Helper methods
//_______________________________________________________________________

// IsStringEmpty method tells whether given string is empty or not
func IsStringEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// DetectContentType method is used to figure out `Request.Body` content type for request header
func DetectContentType(body interface{}) string {
	contentType := plainTextType
	kind := kindOf(body)
	switch kind {
	case reflect.Struct, reflect.Map:
		contentType = jsonContentType
	case reflect.String:
		contentType = plainTextType
	default:
		if b, ok := body.([]byte); ok {
			contentType = http.DetectContentType(b)
		} else if kind == reflect.Slice {
			contentType = jsonContentType
		}
	}

	return contentType
}

// IsJSONType method is to check JSON content type or not
func IsJSONType(ct string) bool {
	return jsonCheck.MatchString(ct)
}

// IsXMLType method is to check XML content type or not
func IsXMLType(ct string) bool {
	return xmlCheck.MatchString(ct)
}

// way to disable the HTML escape as opt-in
func jsonMarshal(r *Request) (*bytes.Buffer, error) {
	data, err := r.client.JSONMarshal(r.Body)
	if err != nil {
		return nil, err
	}
	r.bodyBytes = data
	buf := acquireBuffer()
	_, _ = buf.Write(data)
	return buf, nil
}
func parseRequestBody(r *Request) error {
	switch {
	case r.GetFormDataEncode() != "": // Handling Form Data
		r.bodyBuf = acquireBuffer()
		r.bodyBuf.WriteString(r.GetFormDataEncode())
		r.SetHeaderContentType(formContentType)
	case r.Body != nil: // Handling Request body
		contentType := r.GetHeaderContentType()
		if IsStringEmpty(contentType) {
			contentType = DetectContentType(r.Body)
			r.SetHeaderContentType(contentType)
		}
		if err := handleRequestBody(r); err != nil {
			return err
		}
	}
	return nil
}

func handleRequestBody(r *Request) error {
	releaseBuffer(r.bodyBuf)
	r.bodyBuf = nil

	switch body := r.Body.(type) {
	case io.Reader:
		r.bodyBuf = acquireBuffer()
		if _, err := r.bodyBuf.ReadFrom(body); err != nil {
			return err
		}
		r.Body = nil

	case []byte:
		r.bodyBytes = body
	case string:
		r.bodyBytes = []byte(body)
	default:
		contentType := r.GetHeaderContentType()
		kind := kindOf(r.Body)
		var err error
		if IsJSONType(contentType) && (kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice) {
			r.bodyBuf, err = jsonMarshal(r)
		} else if IsXMLType(contentType) && (kind == reflect.Struct) {
			r.bodyBytes, err = r.client.XMLMarshal(r.Body)
		}
		if err != nil {
			return err
		}
	}

	if r.bodyBytes == nil && r.bodyBuf == nil {
		return errors.New("unsupported 'Body' type/value")
	}

	// []byte into Buffer
	if r.bodyBytes != nil && r.bodyBuf == nil {
		r.bodyBuf = acquireBuffer()
		_, _ = r.bodyBuf.Write(r.bodyBytes)
	}

	return nil
}

func typeOf(i interface{}) reflect.Type {
	return indirect(valueOf(i)).Type()
}

func valueOf(i interface{}) reflect.Value {
	return reflect.ValueOf(i)
}

func indirect(v reflect.Value) reflect.Value {
	return reflect.Indirect(v)
}

func kindOf(v interface{}) reflect.Kind {
	return typeOf(v).Kind()
}
