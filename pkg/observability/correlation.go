package observability

import (
	"github.com/google/uuid"
)

func GetOrGenerateCorrelationID(headerValue string) string {
	if headerValue != "" {
		return headerValue
	}
	return uuid.New().String()
}
