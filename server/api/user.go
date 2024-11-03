package api

import (
	"context"
	"errors"
	"go-auth/server/models"
	"go-auth/server/pb"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
)

func (s *Server) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.DefaultResponse, error) {

	s.Logger.Info("RegisterUser function was invoked with request: ", req)

	existingUser, err := s.GetUserByName(ctx, req.Name)
	if err == nil && existingUser != nil {
		s.Logger.Warn("User with name already exists: ", req.Name)
		return nil, errors.New("username already taken")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := HashPassword(req.Password)

	user := &models.User{
		Name:     req.Name,
		Password: hashedPassword,
	}

	s.Logger.Info("Creating user to db: ", user)
	if err := s.Db.Create(user).Error; err != nil {
		return nil, err
	}

	return &pb.DefaultResponse{
		Error:   false,
		Code:    http.StatusOK,
		Message: "Success",
	}, nil
}

func (s *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	user, err := s.GetUserByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	if req.GetName() == "" {
		return nil, errors.New("name is required")
	}

	if req.GetPassword() == "" {
		return nil, errors.New("password is required")
	}

	if !VerifyPassword(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	jwtToken, err := s.Manager.Generate(
		strconv.FormatUint(uint64(user.Id), 10),
	)

	return &pb.LoginUserResponse{
		Error:       false,
		Code:        http.StatusOK,
		Message:     "Success",
		AccessToken: jwtToken.GetAccessToken(),
	}, nil
}

func (s *Server) GetUserByName(ctx context.Context, name string) (*models.User, error) {
	user := &models.User{}
	if err := s.Db.Where("name = ?", name).First(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	
	var user models.User
	
	err := s.Db.Where("id = ?", req.Id).First(&user).Error; 
	if err != nil {
		return nil, err
	}

	return &pb.GetUserResponse{Id: req.Id, Name: user.Name}, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
