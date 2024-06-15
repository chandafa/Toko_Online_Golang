package app

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	DB     *gorm.DB
	Router *mux.Router
}

type AppConfig struct {
	AppName string
	AppEnv  string
	AppPort string
}

type DBConfig struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	DBDriver   string
}

func (server *Server) Initialize(AppConfig AppConfig, dbConfig DBConfig) {
	fmt.Println("Welcome to" + AppConfig.AppName)

	server.initializeDB(dbConfig)
	server.InitializeRoutes()
}

func (server *Server) Run(addr string) {
	fmt.Printf("Listening to port %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.Router))
}

func (server *Server) initializeDB(dbConfig DBConfig) {
	var err error
	if dbConfig.DBDriver == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBHost, dbConfig.DBPort, dbConfig.DBName)
		server.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	} else {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta", dbConfig.DBHost, dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBName, dbConfig.DBPort)
		server.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		panic("failed on connect to the database")
	}

	for _, model := range RegisterModel() {
		err = server.DB.Debug().AutoMigrate(model.Model)

		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Database migrated successfully.")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func Run() {
	var server = Server{}
	var AppConfig = AppConfig{}
	var dbConfig = DBConfig{}

	err := godotenv.Load()
	if err != nil {
		log.Fatalf(("Error on loading .env file"))
	}

	AppConfig.AppName = getEnv("APP_NAME", "Gotoko")
	AppConfig.AppEnv = getEnv("APP_ENV", "development")
	AppConfig.AppPort = getEnv("APP_PORT", "9000")

	dbConfig.DBHost = getEnv(("DB_HOST"), "localhost")
	dbConfig.DBUser = getEnv(("DB_USER"), "root")
	dbConfig.DBPassword = getEnv(("DB_PASSWORD"), "root")
	dbConfig.DBName = getEnv(("DB_NAME"), "chandafa_gotoko")
	dbConfig.DBPort = getEnv(("DB_PORT"), "5432")
	dbConfig.DBDriver = getEnv(("DB_DRIVER"), "postgres")

	server.Initialize(AppConfig, dbConfig)
	server.Run(":" + AppConfig.AppPort)
}
