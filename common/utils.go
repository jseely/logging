package common

import (
	"fmt"
	"strings"
)

func FormatTemplate(messageTemplate string, fields map[string]interface{}) string {
	for k, v := range fields {
		messageTemplate = strings.Replace(messageTemplate, "{"+k+"}", fmt.Sprintf("%v", v), -1)
	}
	return messageTemplate
}
