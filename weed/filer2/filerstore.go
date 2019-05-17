package filer2

import (
	"context"
	"errors"
	"github.com/chrislusf/seaweedfs/weed/pb/filer_pb"
	"github.com/chrislusf/seaweedfs/weed/util"
)

type FilerStore interface {
	// GetName gets the name to locate the configuration in filer.toml file
	GetName() string
	// Initialize initializes the file store
	Initialize(configuration util.Configuration) error
	InsertEntry(context.Context, *Entry) error
	UpdateEntry(context.Context, *Entry) (err error)
	// err == filer2.ErrNotFound if not found
	FindEntry(context.Context, FullPath) (entry *Entry, err error)
	DeleteEntry(context.Context, FullPath) (err error)
	ListDirectoryEntries(ctx context.Context, dirPath FullPath, startFileName string, includeStartFile bool, limit int) ([]*Entry, error)

	BeginTransaction(ctx context.Context) (context.Context, error)
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error
}

var ErrNotFound = errors.New("filer: no entry is found in filer store")

type FilerStoreWrapper struct {
	actualStore FilerStore
}

func NewFilerStoreWrapper(store FilerStore) *FilerStoreWrapper{
	return &FilerStoreWrapper{
		actualStore:store,
	}
}

func (fsw *FilerStoreWrapper) GetName() string {
	return fsw.actualStore.GetName()
}

func (fsw *FilerStoreWrapper) Initialize(configuration util.Configuration) error {
	return fsw.actualStore.Initialize(configuration)
}

func (fsw *FilerStoreWrapper) InsertEntry(ctx context.Context, entry *Entry) error {
	filer_pb.BeforeEntrySerialization(entry.Chunks)
	return fsw.actualStore.InsertEntry(ctx, entry)
}

func (fsw *FilerStoreWrapper) UpdateEntry(ctx context.Context, entry *Entry) error {
	filer_pb.BeforeEntrySerialization(entry.Chunks)
	return fsw.actualStore.UpdateEntry(ctx, entry)
}

func (fsw *FilerStoreWrapper) FindEntry(ctx context.Context, fp FullPath) (entry *Entry, err error) {
	entry, err = fsw.actualStore.FindEntry(ctx, fp)
	filer_pb.AfterEntryDeserialization(entry.Chunks)
	return
}

func (fsw *FilerStoreWrapper) DeleteEntry(ctx context.Context, fp FullPath) (err error) {
	return fsw.actualStore.DeleteEntry(ctx, fp)
}

func (fsw *FilerStoreWrapper) ListDirectoryEntries(ctx context.Context, dirPath FullPath, startFileName string, includeStartFile bool, limit int) ([]*Entry, error) {
	entries, err := fsw.actualStore.ListDirectoryEntries(ctx, dirPath, startFileName, includeStartFile, limit)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		filer_pb.AfterEntryDeserialization(entry.Chunks)
	}
	return entries, err
}

func (fsw *FilerStoreWrapper) BeginTransaction(ctx context.Context) (context.Context, error) {
	return fsw.actualStore.BeginTransaction(ctx)
}

func (fsw *FilerStoreWrapper) CommitTransaction(ctx context.Context) error {
	return fsw.actualStore.CommitTransaction(ctx)
}

func (fsw *FilerStoreWrapper) RollbackTransaction(ctx context.Context) error {
	return fsw.actualStore.RollbackTransaction(ctx)
}
