package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"strings"
	"time"
)

var ErrNotFoundKey = errors.New("not found key")

type Raw []byte

func (r Raw) ToReader() (io.Reader, error) {
	return bytes.NewBuffer(r), nil
}

type Parameter map[string]string

func (p Parameter) ToReader() (io.Reader, string, error) {
	return bytes.NewBuffer([]byte(p.Encode())), "application/x-www-form-urlencoded", nil
}

func (p Parameter) Encode() string {
	var parameters []string
	for key, value := range p {
		parameters = append(parameters, key+"="+PercentEncode(value))
	}
	sort.Strings(parameters)
	return strings.Join(parameters, "&")
}

type JSON map[string]any

type Array []JSON

func (j JSON) ToReader() (io.Reader, string, error) {
	b, err := json.Marshal(j)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewBuffer(b), "application/json", nil
}

func (m JSON) GetJSON(key string) (JSON, error) {
	value, err := m.Get(key)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	}

	if j, ok := value.(map[string]any); ok {
		return j, nil
	}

	return nil, fmt.Errorf("invalid type of field %s (%T)", key, value)
}

func (m JSON) GetInt(key string) (int, error) {
	value, err := m.Get(key)
	if err != nil {
		return 0, err
	}

	if value == nil {
		return 0, nil
	}

	switch t := value.(type) {
	case int:
		return t, nil

	case float64:
		if t == float64(int(t)) {
			return int(t), nil
		}

		return 0, fmt.Errorf("invalid type of field %s (actually float64)", key)
	}

	return 0, fmt.Errorf("invalid type of field %s (%T)", key, value)
}

func (m JSON) GetBool(key string) (bool, error) {
	value, err := m.Get(key)
	if err != nil {
		return false, err
	}

	if value == nil {
		return false, nil
	}

	if b, ok := value.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("invalid type of field %s (%T)", key, value)
}

func (m JSON) GetArray(key string) (Array, error) {
	value, err := m.Get(key)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	}

	if a, ok := value.([]any); ok {
		array := Array{}
		for i := range a {
			if m, ok := a[i].(map[string]any); !ok {
				return nil, fmt.Errorf("invalid array element %T", a[i])
			} else {
				array = append(array, m)
			}
		}

		return array, nil
	}

	return nil, fmt.Errorf("invalid type of field %s", key)
}

func (m JSON) GetString(key string) (string, error) {
	value, err := m.Get(key)
	if err != nil {
		return "", err
	}

	if value == nil {
		return "", nil
	}

	if s, ok := value.(string); ok {
		return s, nil
	}

	return "", fmt.Errorf("invalid type of field %s (%T)", key, value)
}

func (m JSON) GetTime(key string, layout string) (time.Time, error) {
	value, err := m.Get(key)
	if err != nil {
		return time.Time{}, err
	}

	if value == nil {
		return time.Time{}, nil
	}

	if s, ok := value.(string); ok {
		result, err := time.Parse(layout, s)
		if err != nil {
			return time.Time{}, err
		}

		return result, nil
	}

	return time.Time{}, fmt.Errorf("invalid type of field %s (%T)", key, value)
}

func (m JSON) Get(key string) (any, error) {
	key, subKey, found := strings.Cut(key, ".")

	value, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFoundKey, key)
	}

	if found {
		if mvalue, ok := value.(map[string]any); ok {
			return JSON(mvalue).Get(subKey)
		}
		return nil, fmt.Errorf("invalid type of field %s (%T)", key, value)
	}

	return value, nil
}

func bytesToJSON(body []byte) (JSON, error) {
	result := JSON{}
	err := json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func bytesToArray(body []byte) (Array, error) {
	result := Array{}
	err := json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type IFormData interface {
	Body
	ContentType() string
}

type FormDataFile struct {
	Content io.Reader
	Name    string
}

type FormData struct {
	Texts map[string]string
	Files map[string]FormDataFile
}

func (d FormData) ToReader() (io.Reader, string, error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, data := range d.Files {
		var fw io.Writer
		if x, ok := data.Content.(io.Closer); ok {
			defer x.Close()
		}

		var err error
		if data.Name != "" {
			if fw, err = w.CreateFormFile(key, data.Name); err != nil {
				return nil, "", err
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				return nil, "", err
			}
		}

		if _, err = io.Copy(fw, data.Content); err != nil {
			return nil, "", err
		}
	}

	for key, data := range d.Texts {
		fw, err := w.CreateFormField(key)
		if err != nil {
			return nil, "", err
		}

		if _, err := fw.Write([]byte(data)); err != nil {
			return nil, "", err
		}
	}

	w.Close()
	return &b, w.FormDataContentType(), nil
}

type Response struct {
	Code    int
	Header  http.Header
	Body    any
	RawBody []byte
}
