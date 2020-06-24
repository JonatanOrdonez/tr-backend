package main

import (
	"fmt"
	"log"
	"os"

	cors "github.com/AdhityaRamadhanus/fasthttpcors"
	controllers "github.com/JonatanOrdonez/tr-backend/controllers"
	db "github.com/JonatanOrdonez/tr-backend/db"
	repositories "github.com/JonatanOrdonez/tr-backend/repositories"
	"github.com/JonatanOrdonez/tr-backend/services"
	fasthttprouter "github.com/buaazp/fasthttprouter"
	goDotenv "github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
)

func getServers(ctx *fasthttp.RequestCtx) {
}

func main() {
	env := os.Getenv("GO_ENV")
	if env == "dev" {
		envErr := goDotenv.Load(".env.local")
		if envErr != nil {
			fmt.Println("Local variables cannot be loaded")
		}
	}

	// Init environment variables...
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	whiteList := os.Getenv("WHITE_LIST")
	port := os.Getenv("PORT")

	// Init database...
	db, err := db.StartPostgresqlConnection(dbUser, dbHost, dbName)
	if err != nil {
		log.Fatal(err.Error())
	} else {
		// Init repositories...
		domainRepo := repositories.NewDomainRepository(db)
		domainService := services.NewDomainService(domainRepo)
		domainController := controllers.NewDomainController(domainService)

		// Init router...
		router := fasthttprouter.New()
		router.GET("/api/v1/analyze", domainController.ResponseCheckDomain)

		withCors := cors.NewCorsHandler(cors.Options{
			AllowedOrigins:   []string{whiteList},
			AllowedHeaders:   []string{"x-something-client", "Content-Type"},
			AllowedMethods:   []string{"GET"},
			AllowCredentials: false,
			AllowMaxAge:      5600,
			Debug:            true,
		})

		fmt.Println("Starting server...")
		if err := fasthttp.ListenAndServe(":"+port, withCors.CorsMiddleware(router.Handler)); err != nil {
			log.Fatalf("Error in ListenAndServe: %s", err.Error())
		}
	}
}
