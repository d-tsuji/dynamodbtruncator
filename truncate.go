// Package dynamodbtruncator truncates the DynamoDB tables.
package dynamodbtruncator

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"golang.org/x/sync/errgroup"
)

const (
	// DynamoDB API limit, 25 operations per request.
	maxBatchWriteOperationCount = 25
	maxRetryCount               = 3
)

// DB is a DynamoDB client.
type DB struct {
	client dynamodbiface.DynamoDBAPI
}

// New generates the DB from the AWS configuration.
func New(p client.ConfigProvider, cfgs ...*aws.Config) *DB {
	return &DB{dynamodb.New(p, cfgs...)}
}

// Tables handles multiple tables.
type Tables []Table

// Tables generates Tables from comma-separated table names.
func (db *DB) Tables(name string) Tables {
	names := strings.Split(name, ",")

	var tables Tables
	for _, n := range names {
		tables = append(tables, db.Table(strings.TrimSpace(n)))
	}
	return tables
}

// Table handles the DynamoDB table specified by name.
type Table struct {
	name string
	db   *DB
}

// Table generates the Table from the table name name.
func (db *DB) Table(name string) Table {
	return Table{
		name: name,
		db:   db,
	}
}

// TruncateAll truncates multiple tables in parallel.
func (ts Tables) TruncateAll(ctx context.Context) error {
	var eg errgroup.Group
	for _, t := range ts {
		// capture loop variable
		table := t
		eg.Go(func() error {
			if err := table.Truncate(ctx); err != nil {
				return fmt.Errorf("table %s truncate: %w", table.name, err)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("truncate all: %w", err)
	}
	return nil
}

// Truncate will truncate a specific table.
func (t Table) Truncate(ctx context.Context) error {
	keys, err := t.getKeys(ctx)
	if err != nil {
		return fmt.Errorf("get table keys: %w", err)
	}

	items, err := t.scan(ctx, keys)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	log.Printf("[%s] contains %d items\n", t.name, len(items))

	if err := t.batchDelete(ctx, items); err != nil {
		return fmt.Errorf("batch delete: %w", err)
	}

	log.Printf("[%s] complete to trucate table\n", t.name)

	return nil
}

func (t Table) scan(ctx context.Context, keys []string) ([]map[string]*dynamodb.AttributeValue, error) {
	var startKey map[string]*dynamodb.AttributeValue
	expressionAttributeNames := make(map[string]*string, len(keys))
	projectKeys := make([]string, 0, len(keys))

	for _, v := range keys {
		expressionAttributeNames["#"+v] = aws.String(v)
		projectKeys = append(projectKeys, "#"+v)
	}

	input := &dynamodb.ScanInput{
		ExclusiveStartKey:        startKey,
		ExpressionAttributeNames: expressionAttributeNames,
		ProjectionExpression:     aws.String(strings.Join(projectKeys, ",")),
		TableName:                aws.String(t.name),
	}

	items := make([]map[string]*dynamodb.AttributeValue, 0)
	err := t.db.client.ScanPagesWithContext(ctx, input, func(page *dynamodb.ScanOutput, lastPage bool) bool {
		items = append(items, page.Items...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (t Table) getKeys(ctx context.Context) ([]string, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(t.name),
	}

	desc, err := t.db.client.DescribeTableWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(desc.Table.KeySchema))
	for _, schema := range desc.Table.KeySchema {
		keys = append(keys, *schema.AttributeName)
	}
	return keys, nil
}

func (t Table) batchDelete(ctx context.Context, deletes []map[string]*dynamodb.AttributeValue) error {
	var items []*dynamodb.WriteRequest
	for _, v := range deletes {
		items = append(items, &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: v,
			},
		})

		if maxBatchWriteOperationCount <= len(items) {
			out, err := t.db.client.BatchWriteItemWithContext(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]*dynamodb.WriteRequest{
					t.name: items,
				},
			})
			if err != nil {
				return err
			}

			// Initialize the request after a successful write.
			items = items[:0]

			// If there are unprocessed items, reset them.
			remain := out.UnprocessedItems[t.name]
			if len(remain) > 0 {
				items = append(items, remain...)
			}
		}
	}

	for i := 0; i < maxRetryCount; i++ {
		if len(items) > 0 {
			out, err := t.db.client.BatchWriteItemWithContext(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]*dynamodb.WriteRequest{
					t.name: items,
				},
			})
			if err != nil {
				return err
			}

			// Initialize the request after a successful write.
			items = items[:0]

			// If there are unprocessed items, reset them.
			remain := out.UnprocessedItems[t.name]
			if len(remain) > 0 {
				items = append(items, remain...)
			}
		} else {
			break
		}
	}

	return nil
}
