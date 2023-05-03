package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/jon-whit/openfga-demo/middleware/auth"
	"github.com/jon-whit/openfga-demo/service"
	_ "github.com/lib/pq"
	. "github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
)

type Document struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "postgres"
)

func main() {
	storeID, ok := os.LookupEnv("FGA_STORE_ID")
	if !ok {
		log.Fatal("'FGA_STORE_ID' environment variable must be set")
	}

	clientID, ok := os.LookupEnv("FGA_CLIENT_ID")
	if !ok {
		log.Fatal("'FGA_CLIENT_ID' environment variable must be set")
	}

	clientSecret, ok := os.LookupEnv("FGA_CLIENT_SECRET")
	if !ok {
		log.Fatal("'FGA_CLIENT_SECRET' environment variable must be set")
	}

	config := &ClientConfiguration{
		ApiHost: "api.us1.fga.dev", // required, define without the scheme (e.g. api.fga.example instead of https://api.fga.example)
		StoreId: storeID,           // not needed when calling `CreateStore` or `ListStores`
		//AuthorizationModelId: openfga.PtrString(os.Getenv("OPENFGA_AUTHORIZATION_MODEL_ID")),
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodClientCredentials,
			Config: &credentials.Config{
				ClientCredentialsClientId:       clientID,
				ClientCredentialsClientSecret:   clientSecret,
				ClientCredentialsApiAudience:    "https://api.us1.fga.dev/",
				ClientCredentialsApiTokenIssuer: "fga.us.auth0.com",
			},
		},
	}
	uri := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", uri)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	client, err := NewSdkClient(config)
	if err != nil {
		log.Fatalf("failed to initialize fga client: %v", err)
	}

	c := controller{
		s: service.Service{
			Database:  db,
			FGAClient: client,
		},
	}

	r := mux.NewRouter()

	r.Use(auth.JWTTokenVerifierMiddleware("mysecret"))

	r.HandleFunc("/groups", c.CreateGroupHandler).Methods(http.MethodPost)

	r.HandleFunc("/share", c.ShareHandler).Methods(http.MethodPut)

	r.HandleFunc("/documents", c.CreateDocumentHandler).Methods(http.MethodPost)
	r.HandleFunc("/documents", c.GetDocumentsHandler).Methods(http.MethodGet)
	r.HandleFunc("/documents/{id}", c.GetDocumentHandler).Methods(http.MethodGet)

	r.HandleFunc("/folders", c.CreateFolderHandler).Methods(http.MethodPost)
	r.HandleFunc("/folders/{id}", c.GetFolderHandler).Methods(http.MethodGet)
	r.HandleFunc("/folders", c.GetFoldersHandler).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:         ":8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
}

type controller struct {
	s service.Service
}

func (c *controller) CreateDocumentHandler(w http.ResponseWriter, r *http.Request) {

	var req service.CreateDocumentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		// handle error
	}

	resp, err := c.s.CreateDocument(r.Context(), &req)
	if err != nil {
		// handle error
	}

	body, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to json marshal response", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {

	}
}

func (c *controller) CreateFolderHandler(w http.ResponseWriter, r *http.Request) {

	var req service.CreateFolderRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		// handle error
	}

	resp, err := c.s.CreateFolder(r.Context(), &req)
	if err != nil {
		// handle error

		http.Error(w, "failed to create folder", http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to json marshal response", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {

	}
}

func (c *controller) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {

	var req service.CreateGroupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		// handle error
	}

	resp, err := c.s.CreateGroup(r.Context(), &req)
	if err != nil {
		log.Printf("failed to create group: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to json marshal response", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {

	}
}

func (c *controller) GetDocumentHandler(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	id := params["id"]

	resp, err := c.s.GetDocument(r.Context(), &service.GetDocumentRequest{
		ID: id,
	})
	if err != nil {
		if err == service.ErrUnauthorized {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	body, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to json marshal response", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {

	}
}

func (c *controller) GetDocumentsHandler(w http.ResponseWriter, r *http.Request) {

	resp, err := c.s.GetDocuments(r.Context())
	if err != nil {
		if err == service.ErrUnauthorized {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	body, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to json marshal response", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {

	}
}

func (c *controller) GetFolderHandler(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	id := params["id"]

	resp, err := c.s.GetFolder(r.Context(), &service.GetFolderRequest{
		ID: id,
	})
	if err != nil {
		if err == service.ErrUnauthorized {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	body, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to json marshal response", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {

	}
}

func (c *controller) GetFoldersHandler(w http.ResponseWriter, r *http.Request) {

	resp, err := c.s.GetFolders(r.Context())
	if err != nil {
		if err == service.ErrUnauthorized {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	body, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to json marshal response", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {

	}
}

func (c *controller) ShareHandler(w http.ResponseWriter, r *http.Request) {

	var req service.ShareObjectRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		// handle error
	}

	resp, err := c.s.ShareObject(r.Context(), &req)
	if err != nil {
		// handle error
	}

	body, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to json marshal response", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {

	}
}
