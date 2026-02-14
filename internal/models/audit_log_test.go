package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuditLog_TableName(t *testing.T) {
	auditLog := AuditLog{}

	tableName := auditLog.TableName()

	assert.Equal(t, "audit_logs", tableName)
}
