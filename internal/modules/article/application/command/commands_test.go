package command_test

import (
	"testing"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/article/application/command"
	"github.com/stretchr/testify/assert"
)

func TestArticleCommands_Validation(t *testing.T) {
	t.Run("CreateArticleCommand", func(t *testing.T) {
		assert.NoError(t, command.CreateArticleCommand{UserID: "u1", Title: "T1", Content: "C1"}.Validate())
		assert.Error(t, command.CreateArticleCommand{UserID: "", Title: "T1", Content: "C1"}.Validate())
		assert.Error(t, command.CreateArticleCommand{UserID: "u1", Title: "", Content: "C1"}.Validate())
		assert.Error(t, command.CreateArticleCommand{UserID: "u1", Title: "T1", Content: ""}.Validate())
	})

	t.Run("UpdateArticleCommand", func(t *testing.T) {
		assert.NoError(t, command.UpdateArticleCommand{ID: "a1"}.Validate())
		assert.Error(t, command.UpdateArticleCommand{ID: ""}.Validate())
	})

	t.Run("DeleteArticleCommand", func(t *testing.T) {
		assert.NoError(t, command.DeleteArticleCommand{ID: "a1"}.Validate())
		assert.Error(t, command.DeleteArticleCommand{ID: ""}.Validate())
	})
}
