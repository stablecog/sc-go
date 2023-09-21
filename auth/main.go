package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/go-co-op/gocron"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"github.com/stablecog/sc-go/auth/api"
	"github.com/stablecog/sc-go/auth/store"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/middleware"
	"github.com/stablecog/sc-go/utils"
	pg "github.com/vgarvardt/go-oauth2-pg/v4"
	"github.com/vgarvardt/go-pg-adapter/pgx4adapter"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
)

func main() {
	// Load .env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Warn("Error loading .env file (this is fine)", "err", err)
	}

	dbconn, err := database.GetSqlDbConn(false)
	if err != nil {
		log.Fatal("Failed to connect to database", "err", err)
		os.Exit(1)
	}

	entClient, err := database.NewEntClient(dbconn)
	if err != nil {
		log.Fatal("Failed to create ent client", "err", err)
		os.Exit(1)
	}
	defer entClient.Close()

	pgxConn, _ := pgx.Connect(context.TODO(), dbconn.DSN())

	manager := manage.NewDefaultManager()

	// use PostgreSQL token store with pgx.Connection adapter
	adapter := pgx4adapter.NewConn(pgxConn)
	tokenStore, _ := pg.NewTokenStore(adapter, pg.WithTokenStoreGCInterval(time.Minute))
	defer tokenStore.Close()

	clientStore, _ := pg.NewClientStore(adapter)

	manager.MapTokenStorage(tokenStore)
	manager.MapClientStorage(clientStore)

	cfg := server.NewConfig()
	cfg.ForcePKCE = true
	srv := server.NewServer(cfg, manager)

	ctx := context.Background()
	redisStore := store.NewRedisStore(ctx)
	apiWrapper := &api.ApiWrapper{
		RedisStore:   redisStore,
		SupabaseAuth: database.NewSupabaseAuth(),
		AesCrypt:     utils.NewAesCrypt(os.Getenv("DATA_ENCRYPTION_PASSWORD")),
		DB:           entClient,
	}

	srv.SetUserAuthorizationHandler(apiWrapper.UserAuthorizeHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Infof("Internal Error: %v", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Infof("Response Error: %v", re.Error.Error())
	})

	// Create router
	app := chi.NewRouter()

	// Cors middleware
	app.Use(cors.Handler(cors.Options{
		AllowedOrigins: utils.GetCorsOrigins(),
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Health check endpoint
	app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		render.Status(r, http.StatusOK)
		render.PlainText(w, r, "OK")
	})

	app.Route("/oauth", func(r chi.Router) {
		r.Use(middleware.Logger)

		r.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
			err := srv.HandleAuthorizeRequest(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})

		r.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
			_ = dumpRequest(os.Stdout, "oauthTokenRequest", r) // Ignore the error

			srv.HandleTokenRequest(w, r)
		})

		r.Post("/approve", apiWrapper.ApproveAuthorization)
	})

	// Cron auth client update
	clients, err := entClient.AuthClient.Query().All(ctx)
	if err != nil {
		log.Fatal("Error querying clients", "err", err)
		return
	}
	store.GetCache().UpdateAuthClients(clients)

	s := gocron.NewScheduler(time.UTC)
	const cacheIntervalSec = 300
	s.Every(cacheIntervalSec).Seconds().StartAt(time.Now().Add(cacheIntervalSec * time.Second)).Do(func() {
		log.Info("📦 Updating cache...")
		clients, err := entClient.AuthClient.Query().All(ctx)
		if err != nil {
			log.Error("Error querying clients", "err", err)
			return
		}
		store.GetCache().UpdateAuthClients(clients)
	})
	s.StartAsync()

	// Start server
	port := utils.GetEnv("PORT", "9096")
	log.Info("Starting language server", "port", port)

	h2s := &http2.Server{}
	authSrv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: h2c.NewHandler(app, h2s),
	}

	log.Info(authSrv.ListenAndServe())
}

func dumpRequest(writer io.Writer, header string, r *http.Request) error {
	data, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err
	}
	writer.Write([]byte("\n" + header + ": \n"))
	writer.Write(data)
	return nil
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	// _ = dumpRequest(os.Stdout, "auth", r) // Ignore the error
	redirectURI := r.FormValue("redirect_uri")

	w.Header().Set("Location", fmt.Sprintf("%s&code=%s", redirectURI, "000000"))
	w.WriteHeader(http.StatusFound)
	return
}
