package datastore

type DocumentStore interface {
}

type FolderStore interface {
}

type Store interface {
	DocumentStore
	FolderStore
}
