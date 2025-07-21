package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

/* Takes the error from json.Unmarshal and provides a contextual error */
func FormatJsonError(contents []byte, err error) error {
	jsonErr, ok := err.(*json.SyntaxError)
	if !ok {
		return err
	}

	if strings.Count(string(contents), "\n") > 1 {
		offset := 0
		problemPart := ""
		for _, line := range strings.Split(string(contents), "\n") {
			if jsonErr.Offset < int64(offset+len(line)) {
				problemPart = strings.TrimRight(line, "\n")
				break
			}
			offset += len(line)
		}
		err = fmt.Errorf("%w: error near (offset %d):\n\t'%s'", err, jsonErr.Offset-int64(offset), problemPart)
	} else {
		offset := 10
		problemPart := contents[jsonErr.Offset-int64(offset) : jsonErr.Offset+int64(offset)]
		err = fmt.Errorf("%w: error near (offset %d):\n\t'%s'", err, offset, problemPart)
	}

	return err
}
