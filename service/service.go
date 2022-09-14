package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	auth0fga "github.com/auth0-lab/fga-go-sdk"
	"github.com/google/uuid"
	"github.com/jon-whit/openfga-demo/middleware/auth"
)

type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Document struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Parent string `json:"parent"`
}

var ErrUnauthorized = fmt.Errorf("unauthorized")

type Service struct {
	//Datastore datastore.Store
	Database  *sql.DB
	FGAClient auth0fga.Auth0FgaApi
}

type CreateGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type CreateGroupResponse struct {
	ID string `json:"id"`
}

func (s *Service) CreateGroup(ctx context.Context, req *CreateGroupRequest) (*CreateGroupResponse, error) {

	id := uuid.NewString()

	var placeholders []string
	var values []interface{}

	var tuples []auth0fga.TupleKey
	for i, member := range req.Members {
		tuples = append(tuples, auth0fga.TupleKey{
			Object:   auth0fga.PtrString(fmt.Sprintf("group:%s", id)),
			Relation: auth0fga.PtrString("member"),
			User:     auth0fga.PtrString(member),
		})

		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d)", 3*i+1, 3*i+2, 3*i+3))
		values = append(values, id, req.Name, member)
	}

	stmt := fmt.Sprintf(`INSERT INTO groups (id, name, member) VALUES %s`, strings.Join(placeholders, ","))

	body := auth0fga.WriteRequest{
		Writes: &auth0fga.TupleKeys{
			TupleKeys: tuples,
		},
	}

	_, _, err := s.FGAClient.Write(ctx).Body(body).Execute()
	if err != nil {
		return nil, err
	}

	txn, err := s.Database.Begin()
	if err != nil {
		return nil, err
	}

	_, err = s.Database.Exec(stmt, values...)
	if err != nil {
		return nil, err
	}

	if err := txn.Commit(); err != nil {
		return nil, err
	}

	return &CreateGroupResponse{
		ID: id,
	}, nil
}

type CreateFolderRequest struct {
	Name string `json:"name"`
}

type CreateFolderResponse struct {
	ID string `json:"id"`
}

func (s *Service) CreateFolder(ctx context.Context, req *CreateFolderRequest) (*CreateFolderResponse, error) {

	authCtx, ok := auth.AuthContextFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	id := uuid.NewString()

	body := auth0fga.WriteRequest{
		Writes: &auth0fga.TupleKeys{
			TupleKeys: []auth0fga.TupleKey{
				{
					Object:   auth0fga.PtrString(fmt.Sprintf("folder:%s", id)),
					Relation: auth0fga.PtrString("owner"),
					User:     auth0fga.PtrString(authCtx.Subject),
				},
			},
		},
	}

	_, _, err := s.FGAClient.Write(ctx).Body(body).Execute()
	if err != nil {
		return nil, err
	}

	stmt := `INSERT INTO folders (id, name) VALUES ($1, $2)`
	_, err = s.Database.Exec(stmt, id, req.Name)
	if err != nil {
		return nil, err
	}

	return &CreateFolderResponse{
		ID: id,
	}, nil
}

type CreateDocumentRequest struct {
	Name   string `json:"name"`
	Parent string `json:"parent"`
}

type CreateDocumentResponse struct {
	ID string `json:"id"`
}

type GetFolderRequest struct {
	ID string `json:"id"`
}

type GetFolderResponse struct {
	Folder
}

func (s *Service) GetFolder(ctx context.Context, req *GetFolderRequest) (*GetFolderResponse, error) {

	authCtx, ok := auth.AuthContextFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	body := auth0fga.CheckRequest{
		TupleKey: &auth0fga.TupleKey{
			Object:   auth0fga.PtrString(fmt.Sprintf("folder:%s", req.ID)),
			Relation: auth0fga.PtrString("viewer"),
			User:     auth0fga.PtrString(authCtx.Subject),
		},
	}
	resp, _, err := s.FGAClient.Check(ctx).Body(body).Execute()
	if err != nil {
		// handle error
	}

	if !resp.GetAllowed() {
		return nil, ErrUnauthorized
	}

	row := s.Database.QueryRow(`SELECT id, name FROM folders WHERE id=$1`, req.ID)

	var id, name string
	err = row.Scan(&id, &name)
	if err != nil {
		return nil, err
	}

	return &GetFolderResponse{
		Folder: Folder{
			ID:   id,
			Name: name,
		},
	}, nil
}

