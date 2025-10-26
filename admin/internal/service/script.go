package service

import (
	"errors"
	"fmt"
	"pulse/admin/internal/model/request"
	"pulse/common/models"
	"pulse/common/pkg/dbclient"
	"time"
)

// ScriptService 结构体定义了脚本相关的服务
type ScriptService struct{}

// DefaultJobService 是ScriptService的一个默认实例
var DefaultScriptService = new(ScriptService)

// Serach 方法用于根据查询条件搜索脚本
func (script *ScriptService) Search(s *request.ReqScriptSearch) ([]models.Script, int64, error) {
	db := dbclient.GetMysqlDB().Table(models.PulseScriptTableName)

	if s.ID > 0 {
		db = db.Where("id = ?", s.ID)
	}
	if len(s.Name) > 0 {
		db = db.Where("name like ?", s.Name+"%")
	}

	// 这是您要用来接收结果的切片
	scripts := make([]models.Script, 0)
	var total int64

	// Count 操作是正确的，所以 total 能获取到值
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 修正这里：将查询结果填充到 &scripts 中
	err = db.Limit(s.PageSize).Offset((s.Page - 1) * s.PageSize).Find(&scripts).Error
	if err != nil {
		return nil, 0, err
	}

	return scripts, total, nil
}

func (s *ScriptService) Update(req *request.ReqScriptUpdate) error {
	// 更新
	if req.ID != nil && *req.ID > 0 {
		scriptID := *req.ID
		updates := make(map[string]any)

		if req.Name != nil && *req.Name != "" {
			updates["name"] = *req.Name
		}

		if req.Command != nil && *req.Command != "" {
			updates["command"] = *req.Command
		}

		if len(updates) == 0 {
			updates["updated"] = time.Now().Unix()

			db := dbclient.GetMysqlDB()
			err := db.Model(&models.Script{}).Where("id = ?", scriptID).Updates(updates).Error
			if err != nil {
				return fmt.Errorf("failed to update script with id %d: %w", scriptID, err)
			}
		}
		return nil
	}

	// 新建
	if req.Name == nil || *req.Name == "" {
		return errors.New("script name is required for creation")
	}
	if req.Command == nil || *req.Command == "" {
		return errors.New("script command is required for creation")
	}
	newScript := models.Script{
		Name:    *req.Name,
		Command: *req.Command,
		Created: time.Now().Unix(),
	}
	if err := dbclient.GetMysqlDB().Create(&newScript).Error; err != nil {
		return fmt.Errorf("failed to create script: %w", err)
	}

	return nil
}
