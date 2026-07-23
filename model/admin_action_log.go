package model

// AdminActionLog records mutating actions performed by admins/users through
// the admin API, for accountability/traceability separate from the
// end-user connection (AuditConn) and file-transfer (AuditFile) audit logs.
type AdminActionLog struct {
	IdModel
	OperatorId   uint   `json:"operator_id" gorm:"default:0;not null;index"`
	OperatorName string `json:"operator_name" gorm:"default:'';not null;"`
	Module       string `json:"module" gorm:"default:'';not null;index"`
	Action       string `json:"action" gorm:"default:'';not null;"`
	Method       string `json:"method" gorm:"default:'';not null;"`
	Path         string `json:"path" gorm:"default:'';not null;"`
	Ip           string `json:"ip" gorm:"default:'';not null;"`
	StatusCode   int    `json:"status_code" gorm:"default:0;not null;"`
	Success      bool   `json:"success" gorm:"default:0;not null;index"`
	Detail       string `json:"detail" gorm:"type:text;"`
	TimeModel
}

type AdminActionLogList struct {
	AdminActionLogs []*AdminActionLog `json:"list"`
	Pagination
}
