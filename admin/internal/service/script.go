package service

import (
	"pulse/admin/internal/model/request"
	"pulse/common/models"
	"pulse/common/pkg/dbclient"
)

// ScriptService 结构体定义了脚本相关的服务
type ScriptService struct{}

// DefaultJobService 是ScriptService的一个默认实例
var DefaultScriptService = new(ScriptService)

// Serach 方法用于根据查询条件搜索脚本
func (script *ScriptService) Search(s *request.ReqScriptSearch) ([]models.Script, int64, error) {
	// 获取数据库连接实例，并指定要查询的表
	db := dbclient.GetMysqlDB().Table(models.PulseScriptTableName)

	if s.ID > 0 {
		db = db.Where("id = ?", s.ID)
	}
	if len(s.Name) > 0 {
		db = db.Where("name like ?", s.Name+"%")
	}
	scripts := make([]models.Script, 2)
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 设置分页查询的Limit和Offset，并执行查询
	err = db.Limit(s.PageSize).Offset((s.Page - 1) * s.PageSize).Find(&script).Error
	if err != nil {
		return nil, 0, err
	}

	return scripts, total, nil
}
