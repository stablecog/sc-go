package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/log"
	pg "github.com/vgarvardt/go-oauth2-pg/v4"
	"github.com/vgarvardt/go-pg-adapter/pgx4adapter"

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

	srv.SetUserAuthorizationHandler(userAuthorizeHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Infof("Internal Error: %v", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Infof("Response Error: %v", re.Error.Error())
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		// _ = dumpRequest(os.Stdout, "oauthTokenRequest", r) // Ignore the error

		srv.HandleTokenRequest(w, r)
	})

	portvar := 9096

	log.Infof("Server is running at %d port.\n", portvar)
	log.Infof("Point your OAuth client Auth endpoint to %s:%d%s", "http://localhost", portvar, "/oauth/authorize")
	log.Infof("Point your OAuth client Token endpoint to %s:%d%s", "http://localhost", portvar, "/oauth/token")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portvar), nil))
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

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	// _ = dumpRequest(os.Stdout, "userAuthorizeHandler", r) // Ignore the error
	redirectURI := r.FormValue("redirect_uri")
	log.Infof("redirect uri %s", redirectURI)

	w.Header().Set("Location", fmt.Sprintf("%s&code=%s", redirectURI, "000000"))
	w.WriteHeader(http.StatusFound)
	return

	// store, err := session.Start(r.Context(), w, r)
	// if err != nil {
	// 	return
	// }

	// uid, ok := store.Get("LoggedInUserID")
	// if !ok {
	// 	if r.Form == nil {
	// 		r.ParseForm()
	// 	}

	// 	store.Set("ReturnUri", r.Form)
	// 	store.Save()

	// 	w.Header().Set("Location", "/login")
	// 	w.WriteHeader(http.StatusFound)
	// 	return
	// }

	// userID = uid.(string)
	// store.Delete("LoggedInUserID")
	// store.Save()
	return
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	// _ = dumpRequest(os.Stdout, "auth", r) // Ignore the error
	redirectURI := r.FormValue("redirect_uri")
	log.Infof()

	w.Header().Set("Location", fmt.Sprintf("%s&code=%s", redirectURI, "000000"))
	w.WriteHeader(http.StatusFound)
	return
}
