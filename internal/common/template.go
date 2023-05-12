package common

import (
	"bytes"
	"text/template"
)

func ExecuteTemplate(source string, data any) (string, error) {
	tmpl, err := template.New("template").Parse(source)
	if err != nil {
		return "", err
	}

	buffer := bytes.NewBuffer(nil)
	err = tmpl.Execute(buffer, data)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
