package validation

import (
	"regexp"

	"github.com/instaclustr/operator/pkg/models"
)

func ValidateISODate(date string) (bool, error) {
	return regexp.Match(models.ISODateFormatRegExp, []byte(date))
}

func Contains(str string, s []string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
