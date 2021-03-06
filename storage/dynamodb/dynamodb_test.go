package dynamodb

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"github.com/craigpastro/crudapp/telemetry"
	"github.com/stretchr/testify/require"
)

const data = "some data"

var (
	db storage.Storage
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	client, err := CreateClient(ctx, Config{Region: "us-west-2", Local: true})
	if err != nil {
		log.Fatal(err)
	}

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(userIDAttribute),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(postIDAttribute),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(userIDAttribute),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(postIDAttribute),
				KeyType:       aws.String("RANGE"),
			},
		},
		BillingMode: aws.String("PAY_PER_REQUEST"),
	}
	if _, err := client.CreateTableWithContext(ctx, input); err != nil {
		if !strings.Contains(err.Error(), "Cannot create preexisting table") {
			log.Fatalf("error creating table: %v\n", err)
		}
	}

	db = New(client, telemetry.NewNoopTracer())

	os.Exit(m.Run())
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	created, err := db.Create(ctx, userID, data)
	require.NoError(t, err)
	record, err := db.Read(ctx, created.UserID, created.PostID)
	require.NoError(t, err)

	require.Equal(t, record.UserID, created.UserID, "got '%s', want '%s'", record.UserID, userID)
	require.Equal(t, record.PostID, created.PostID, "got '%s', want '%s'", record.PostID, created.PostID)
	require.Equal(t, record.Data, data, "got '%s', want '%s'", record.Data, data)
}

func TestReadNotExists(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()

	_, err := db.Read(ctx, userID, "1")
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func TestReadAll(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	_, err := db.Create(ctx, userID, "data 1")
	require.NoError(t, err)
	_, err = db.Create(ctx, userID, "data 2")
	require.NoError(t, err)

	records, err := db.ReadAll(ctx, userID)
	require.NoError(t, err)

	require.Len(t, records, 2, "got '%d', want '%d'", len(records), 2)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	created, err := db.Create(ctx, userID, data)
	require.NoError(t, err)

	newData := "new data"
	_, err = db.Update(ctx, userID, created.PostID, newData)
	require.NoError(t, err)
	record, err := db.Read(ctx, created.UserID, created.PostID)
	require.NoError(t, err)

	require.Equal(t, record.Data, newData, "got '%s', want '%s'")
	require.True(t, record.CreatedAt.Before(record.UpdatedAt))
}

func TestUpdateNotExists(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()

	_, err := db.Update(ctx, userID, "1", "new data")
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	created, _ := db.Create(ctx, userID, data)

	err := db.Delete(ctx, userID, created.PostID)
	require.NoError(t, err)

	// Now try to read the deleted record; it should not exist.
	_, err = db.Read(ctx, userID, created.PostID)
	require.ErrorIs(t, err, storage.ErrPostDoesNotExist)
}

func TestDeleteNotExists(t *testing.T) {
	ctx := context.Background()
	userID := myid.New()
	postID := myid.New()

	err := db.Delete(ctx, userID, postID)
	require.NoError(t, err)
}
