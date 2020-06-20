package main

import (
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
	err := goDotenv.Load()
	if err != nil {
		log.Fatal("Environmnet variables cannot be loaded")
	} else {
		// Init environment variables...
		dbHost := os.Getenv("DB_HOST")
		dbName := os.Getenv("DB_NAME")
		dbUser := os.Getenv("DB_USER")
		port := os.Getenv("PORT")

		// Init database...
		db, err := db.StartPostgresqlConnection(dbUser, dbHost, dbName)
		if err != nil {
			log.Fatal("Database cannot be started")
		} else {
			// Init repositories...
			domainRepo := repositories.NewDomainRepository(db)
			domainService := services.NewDomainService()
			domainHandler := controllers.NewDomainHandler(domainRepo, domainService)

			// Init router...
			router := fasthttprouter.New()
			router.GET("/api/v1/analyze", domainHandler.CheckDomain)

			withCors := cors.NewCorsHandler(cors.Options{
				// if you leave allowedOrigins empty then fasthttpcors will treat it as "*"
				AllowedOrigins: []string{"*"}, // Only allow example.com to access the resource
				// if you leave allowedHeaders empty then fasthttpcors will accept any non-simple headers
				AllowedHeaders: []string{"x-something-client", "Content-Type"}, // only allow x-something-client and Content-Type in actual request
				// if you leave this empty, only simple method will be accepted
				AllowedMethods:   []string{"GET", "POST"}, // only allow get or post to resource
				AllowCredentials: false,                   // resource doesn't support credentials
				AllowMaxAge:      5600,                    // cache the preflight result
				Debug:            true,
			})
			if err := fasthttp.ListenAndServe(":"+port, withCors.CorsMiddleware(router.Handler)); err != nil {
				log.Fatalf("Error in ListenAndServe: %s", err.Error())
			}
		}
	}
}
