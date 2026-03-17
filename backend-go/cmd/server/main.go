package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ruankao/gaojia-backend-go/internal/db"
	"github.com/ruankao/gaojia-backend-go/internal/handler"
	"github.com/ruankao/gaojia-backend-go/internal/repository"
	"github.com/ruankao/gaojia-backend-go/internal/service"
)

func main() {
	driver := flag.String("driver", "sqlite", "database driver: sqlite | mysql")
	dbPath := flag.String("db", "gaojia.db", "SQLite database path (when driver=sqlite)")
	mysqlDSN := flag.String("mysql-dsn", "", "MySQL DSN (when driver=mysql), e.g. user:pass@tcp(127.0.0.1:3306)/dbname")
	port := flag.String("port", "5170", "HTTP port")
	flag.Parse()

	var database *sql.DB
	var err error
	var driverType db.Driver

	switch *driver {
	case "mysql":
		driverType = db.DriverMySQL
		dsn := *mysqlDSN
		if dsn == "" {
			dsn = os.Getenv("MYSQL_DSN")
		}
		if dsn == "" {
			log.Fatal("mysql driver requires -mysql-dsn or MYSQL_DSN env")
		}
		database, err = db.Open(driverType, dsn)
	default:
		driverType = db.DriverSQLite
		database, err = db.Open(driverType, *dbPath)
	}
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	defaultUserID := int64(1)
	if v := os.Getenv("DEFAULT_USER_ID"); v != "" {
		if id, err := parseInt64(v); err == nil {
			defaultUserID = id
		}
	}

	chapterRepo := repository.NewChapterRepository(database)
	questionRepo := repository.NewQuestionRepository(database, driverType)
	chapterService := service.NewChapterService(chapterRepo)
	questionService := service.NewQuestionService(questionRepo, chapterRepo, defaultUserID)
	h := handler.NewHandler(chapterService, questionService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(corsMiddleware())
	h.Register(r)

	log.Printf("gaojia-backend-go listening on :%s", *port)
	if err := r.Run(":" + *port); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
