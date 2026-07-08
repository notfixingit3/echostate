package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/audit"
	"github.com/notfixingit3/echostate/internal/auth"
	"github.com/notfixingit3/echostate/internal/models"
)

func TestListAuditEvents_AdminOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	d := setupTestDB(t)
	h := &Handler{db: d, audit: audit.New(d)}

	require.NoError(t, h.audit.Record(t.Context(), audit.Entry{
		ActorName: "Admin",
		ActorRole: auth.RoleAdmin,
		Action:    audit.ActionSettingsUpdate,
	}))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/audit?page=1&limit=20", nil)

	h.listAuditEvents(c)
	require.Equal(t, http.StatusOK, w.Code)

	var resp models.PaginatedResponse[models.AuditEvent]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, 1, resp.Total)
	require.Equal(t, audit.ActionSettingsUpdate, resp.Data[0].Action)
}
