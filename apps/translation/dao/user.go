package dao

import (
	"github.com/jovian1994/cxh-1207-be-interview/models"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/logger"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/mysql_tool"
	"github.com/jovian1994/cxh-1207-be-interview/unify_response"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type IUserDao interface {
	CreateUser(username, password string, role int) (id int64, err error)
	AuthUser(username, password string) (id int64, err error)
}

func NewUserDao(dbClientName string) IUserDao {
	return &userDao{
		dbClientName: dbClientName}
}

type userDao struct {
	dbClientName string
	db           *mysql_tool.DB
}

func (u *userDao) AuthUser(username, password string) (id int64, err error) {

}

func (u *userDao) userExisted(username string) (bool, error) {
	db := u.getDBClient()
	var count int64
	err := db.Model(models.UserModel{}).Where("username =?", username).
		Count(&count).Error
	if err != nil {
		logger.Error("查询用户失败",
			zap.String("err:", err.Error()),
			zap.String("username", username))
		return false, unify_response.DBError("查询用户失败")
	}
	return count > 0, nil
}

func (u *userDao) CreateUser(username, password string, role int) (int64, error) {
	existed, err := u.userExisted(username)
	if err != nil {
		return 0, err
	}
	if existed {
		return 0, unify_response.UserAlreadyExist("用户已存在")
	}
	var userModel = &models.UserModel{}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	db := u.getDBClient()
	err := db.Model(models.UserModel{}).Create(&models.UserModel{
		Username: username,
		Password: password,
		Role:     "user",
		Version:  1,
	}).Error
	if err != nil {
		logger.Error("创建用户失败",
			zap.String("err:", err.Error()),
			zap.String("username", username),
			zap.Int("role", role))
		return 0, unify_response.DBError("创建用户失败")
	}
	return int64(userModel.ID), nil
}

func (u *userDao) getDBClient() *mysql_tool.DB {
	if u.db != nil {
		return u.db
	}
	u.db = mysql_tool.GetMysqlClient(u.dbClientName)
	return u.db
}
