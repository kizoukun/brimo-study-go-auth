package api

import (
	manager "go-auth/server/jwt"
	"go-auth/server/pb"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Server struct {
	pb.UserServiceServer
	Db      *gorm.DB
	Logger  *logrus.Logger
	Manager *manager.JWTManager
}