func (s *Service) CreateDocument(ctx context.Context, req *CreateDocumentRequest) (*CreateDocumentResponse, error) {

	authCtx, ok := auth.AuthContextFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	id := uuid.NewString()

	tuples := []auth0fga.TupleKey{
		{
			Object:   auth0fga.PtrString(fmt.Sprintf("document:%s", id)),
			Relation: auth0fga.PtrString("owner"),
			User:     auth0fga.PtrString(authCtx.Subject),
		},
	}
	if req.Parent != "" {
		tuples = append(tuples, auth0fga.TupleKey{
			Object:   auth0fga.PtrString(fmt.Sprintf("document:%s", id)),
			Relation: auth0fga.PtrString("parent"),
			User:     auth0fga.PtrString(req.Parent),
		})
	}

	body := auth0fga.WriteRequest{
		Writes: &auth0fga.TupleKeys{
			TupleKeys: tuples,
		},
	}

	_, _, err := s.FGAClient.Write(ctx).Body(body).Execute()
	if err != nil {
		return nil, err
	}

	stmt := `INSERT INTO documents (id, name, parent) VALUES ($1, $2, $3)`
	_, err = s.Database.Exec(stmt, id, req.Name, req.Parent)
	if err != nil {
		return nil, err
	}

	return &CreateDocumentResponse{
		ID: id,
	}, nil
}

type GetDocumentRequest struct {
	ID string `json:"id"`
}

type GetDocumentResponse struct {
	Document
}

func (s *Service) GetDocument(ctx context.Context, req *GetDocumentRequest) (*GetDocumentResponse, error) {

	authCtx, ok := auth.AuthContextFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	body := auth0fga.CheckRequest{
		TupleKey: &auth0fga.TupleKey{
			Object:   auth0fga.PtrString(fmt.Sprintf("document:%s", req.ID)),
			Relation: auth0fga.PtrString("viewer"),
			User:     auth0fga.PtrString(authCtx.Subject),
		},
	}
	resp, _, err := s.FGAClient.Check(ctx).Body(body).Execute()
	if err != nil {
		// handle error
	}

	if !resp.GetAllowed() {
		return nil, ErrUnauthorized
	}

	row := s.Database.QueryRow(`SELECT id, name, parent FROM documents WHERE id=$1`, req.ID)

	var id, name, parent string
	err = row.Scan(&id, &name, &parent)
	if err != nil {
		return nil, err
	}

	return &GetDocumentResponse{
		Document: Document{
			ID:     id,
			Name:   name,
			Parent: parent,
		},
	}, nil
}

type ShareObjectRequest struct {
	UserID   string `json:"user"`
	Relation string `json:"relation"`
	Object   string `json:"object"`
}

type ShareObjectResponse struct{}

func (s *Service) ShareObject(ctx context.Context, req *ShareObjectRequest) (*ShareObjectResponse, error) {

	body := auth0fga.WriteRequest{
		Writes: &auth0fga.TupleKeys{
			TupleKeys: []auth0fga.TupleKey{
				{
					Object:   auth0fga.PtrString(req.Object),
					Relation: auth0fga.PtrString(req.Relation),
					User:     auth0fga.PtrString(req.UserID),
				},
			},
		},
	}

	_, _, err := s.FGAClient.Write(ctx).Body(body).Execute()
	if err != nil {
		return nil, err
	}

	return &ShareObjectResponse{}, nil
}
