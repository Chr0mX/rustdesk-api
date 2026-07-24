package service

import (
	"github.com/lejianwen/rustdesk-api/v2/model"
	"gorm.io/gorm"
	"time"
)

type ShareRecordService struct {
}

// InfoById 根据用户id取用户信息
func (srs *ShareRecordService) InfoById(id uint) *model.ShareRecord {
	u := &model.ShareRecord{}
	DB.Where("id = ?", id).First(u)
	return u
}

// InfoByShareToken looks up a still-valid (non-expired) share record by its
// token. Expire is a TTL in seconds counted from CreatedAt (0 = never
// expires), same convention as WebClient.SharedPeer.
func (srs *ShareRecordService) InfoByShareToken(token string) *model.ShareRecord {
	sr := &model.ShareRecord{}
	if token == "" {
		return sr
	}
	DB.Where("share_token = ?", token).First(sr)
	if sr.Id == 0 {
		return sr
	}
	if sr.Expire != 0 {
		ca := time.Time(sr.CreatedAt)
		if ca.Add(time.Second * time.Duration(sr.Expire)).Before(time.Now()) {
			return &model.ShareRecord{}
		}
	}
	return sr
}

func (srs *ShareRecordService) List(page, pageSize uint, where func(tx *gorm.DB)) (res *model.ShareRecordList) {
	res = &model.ShareRecordList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	tx := DB.Model(&model.ShareRecord{})
	if where != nil {
		where(tx)
	}
	tx.Count(&res.Total)
	tx.Scopes(Paginate(page, pageSize))
	tx.Find(&res.ShareRecords)
	return
}

// Create 创建
func (srs *ShareRecordService) Create(u *model.ShareRecord) error {
	res := DB.Create(u).Error
	return res
}
func (srs *ShareRecordService) Delete(u *model.ShareRecord) error {
	return DB.Delete(u).Error
}

// Update 更新
func (srs *ShareRecordService) Update(u *model.ShareRecord) error {
	return DB.Model(u).Updates(u).Error
}

func (srs *ShareRecordService) BatchDelete(ids []uint) error {
	return DB.Where("id in (?)", ids).Delete(&model.ShareRecord{}).Error
}
