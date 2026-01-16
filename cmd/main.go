package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"bone_appetit_r4_service/internal/config"
	"bone_appetit_r4_service/internal/handlers"
	"bone_appetit_r4_service/internal/routers"
	"bone_appetit_r4_service/internal/services"
	"bone_appetit_r4_service/pkg/db"
	"bone_appetit_r4_service/pkg/ipfy"
	"bone_appetit_r4_service/pkg/logs"
	"bone_appetit_r4_service/pkg/middleware"
	"bone_appetit_r4_service/pkg/r4bank"
	"fmt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	logger := logs.NewZapLogger()
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("error syncing logger: %v\n", err)
		}
	}()

	sslmode := cfg.SSLMode
	fmt.Printf("sslmode -> %s\n", sslmode)
	if len(sslmode) > 0 {
		sslmode = "sslmode=" + sslmode
	}

	//connect the database
	connStr := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s %s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, sslmode)
	// gorm connect
	gormDB, err := db.NewDBSQLHandler(connStr)
	if err != nil {
		logger.Fatal(err.Error(), zap.Any("host", cfg.DBHost), zap.Any("port", cfg.DBPort), zap.Any("user", cfg.DBUser), zap.Any("dbname", cfg.DBName))
	}

	db, err := gormDB.DB()
	if err != nil {
		logger.Fatal(err.Error(), zap.Any("host", cfg.DBHost), zap.Any("port", cfg.DBPort), zap.Any("user", cfg.DBUser), zap.Any("dbname", cfg.DBName))
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("error db body: %v\n", err)
		}
	}()

	loc, err := time.LoadLocation("America/Caracas")
	if err != nil {
		logger.Fatal("could not load Venezuela time zone", zap.Error(err))
	}

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	// Init resources
	r4BoneRestClient := r4bank.NewClient(cfg.R4BoneEntryPoint, cfg.R4BoneCommerceToken, logger)
	r4AppaRestClient := r4bank.NewClient(cfg.R4APPAEntryPoint, cfg.R4APPACommerceToken, logger)

	// initialize services
	r4BoneService := services.NewR4Service(logger, r4BoneRestClient)
	r4AppaService := services.NewR4Service(logger, r4AppaRestClient)
	webhookService := services.NewWebhookService(gormDB, loc, logger)

	// initialize middleware
	authBoneMiddleware := middleware.NewWebhookAuthMiddleware(cfg.BoneSecret, cfg.R4BoneCommerceToken)
	authAppaMiddleware := middleware.NewWebhookAuthMiddleware(cfg.APPASecret, cfg.R4APPACommerceToken)

	// Initialize handlers
	r4BoneHandler := handlers.NewR4Handler(r4BoneService)
	r4AppaHandler := handlers.NewR4Handler(r4AppaService)
	webhookHandler := handlers.NewWebhookHandler(webhookService)

	// Initialize webhook routes
	r4BoneRoutes := routers.NewR4Routes(r4BoneHandler)
	r4AppaRoutes := routers.NewR4AppaRoutes(r4AppaHandler)
	webhookBoneRouter := routers.NewWebhookRouter(webhookHandler)
	webhookAppaRouter := routers.NewWebhookAppaRouter(webhookHandler)

	// Set up routes with authentication middleware
	r4BoneRoutes.SetRouter(router, authBoneMiddleware)
	webhookBoneRouter.SetRouter(router, authBoneMiddleware)
	r4AppaRoutes.SetRouter(router, authAppaMiddleware)
	webhookAppaRouter.SetRouter(router, authAppaMiddleware)

	// Get IP public
	ipfy.GetIPInfo()
	if gormDB != nil {
		logger.Info("database connected successfully")
	} else {
		logger.Info("database connection failed")
	}

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
