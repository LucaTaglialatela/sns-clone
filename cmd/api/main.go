package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/HENNGE/snsclone-202506-golang-luca/api"
	"github.com/HENNGE/snsclone-202506-golang-luca/database"
	"github.com/HENNGE/snsclone-202506-golang-luca/frontend"
	"github.com/HENNGE/snsclone-202506-golang-luca/repository"
	"github.com/HENNGE/snsclone-202506-golang-luca/service"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found")
	}

	awsRegion, exists := os.LookupEnv("AWS_REGION")
	if !exists {
		log.Fatal("Undefined AWS region")
	}

	awsEndpoint, exists := os.LookupEnv("AWS_ENDPOINT")
	if !exists {
		log.Print("Undefined AWS endpoint, falling back to default")
	}

	db, err := database.GetDatabase(ctx, awsRegion, awsEndpoint)
	if err != nil {
		log.Fatal("Failed to get database")
	}

	s3Client, err := database.GetS3Client(ctx, awsRegion, awsEndpoint)
	if err != nil {
		log.Fatal("Failed to get s3 client")
	}

	presignClient := s3.NewPresignClient(s3Client)

	tableName, exists := os.LookupEnv("TABLE_NAME")
	if !exists {
		log.Fatal("Undefined table name")
	}

	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("Undefined bucket name")
		return
	}

	clientID, exists := os.LookupEnv("GOOGLE_OAUTH2_CLIENT_ID")
	if !exists {
		log.Fatal("Undefined Google client id")
	}

	clientSecret, exists := os.LookupEnv("GOOGLE_OAUTH2_CLIENT_SECRET")
	if !exists {
		log.Fatal("Undefined Google client secret")
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		log.Fatal("Undefined port")
	}

	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		log.Fatal(err)
	}

	baseUrl, exists := os.LookupEnv("BASE_URL")
	if !exists {
		log.Fatal("Undefined base url")
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  fmt.Sprintf("%s/auth/google/callback", baseUrl),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	secret, exists := os.LookupEnv("JWT_SECRET")
	if !exists {
		log.Fatal("Undefined secret")
	}

	authConfig := &api.AuthConfig{
		BaseUrl:  baseUrl,
		Secret:   secret,
		Config:   oauth2Config,
		Provider: provider,
	}

	// For serving the contents of the frontend/dist folder as static pages
	// Make sure to create the dist folder by running `npm run build` inside
	// the frontend folder
	fs := http.FileServer(http.FS(frontend.DistFS))

	repositories := repository.InitRepositories(db, s3Client, presignClient, tableName, bucketName)
	services := service.InitServices(repositories)
	handlers := api.InitHandlers(services, authConfig, fs, presignClient)

	router := api.NewRouter(handlers, secret)

	log.Printf("Listening on %s", baseUrl)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
