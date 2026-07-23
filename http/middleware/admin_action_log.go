package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

// sensitiveFieldPattern matches JSON keys whose values should be redacted
// before being stored in the admin action log detail column.
var sensitiveFieldPattern = regexp.MustCompile(`(?i)"(password|pwd|secret|token|api-token)"\s*:\s*"[^"]*"`)

const adminActionLogDetailMaxLen = 4000

type bodyCaptureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AdminActionLog records every mutating (non-GET) request made against the
// /api/admin API, independent of the existing end-user connection/file
// audit logs, so admin console changes are individually traceable.
func AdminActionLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		writer := &bodyCaptureWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = writer

		c.Next()

		operatorId := uint(0)
		operatorName := ""
		if u := service.AllService.UserService.CurUser(c); u != nil {
			operatorId = u.Id
			operatorName = u.Username
		}

		success := false
		var resp struct {
			Code int `json:"code"`
		}
		if err := json.Unmarshal(writer.body.Bytes(), &resp); err == nil {
			success = resp.Code == 0
		}

		module, action := adminActionLogModuleAndAction(c.FullPath())

		detail := sensitiveFieldPattern.ReplaceAllString(string(reqBody), `"$1":"***"`)
		if len(detail) > adminActionLogDetailMaxLen {
			detail = detail[:adminActionLogDetailMaxLen]
		}

		log := &model.AdminActionLog{
			OperatorId:   operatorId,
			OperatorName: operatorName,
			Module:       module,
			Action:       action,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			Ip:           c.ClientIP(),
			StatusCode:   c.Writer.Status(),
			Success:      success,
			Detail:       detail,
		}
		_ = service.AllService.AdminActionLogService.Create(log)
	}
}

// adminActionLogModuleAndAction derives a module/action pair from a route's
// registered path pattern, e.g. "/api/admin/user/edit/:id" -> ("user", "edit").
func adminActionLogModuleAndAction(fullPath string) (module string, action string) {
	trimmed := strings.TrimPrefix(fullPath, "/api/admin/")
	parts := make([]string, 0, 4)
	for _, p := range strings.Split(trimmed, "/") {
		if p == "" || strings.HasPrefix(p, ":") {
			continue
		}
		parts = append(parts, p)
	}
	if len(parts) == 0 {
		return "unknown", "unknown"
	}
	if len(parts) == 1 {
		return parts[0], "default"
	}
	return parts[0], strings.Join(parts[1:], "_")
}
