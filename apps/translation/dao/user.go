package dao

import (
	"github.com/jovian1994/cxh-1207-be-interview/models"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/jwt"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/logger"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/mysql_tool"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/unify_response"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type IUserDao interface {
	CreateUser(username, password string, role int) (id int64, err error)
	AuthUser(username, password string) (jwtString string, err error)
}

func NewUserDao(dbClientName string, jwtVerify jwt.ITokenVerify) IUserDao {
	return &userDao{
		dbClientName: dbClientName,
		jwtVerify:    jwtVerify,
	}
}

type userDao struct {
	dbClientName string
	db           *mysql_tool.DB
	jwtVerify    jwt.ITokenVerify
}

func (u *userDao) AuthUser(username, password string) (string, error) {
	existed, userModel, err := u.userExisted(username)
	if err != nil {
		return "", err
	}
	if !existed || userModel == nil {
		return "", unify_response.UseNotExist("用户不存在")
	}
	err = bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(password))
	if err != nil {
		return "", unify_response.UseNotExist("密码错误")
	}
	token, err := u.jwtVerify.GenerateJWT(username)
	if err != nil {
		// todo 记录日志
		return "", unify_response.ServerError("生成 jwt 失败")
	}
	return token, nil
}

func (u *userDao) userExisted(username string) (bool, *models.UserModel, error) {
	db := u.getDBClient()
	var count int64
	var userModel = &models.UserModel{}
	err := db.Model(models.UserModel{}).Where("username =?", username).
		Count(&count).Scan(userModel).
		Error
	if err != nil {
		logger.Error("查询用户失败",
			zap.String("err:", err.Error()),
			zap.String("username", username))
		return false, nil, unify_response.DBError("查询用户失败")
	}
	return count > 0, userModel, nil
}

func (u *userDao) CreateUser(username, password string, role int) (int64, error) {
	existed, _, err := u.userExisted(username)
	if err != nil {
		return 0, err
	}
	if existed {
		return 0, unify_response.UserAlreadyExist("用户已存在")
	}
	var userModel = &models.UserModel{}
	userModel.Username = username
	hashedPassword, err := bcrypt.
		GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("密码加密失败",
			zap.String("err:", err.Error()))
		return 0, err
	}
	userModel.Password = string(hashedPassword)
	userModel.Role = role
	db := u.getDBClient()
	err = db.Model(models.UserModel{}).Create(&userModel).Error
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
