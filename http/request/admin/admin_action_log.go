package admin

type AdminActionLogQuery struct {
	OperatorName string `form:"operator_name"`
	Module       string `form:"module"`
	Success      string `form:"success"`
	PageQuery
}

type AdminActionLogIds struct {
	Ids []uint `json:"ids" validate:"required"`
}
