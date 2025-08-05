package main

import (
	"GameServer/internal/config"
	"GameServer/internal/database"
	"GameServer/internal/handlers/online"
	"GameServer/internal/server"
	"GameServer/pkg/logger"
	"GameServer/pkg/metrics"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("Starting Game Server...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded successfully")

	// 初始化日志系统
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	logger.Info("Logging system initialized", map[string]interface{}{
		"level":  cfg.Logging.Level,
		"format": cfg.Logging.Format,
	})

	// 初始化监控系统
	metrics.Init()
	logger.Info("Metrics system initialized")
	
	// 初始化限流器
	server.InitRateLimiter(60, time.Minute) // 每分钟60个请求，每分钟清理一次
	logger.Info("Rate limiter initialized")
	
	// 初始化缓存系统
	server.InitCaches()
	logger.Info("Cache system initialized")

	// 连接数据库
	db, err := database.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 创建缺失的表
	log.Println("Creating missing tables...")
	if err := database.CreateMissingTables(db); err != nil {
		log.Fatalf("Error creating missing tables: %v", err)
	}

	// 检查数据库表是否存在
	log.Println("Checking database tables...")
	if err := database.CheckDatabaseTables(db); err != nil {
		log.Fatalf("Error checking database tables: %v", err)
	}

	// 检查表结构
	log.Println("Checking table structures...")
	if err := database.CheckTableStructure(db); err != nil {
		log.Fatalf("Error checking table structure: %v", err)
	}

	log.Println("Database connection and structure check completed successfully!")

	// 创建路由器
	router := server.NewRouter()
	
	// 添加中间件
	router.Use(server.ValidationMiddleware())
	router.Use(server.AuthMiddleware())
	router.Use(server.LoggingMiddleware())
	router.Use(server.RateLimitMiddleware())
	
	// 设置处理器
	server.SetupHandlers(router, db)
	
	// 初始化在线状态管理，将所有用户设置为离线状态
	onlineService := online.NewOnlineService(db)
	if err := onlineService.SetAllUsersOffline(); err != nil {
		log.Printf("Warning: Failed to initialize user online status: %v", err)
	}

	// 创建Hub
	hub := server.NewHub(db, router)
	go hub.Run()
	
	// 初始化性能监控器
	server.InitPerformanceMonitor(hub, 10*time.Second) // 每10秒更新一次指标
	logger.Info("Performance monitor initialized")

	// 设置路由
	http.HandleFunc("/ws", hub.HandleWebSocket)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Game Server is running!"))
	})

	// 健康检查端点
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		}
		json.NewEncoder(w).Encode(response)
	})

	// 监控指标端点（原有的简单指标）
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		currentMetrics := metrics.GetMetrics()
		json.NewEncoder(w).Encode(currentMetrics)
	})

	// 性能监控端点（新的详细性能指标）
	http.HandleFunc("/performance", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if perfMonitor := server.GetPerformanceMonitor(); perfMonitor != nil {
			performanceMetrics := perfMonitor.GetMetricsWithDatabaseStats()
			json.NewEncoder(w).Encode(performanceMetrics)
		} else {
			http.Error(w, "Performance monitor not initialized", http.StatusServiceUnavailable)
		}
	})

	// 路由信息端点
	http.HandleFunc("/routes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		routes := router.GetRoutes()
		json.NewEncoder(w).Encode(routes)
	})

	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", serverAddr)
	log.Printf("WebSocket endpoint: ws://%s/ws", serverAddr)

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
