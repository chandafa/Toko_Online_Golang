package app

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/chandafa/gotoko/database/seeders"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
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
}

func (server *Server) dbMigrate() {
	for _, model := range RegisterModel() {
		err := server.DB.Debug().AutoMigrate(model.Model)

		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Database migrated successfully.")
}

func (server *Server) initCommands(config AppConfig, dbConfig DBConfig) {
	server.initializeDB(dbConfig)

	cmdApp := cli.NewApp()
	cmdApp.Commands = []*cli.Command{
		{
			Name: "db:migrate",
			Action: func(c *cli.Context) error {
				server.dbMigrate()
				return nil
			},
		},
		{
			Name: "db:seed",
			Action: func(c *cli.Context) error {
				err := seeders.DBSeed(server.DB)
				if err != nil {
					log.Fatal(err)
				}
				return nil
			},
		},
	}

	err := cmdApp.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
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

	flag.Parse()
	arg := flag.Arg(0)

	if arg != "" {
		server.initCommands(AppConfig, dbConfig)
	} else {
		server.Initialize(AppConfig, dbConfig)
		server.Run(":" + AppConfig.AppPort)
	}
}
