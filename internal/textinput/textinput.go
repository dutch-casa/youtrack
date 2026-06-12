package textinput

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type Source struct {
	Name       string
	Literal    string
	LiteralSet bool
	File       string
	Stdin      bool
}

func Resolve(source Source, stdin io.Reader) (string, bool, error) {
	provided := 0
	if source.LiteralSet {
		provided++
	}
	if source.File != "" {
		provided++
	}
	if source.Stdin {
		provided++
	}
	if provided == 0 {
		return "", false, nil
	}
	if provided > 1 {
		return "", false, fmt.Errorf("%s accepts only one text source", source.Name)
	}
	if source.LiteralSet {
		return source.Literal, true, nil
	}
	if source.File != "" {
		data, err := os.ReadFile(source.File)
		if err != nil {
			return "", false, fmt.Errorf("read %s file: %w", source.Name, err)
		}
		return string(data), true, nil
	}
	if stdin == nil {
		return "", false, errors.New("stdin is not available")
	}
	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", false, fmt.Errorf("read %s from stdin: %w", source.Name, err)
	}
	return string(data), true, nil
}
