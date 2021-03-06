package memory

import (
	"context"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/craigpastro/crudapp/cache"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/telemetry"
	lru "github.com/hashicorp/golang-lru"
)

const data = "some data"

var (
	c cache.Cache
)

func TestMain(m *testing.M) {
	store, err := lru.New(10)
	if err != nil {
		log.Fatalf("error creating cache: %v\n", err)
	}

	c = &Memory{
		store:  store,
		tracer: telemetry.NewNoopTracer(),
	}
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)
	c.Add(ctx, userID, postID, record)
	gotRecord, ok := c.Get(ctx, userID, postID)

	if ok != true {
		t.Error("did not get record")
	}

	if !reflect.DeepEqual(gotRecord, record) {
		t.Errorf("gotRecord is not the same as added record: got=%v, added=%v", gotRecord, record)
	}
}

func TestRemove(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)
	c.Add(ctx, userID, postID, record)
	if _, ok := c.Get(ctx, userID, postID); ok == false {
		t.Error("error inserting record")
	}

	c.Remove(ctx, userID, postID)
	if _, ok := c.Get(ctx, userID, postID); ok == true {
		t.Error("error removing record")
	}
}
