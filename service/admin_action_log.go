package service

import (
	"github.com/lejianwen/rustdesk-api/v2/model"
	"gorm.io/gorm"
)

type AdminActionLogService struct {
}

func (s *AdminActionLogService) List(page, pageSize uint, where func(tx *gorm.DB)) (res *model.AdminActionLogList) {
	res = &model.AdminActionLogList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	tx := DB.Model(&model.AdminActionLog{})
	if where != nil {
		where(tx)
	}
	tx.Count(&res.Total)
	tx.Scopes(Paginate(page, pageSize))
	tx.Find(&res.AdminActionLogs)
	return
}

func (s *AdminActionLogService) Create(u *model.AdminActionLog) error {
	return DB.Create(u).Error
}

func (s *AdminActionLogService) InfoById(id uint) (res *model.AdminActionLog) {
	res = &model.AdminActionLog{}
	DB.Where("id = ?", id).First(res)
	return
}

func (s *AdminActionLogService) Delete(u *model.AdminActionLog) error {
	return DB.Delete(u).Error
}

func (s *AdminActionLogService) BatchDelete(ids []uint) error {
	return DB.Where("id in (?)", ids).Delete(&model.AdminActionLog{}).Error
}
