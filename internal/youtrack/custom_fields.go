package youtrack

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (f *CustomField) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.Name = strings.TrimSpace(raw.Name)
	f.Value = CustomFieldDisplayValue(raw.Value)
	return nil
}

func (i Issue) CustomFieldValue(names ...string) string {
	return CustomFieldValue(i.CustomFields, names...)
}

func CustomFieldValue(fields []CustomField, names ...string) string {
	for _, field := range fields {
		for _, name := range names {
			if strings.EqualFold(field.Name, name) {
				return field.Value
			}
		}
	}
	return ""
}

func CustomFieldDisplayValue(data json.RawMessage) string {
	if len(data) == 0 || string(data) == "null" {
		return ""
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		return displayText(text)
	}

	var object struct {
		Presentation string `json:"presentation"`
		Name         string `json:"name"`
		Login        string `json:"login"`
		Text         string `json:"text"`
	}
	if err := json.Unmarshal(data, &object); err == nil {
		return displayText(firstNonEmpty(object.Presentation, object.Name, object.Login, object.Text))
	}

	var objects []struct {
		Presentation string `json:"presentation"`
		Name         string `json:"name"`
		Login        string `json:"login"`
		Text         string `json:"text"`
	}
	if err := json.Unmarshal(data, &objects); err == nil {
		values := make([]string, 0, len(objects))
		for _, object := range objects {
			value := displayText(firstNonEmpty(object.Presentation, object.Name, object.Login, object.Text))
			if value != "" {
				values = append(values, value)
			}
		}
		return strings.Join(values, ", ")
	}

	var primitive any
	if err := json.Unmarshal(data, &primitive); err == nil {
		return displayText(fmt.Sprint(primitive))
	}
	return ""
}

func displayText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
