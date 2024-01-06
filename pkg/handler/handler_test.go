package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	handler := New("../../config/handler-slack.json")

	assert.NotNil(t, handler)
	assert.Equal(t, "", (*handler).Event.ID)

	value, _ := handler.Config.Get("slack").Get("token").String()
	assert.Equal(t, "xoxp-01234567890-01234567890-01234567890-abcdefghij", value)
}
