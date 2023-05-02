package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jon-whit/openfga-demo/middleware/auth"
	"github.com/openfga/go-sdk/client"
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
	FGAClient *client.OpenFgaClient
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

	var tuples []client.ClientTupleKey
	for i, member := range req.Members {
		tuples = append(tuples, client.ClientTupleKey{
			Object:   fmt.Sprintf("group:%s", id),
			Relation: "member",
			User:     member,
		})

		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d)", 3*i+1, 3*i+2, 3*i+3))
		values = append(values, id, req.Name, member)
	}

	stmt := fmt.Sprintf(`INSERT INTO groups (id, name, member) VALUES %s`, strings.Join(placeholders, ","))

	body := client.ClientWriteRequest{
		Writes: &tuples,
	}

	_, err := s.FGAClient.Write(ctx).Body(body).Execute()
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

	body := client.ClientWriteRequest{
		Writes: &[]client.ClientTupleKey{
			{
				Object:   fmt.Sprintf("folder:%s", id),
				Relation: "owner",
				User:     authCtx.Subject,
			},
		},
	}

	_, err := s.FGAClient.Write(ctx).Body(body).Execute()
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

	body := client.ClientCheckRequest{
		Object:   fmt.Sprintf("folder:%s", req.ID),
		Relation: "viewer",
		User:     authCtx.Subject,
	}
	resp, err := s.FGAClient.Check(ctx).Body(body).Execute()
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

	tuples := []client.ClientTupleKey{
		{
			Object:   fmt.Sprintf("document:%s", id),
			Relation: "owner",
			User:     authCtx.Subject,
		},
	}
	if req.Parent != "" {
		tuples = append(tuples, client.ClientTupleKey{
			Object:   fmt.Sprintf("document:%s", id),
			Relation: "parent",
			User:     req.Parent,
		})
	}

	body := client.ClientWriteRequest{
		Writes: &tuples,
	}

	_, err := s.FGAClient.Write(ctx).Body(body).Execute()
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

	body := client.ClientCheckRequest{
		Object:   fmt.Sprintf("document:%s", req.ID),
		Relation: "viewer",
		User:     authCtx.Subject,
	}
	resp, err := s.FGAClient.Check(ctx).Body(body).Execute()
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

	body := client.ClientWriteRequest{
		Writes: &[]client.ClientTupleKey{
			{
				Object:   req.Object,
				Relation: req.Relation,
				User:     req.UserID,
			},
		},
	}

	_, err := s.FGAClient.Write(ctx).Body(body).Execute()
	if err != nil {
		return nil, err
	}

	return &ShareObjectResponse{}, nil
}
