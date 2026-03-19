package app

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
)

func ResolveRunInput(runtimeOpts RuntimeOptions, cfg Config, stdin *bufio.Reader) (Input, error) {
	if runtimeOpts.NonInteractive {
		in := runtimeOpts.RunInput
		if strings.TrimSpace(in.URL) == "" {
			return Input{}, errors.New("--url is required in non-interactive mode")
		}
		if strings.TrimSpace(in.Name) == "" {
			return Input{}, errors.New("--name is required in non-interactive mode")
		}
		urlValue, err := normalizeAndValidateURL(in.URL)
		if err != nil {
			return Input{}, err
		}
		in.URL = urlValue

		if runtimeOpts.SkipIcon {
			in.IconURL = ""
		} else if strings.TrimSpace(in.IconURL) == "" {
			in.IconURL = in.URL
		} else {
			iconValue, err := normalizeAndValidateIconLocation(in.IconURL)
			if err != nil {
				return Input{}, fmt.Errorf("invalid icon url: %w", err)
			}
			in.IconURL = iconValue
			in.IconURLExplicit = runtimeOpts.RunInputExplicit.IconURL
		}

		return in, nil
	}

	return AskInput(stdin, cfg)
}
