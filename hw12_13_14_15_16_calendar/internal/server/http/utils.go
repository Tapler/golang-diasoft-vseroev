package internalhttp

import (
	"strings"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// generateUUID генерирует новый UUID v4.
func generateUUID() string {
	return uuid.New().String()
}

// parseUUID парсит строку UUID в openapi_types.UUID.
func parseUUID(id string) openapi_types.UUID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		// Если не удалось распарсить, возвращаем нулевой UUID
		return openapi_types.UUID{}
	}
	return parsed
}

// extractPathParam извлекает параметр из пути URL.
func extractPathParam(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	param := strings.TrimPrefix(path, prefix)
	// Убираем возможный trailing slash
	param = strings.TrimSuffix(param, "/")
	return param
}
