package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"
)

type Body interface {
	ToReader() (io.Reader, error)
}

type Parameter map[string]string

func (p Parameter) ToReader() (io.Reader, error) {
	return bytes.NewBuffer([]byte(p.Encode())), nil
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

func (j JSON) ToReader() (io.Reader, error) {
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

func (m JSON) GetJSON(key string) (JSON, error) {
	value, err := m.Get(key)
	if err != nil {
		return nil, err
	}

	if j, ok := value.(JSON); ok {
		return j, nil
	}

	return nil, fmt.Errorf("invalid type of field %s", key)
}

func (m JSON) GetInt(key string) (int, error) {
	value, err := m.Get(key)
	if err != nil {
		return 0, err
	}

	if i, ok := value.(int); ok {
		return i, nil
	}

	return 0, fmt.Errorf("invalid type of field %s", key)
}

func (m JSON) GetBool(key string) (bool, error) {
	value, err := m.Get(key)
	if err != nil {
		return false, err
	}

	if b, ok := value.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("invalid type of field %s", key)
}

func (m JSON) GetString(key string) (string, error) {
	value, err := m.Get(key)
	if err != nil {
		return "", err
	}

	if s, ok := value.(string); ok {
		return s, nil
	}

	return "", fmt.Errorf("invalid type of field %s", key)
}

func (m JSON) Get(key string) (any, error) {
	key, subKey, found := strings.Cut(key, ".")

	value, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("not found field %s", key)
	}

	if found {
		if mvalue, ok := value.(map[string]any); ok {
			return JSON(mvalue).Get(subKey)
		}
		return nil, fmt.Errorf("invalid type of field %s", key)
	}

	return value, nil
}

func readerToJSON(body io.Reader) (JSON, error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	result := JSON{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
