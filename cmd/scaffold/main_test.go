package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaffoldFeature(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "scaffold-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = ScaffoldFeature(tempDir, "item")
	require.NoError(t, err)

	expectedFiles := []string{
		filepath.Join(tempDir, "internal", "modules", "item", "domain", "entity", "item.go"),
		filepath.Join(tempDir, "internal", "modules", "item", "domain", "repository", "item_repository.go"),
		filepath.Join(tempDir, "internal", "modules", "item", "application", "command", "item_commands.go"),
		filepath.Join(tempDir, "internal", "modules", "item", "application", "query", "item_queries.go"),
		filepath.Join(tempDir, "internal", "modules", "item", "application", "dto", "item_dto.go"),
		filepath.Join(tempDir, "internal", "modules", "item", "application", "service.go"),
		filepath.Join(tempDir, "internal", "modules", "item", "infrastructure", "persistence", "inmemory_item_repository.go"),
		filepath.Join(tempDir, "internal", "modules", "item", "delivery", "http", "request", "item_request.go"),
		filepath.Join(tempDir, "internal", "modules", "item", "delivery", "http", "response", "item_response.go"),
		filepath.Join(tempDir, "internal", "modules", "item", "delivery", "http", "handler", "item_handler.go"),
		filepath.Join(tempDir, "internal", "modules", "item", "delivery", "http", "routes.go"),
	}

	for _, f := range expectedFiles {
		assert.FileExists(t, f)
	}
}
