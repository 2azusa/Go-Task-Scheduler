package service

import (
	"database/sql/driver"
	"errors"
	"pulse/common/models"
	"pulse/common/pkg/dbclient"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 初始化一个模拟的GORM DB实例和sqlmock控制器
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

// 测试登陆方法
func TestUserSErvice_Login(t *testing.T) {
	// 创建一个模拟的密码哈希
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	mockUser := &models.User{
		ID:       1,
		UserName: "testuser",
		Password: string(hashedPassword),
	}

	// 定义测试用例
	testCases := []struct {
		name          string
		UserName      string
		password      string
		mockSetup     func(mock sqlmock.Sqlmock) // 用来设置mock期望
		expectedUser  *models.User
		expectedError string
	}{
		{
			name:     "Success",
			UserName: "testuser",
			password: "correct-password",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// 定义期望返回的行
				rows := sqlmock.NewRows([]string{"id", "UserName", "password"}).
					AddRow(mockUser.ID, mockUser.UserName, mockUser.Password)
				// 期望执行一个查询、并返回上面定义的行
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `pulse_user` WHERE UserName = ? ORDER BY `pulse_user`.`id` LIMIT 1")).
					WithArgs("testuser").
					WillReturnRows(rows)

			},
			expectedUser:  mockUser,
			expectedError: "",
		},
		{
			name:     "User Not Found",
			UserName: "nonexistent",
			password: "any-password",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `pulse_user` WHERE UserName = ? ORDER BY `pulse_user`.`id` LIMIT 1")).
					WithArgs("nonexistent").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: "record not found",
		},
		{
			name:     "Incorrect Password",
			UserName: "testuser",
			password: "wrong-password",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "UserName", "password"}).
					AddRow(mockUser.ID, mockUser.UserName, mockUser.Password)
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `pulse_user` WHERE UserName = ? ORDER BY `pulse_user`.`id` LIMIT 1")).
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			expectedUser:  nil,
			expectedError: "incorrect UserName or password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. 设置Mock
			gormDB, mock := setupMockDB(t)
			dbclient.SetMysqlDB(gormDB) // 注入mock DB
			tc.mockSetup(mock)

			// 2. 执行被测试的函数
			userService := &UserService{}
			user, err := userService.Login(tc.UserName, tc.password)

			// 3. 断言结果
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				// 由于返回的User模型可能包含其他零值字段，我们只比较关键字段
				assert.Equal(t, tc.expectedUser.ID, user.ID)
				assert.Equal(t, tc.expectedUser.UserName, user.UserName)
			}

			// 4. 确保所有mock期望都已满足
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestUserService_ChangePassword 测试修改密码方法
func TestUserService_ChangePassword(t *testing.T) {
	oldHashedPassword, _ := bcrypt.GenerateFromPassword([]byte("old-password"), bcrypt.DefaultCost)
	mockUser := &models.User{
		ID:       1,
		UserName: "testuser",
		Password: string(oldHashedPassword),
		Created:  time.Now().Unix(),
		Updated:  time.Now().Unix(),
	}

	testCases := []struct {
		name          string
		userId        int
		oldPassword   string
		newPassword   string
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedError string
	}{
		{
			name:        "Success",
			userId:      1,
			oldPassword: "old-password",
			newPassword: "new-password",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// 期望查询用户
				rows := sqlmock.NewRows([]string{"id", "UserName", "password", "created_at", "updated_at"}).
					AddRow(mockUser.ID, mockUser.UserName, mockUser.Password, mockUser.Created, mockUser.Updated)
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `pulse_user` WHERE id = ? ORDER BY `pulse_user`.`id` LIMIT 1")).
					WithArgs(1).
					WillReturnRows(rows)

				// 期望开始一个事务
				mock.ExpectBegin()
				// 期望执行更新操作
				mock.ExpectExec(regexp.QuoteMeta(
					"UPDATE `pulse_user` SET `password`=? WHERE id = ?")).
					// 我们无法精确匹配bcrypt后的哈希，所以使用AnyArgs()
					WithArgs(sqlmock.AnyArg(), 1).
					WillReturnResult(driver.RowsAffected(1))
				// 期望提交事务
				mock.ExpectCommit()
			},
			expectedError: "",
		},
		{
			name:        "Old Password Incorrect",
			userId:      1,
			oldPassword: "wrong-old-password",
			newPassword: "new-password",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "UserName", "password"}).
					AddRow(mockUser.ID, mockUser.UserName, mockUser.Password)
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `pulse_user` WHERE id = ? ORDER BY `pulse_user`.`id` LIMIT 1")).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedError: "old password is not correct",
		},
		{
			name:        "User Not Found",
			userId:      99,
			oldPassword: "any",
			newPassword: "any",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `pulse_user` WHERE id = ? ORDER BY `pulse_user`.`id` LIMIT 1")).
					WithArgs(99).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: "record not found",
		},
		{
			name:        "Update Fails",
			userId:      1,
			oldPassword: "old-password",
			newPassword: "new-password",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "UserName", "password"}).
					AddRow(mockUser.ID, mockUser.UserName, mockUser.Password)
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `pulse_user` WHERE id = ? ORDER BY `pulse_user`.`id` LIMIT 1")).
					WithArgs(1).
					WillReturnRows(rows)

				dbError := errors.New("db update error")
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(
					"UPDATE `pulse_user` SET `password`=? WHERE id = ?")).
					WithArgs(sqlmock.AnyArg(), 1).
					WillReturnError(dbError)
				mock.ExpectRollback() // 更新失败时，gorm会回滚
			},
			expectedError: "db update error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gormDB, mock := setupMockDB(t)
			dbclient.SetMysqlDB(gormDB)

			tc.mockSetup(mock)

			userService := &UserService{}
			err := userService.ChangePassword(tc.userId, tc.oldPassword, tc.newPassword)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
