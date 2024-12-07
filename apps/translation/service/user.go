package service

import "github.com/jovian1994/cxh-1207-be-interview/apps/translation/dao"

type IUserService interface {
	HandlerLogin(username, password string) (string, error)
	HandlerRegister(username, password string) error
}

type userService struct {
	userDao dao.IUserDao
}

func NewUserService(userDao dao.IUserDao) IUserService {
	return &userService{
		userDao: userDao,
	}
}

func (u *userService) HandlerLogin(username, password string) (string, error) {
	jwtString, err := u.userDao.AuthUser(username, password)
	if err != nil {
		return "", err
	}
	return jwtString, nil
}

func (u *userService) HandlerRegister(username, password string) error {
	_, err := u.userDao.CreateUser(username, password, 0)
	if err != nil {
		return err
	}
	return nil
}
