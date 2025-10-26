package service

import (
	"errors"
	"fmt"
	"pulse/admin/internal/model/request"
	"pulse/common/models"
	"pulse/common/pkg/dbclient"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 定义了用户相关的服务
type UserService struct{}

// DefaultUserService 是UserService的一个默认实例
var DefaultUserService = new(UserService)

// Login 方法用于用户登陆验证
func (us *UserService) Login(username, password string) (u *models.User, err error) {
	u = new(models.User)
	err = dbclient.GetMysqlDB().
		Table(models.PulseUserTableName).
		Where("username = ?", username).
		First(u).Error
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return nil, errors.New("incorrect username or password")
	}
	return u, nil
}

// FindByUserName 方法通过用户名查询用户
func (us *UserService) FindByUserName(username string) (u *models.User, err error) {
	// 初始化一个新的User实例
	u = new(models.User)
	// 在用户表中根据用户名查找第一条记录
	err = dbclient.GetMysqlDB().Table(models.PulseUserTableName).Where("username = ?", username).First(u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return
}

// ChangePassword 方法用于修改用户密码
func (us *UserService) ChangePassword(userId int, oldPassword, newPassword string) error {
	var user models.User
	err := dbclient.GetMysqlDB().Table(models.PulseUserTableName).Where("id = ?", userId).First(&user).Error
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("old password is not correct")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return dbclient.GetMysqlDB().Table(models.PulseUserTableName).Where("id = ?", userId).Update("password", string(hashedPassword)).Error
}

// func (us *UserService) UpdateUser(userId int, req *request.ReqUserUpdate) error {
// 	// 使用事务保证“检查唯一性”和“更新”的原子性
// 	return dbclient.GetMysqlDB().Transaction(func(tx *gorm.DB) error {
// 		var user models.User
// 		// 1. 查询现有用户数据
// 		if err := tx.Table(models.PulseUserTableName).Where("id = ?", userId).First(&user).Error; err != nil {
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				return errors.New("user not found")
// 			}
// 			return err
// 		}

// 		updates := make(map[string]any)

// 		// 2. 根据请求构建 updates map
// 		if req.UserName != nil && *req.UserName != user.UserName {
// 			// 检查新用户名是否已被其他用户占用
// 			var count int64
// 			err := tx.Model(&models.User{}).Where("username = ? AND id != ?", req.UserName, userId).Count(&count).Error
// 			if err != nil {
// 				return fmt.Errorf("failed to check username uniqueness: %w", err)
// 			}
// 			if count > 0 {
// 				return errors.New("username already exists")
// 			}
// 			updates["username"] = req.UserName
// 		}

// 		if req.Email != nil {
// 			updates["email"] = req.Email
// 		}

// 		if req.Role != nil {
// 			updates["role"] = req.Role
// 		}

// 		if len(updates) != 0 {
// 			updates["updated"] = time.Now().Unix()

// 			// 执行更新
// 			err := tx.Model(&models.User{}).Where("id = ?", userId).Updates(updates).Error
// 			if err != nil {
// 				return fmt.Errorf("failed to update user: %w", err)
// 			}
// 		}

//			return nil
//		})
//	}

func (us *UserService) UpdateUser(userId int, req *request.ReqUserUpdate) error {
	// 使用事务保证“检查唯一性”和“更新”的原子性
	return dbclient.GetMysqlDB().Transaction(func(tx *gorm.DB) error {
		var user models.User
		// 1. 查询现有用户数据
		if err := tx.Table(models.PulseUserTableName).Where("id = ?", userId).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("用户不存在")
			}
			return err
		}

		updates := make(map[string]any)

		// 2. 根据请求构建 updates map
		// ！！！dto层传入指针
		if req.UserName != nil && *req.UserName != user.UserName { // 无传入值，空值为空字符串
			if *req.UserName != "" { // 无传入值，值为空字符串
				// 检查新用户名是否已被其他用户占用
				var count int64
				err := tx.Model(&models.User{}).Where("username = ? AND id != ?", *req.UserName, userId).Count(&count).Error
				if err != nil {
					return fmt.Errorf("检查用户名唯一性失败: %w", err)
				}
				if count > 0 {
					return errors.New("用户名已存在")
				}
				updates["username"] = *req.UserName // 在这里解引用指针
			}
		}

		if req.Email != nil {
			// 增加一个检查，防止更新为空字符串
			if *req.Email != "" {
				updates["email"] = *req.Email // 在这里解引用指针
			}
		}

		if req.Role != nil {
			updates["role"] = *req.Role // 在这里解引用指针
		}

		if len(updates) != 0 {
			updates["updated"] = time.Now().Unix()

			// 执行更新
			err := tx.Model(&models.User{}).Where("id = ?", userId).Updates(updates).Error
			if err != nil {
				return fmt.Errorf("更新用户失败: %w", err)
			}
		}

		return nil
	})
}

// Search 方法用于根据条件搜索用户列表
func (us *UserService) Search(s *request.ReqUserSearch) ([]models.User, int64, error) {
	// 获取数据库连接实例，并指定查询的表
	db := dbclient.GetMysqlDB().Table(models.PulseUserTableName)

	if len(s.UserName) > 0 {
		db = db.Where("username like ?", s.UserName+"%")
	}
	if len(s.Email) > 0 {
		db.Where("email = ?", s.Email)
	}
	if s.Role > 0 {
		db.Where("role = ?", s.Role)
	}
	if s.ID > 0 {
		db.Where("id = ?", s.ID)
	}
	// 初始化一个存放用户的切片
	users := make([]models.User, 0)
	var total int64 // 用于存放总记录数的变量
	// 执行查询以获取总记录数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 执行分页查询，并指定查询的字段
	err = db.Select("id", "username", "email", "created", "updated").
		Limit(s.PageSize).                 // 限制每页返回的记录数
		Offset((s.Page - 1) * s.PageSize). // 计算当前页的起始位置
		Find(&users).Error                 // 执行查询并将结果集填充到users切片中
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
