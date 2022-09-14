package docs

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
)

type service struct {
	db     *sql.DB
	router *mux.Router
}

func (s *service) createFolderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
