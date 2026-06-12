package tui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type workItemDraft struct {
	Minutes int
	Text    string
}

func parseWorkItemInput(input string) (workItemDraft, error) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return workItemDraft{}, errors.New("work item duration is required")
	}

	var minutes int
	var consumed int
	for consumed < len(fields) {
		part, ok, err := parseDurationPart(fields[consumed])
		if err != nil {
			return workItemDraft{}, err
		}
		if !ok {
			break
		}
		minutes += part
		consumed++
	}
	if consumed == 0 {
		return workItemDraft{}, errors.New("work item duration must start with a value like 45m or 1h")
	}
	if minutes <= 0 {
		return workItemDraft{}, errors.New("work item duration must be greater than zero")
	}

	return workItemDraft{
		Minutes: minutes,
		Text:    strings.Join(fields[consumed:], " "),
	}, nil
}

func parseDurationPart(value string) (int, bool, error) {
	var minutes int
	consumedAny := false
	for i := 0; i < len(value); {
		start := i
		for i < len(value) && value[i] >= '0' && value[i] <= '9' {
			i++
		}
		if start == i {
			if !consumedAny {
				return 0, false, nil
			}
			return 0, false, fmt.Errorf("invalid work item duration %q", value)
		}
		if i >= len(value) {
			return 0, false, fmt.Errorf("invalid work item duration %q", value)
		}

		amount, err := strconv.Atoi(value[start:i])
		if err != nil {
			return 0, false, fmt.Errorf("invalid work item duration %q", value)
		}
		switch value[i] {
		case 'h', 'H':
			minutes += amount * 60
		case 'm', 'M':
			minutes += amount
		default:
			return 0, false, fmt.Errorf("invalid work item duration %q", value)
		}
		i++
		consumedAny = true
	}
	return minutes, true, nil
}
