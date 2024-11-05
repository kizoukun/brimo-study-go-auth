package main

import (
	"fmt"
	"go-auth/server/api"
	"go-auth/server/config"
	manager "go-auth/server/jwt"
	"go-auth/server/lib/rabbitmq"
	"go-auth/server/models"
	"go-auth/server/pb"
	"net"
	"net/http"

	"go-auth/www/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	elasticLog "gopkg.in/sohlich/elogrus.v7"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const serviceName = "go-auth-service"
const defaultPort = 50051
const APP_PORT = "9090"

var appConfig *config.Config

func main() {

	appConfig = config.InitConfig(".env")

	docs.SwaggerInfo.Title = "Go Auth API"
	docs.SwaggerInfo.Description = "API documentation for Go Auth Service"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", APP_PORT)
	docs.SwaggerInfo.BasePath = "/api"

	esLogger := logrus.New()
	esLogger.SetFormatter(&ecslogrus.Formatter{})
	esLogger.SetLevel(logrus.InfoLevel)

	client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"), elastic.SetSniff(false))
	if err != nil {
		esLogger.Fatalf("Failed to create elasticsearch client: %v", err)
	}

	hook, err := elasticLog.NewAsyncElasticHook(client, serviceName, logrus.InfoLevel, "go-auth-logs")
	if err != nil {
		logrus.Fatalf("Failed to create Elasticsearch hook: %v", err)
	}
	esLogger.AddHook(hook)

	dsn := "go_user:go_password@tcp(127.0.0.1:3306)/go_auth?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		esLogger.Fatalf("failed to connect database: %v", err)
	}

	db.AutoMigrate(&models.User{})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		esLogger.Fatalf("failed to listen: %v", err)
	}

	rabbitmq.InitRabbitMq()

	// Start the RabbitMQ consumer in a goroutine
	go rabbitmq.StartConsumer("orders")

	grpcServer := grpc.NewServer()
	userService := &api.Server{
		Db:      db,
		Logger:  esLogger,
		Manager: manager.NewJWTManager(appConfig.AppKey, 3600, esLogger),
	}

	router := gin.Default()

	// Route to serve the Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register API endpoints
	router.POST("/api/login", loginUser(userService))

	go router.Run(fmt.Sprintf(":%s", APP_PORT))
	pb.RegisterUserServiceServer(grpcServer, userService)

	if err := grpcServer.Serve(lis); err != nil {
		esLogger.Fatalf("failed to serve: %v", err)
	}
}

type FailureError struct {
	Error string `json:"error"`
}

type LoginUserSuccess struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

// @Summary		Login User
// @Description	Authenticates a user and returns a token
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			requestBody	body		pb.LoginUserRequest	true	"Login Request"
// @Success		200			{object}	LoginUserSuccess
// @Failure		401			{object}	FailureError
// @Router			/login [post]
func loginUser(userService *api.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var credentials pb.LoginUserRequest

		if err := c.ShouldBindJSON(&credentials); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Call userService's login method and handle response
		resp, err := userService.LoginUser(c.Request.Context(), &credentials)
		if err != nil {
			c.JSON(http.StatusUnauthorized, FailureError{
				Error: "Unauthorized",
			})
			return
		}

		c.JSON(http.StatusOK, LoginUserSuccess{
			Message: "Login successful",
			Token:   resp.AccessToken,
		})
	}
}
