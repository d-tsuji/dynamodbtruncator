package dynamodbtruncator

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/d-tsuji/dynamodbtruncator/testonly"
	"github.com/google/go-cmp/cmp"
)

var (
	db *DB

	testDBEndpoint = "http://localhost:4566"
)

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:   aws.String(endpoints.ApNortheast1RegionID),
			Endpoint: aws.String(testDBEndpoint),
		},
	}))
	db = New(sess)
}

func TestTable_Truncate(t *testing.T) {

	type TestTable struct {
		Hkey string `dynamodbav:"hkey"`
		Skey string `dynamodbav:"skey"`
	}

	tests := []struct {
		name            string
		tableName       string
		inputDataCmd    []string
		inputDataFunc   func(t *testing.T)
		cleanUpCmd      []string
		wantOutFilePath string
		wantErr         bool
	}{
		{
			name:      "all items will be deleted",
			tableName: "test-1",
			inputDataCmd: []string{
				"aws dynamodb --endpoint-url http://localhost:4566 create-table --cli-input-json file://./testdata/table1.json",
				"aws dynamodb --endpoint-url http://localhost:4566 batch-write-item --request-items file://./testdata/in1_1.json",
				"aws dynamodb --endpoint-url http://localhost:4566 batch-write-item --request-items file://./testdata/in1_2.json",
			},
			cleanUpCmd: []string{
				`aws dynamodb --endpoint-url http://localhost:4566 delete-table --table test-1`,
			},
			wantOutFilePath: filepath.Join("testdata", "out1.json"),
			wantErr:         false,
		},
		{
			name:      "many items",
			tableName: "test-1",
			inputDataCmd: []string{
				"aws dynamodb --endpoint-url http://localhost:4566 create-table --cli-input-json file://./testdata/table1.json",
			},
			inputDataFunc: func(t *testing.T) {
				for i := 0; i < 1000; i++ {
					item := TestTable{
						Hkey: "x",
						Skey: strconv.Itoa(i),
					}
					av, err := dynamodbattribute.MarshalMap(item)
					if err != nil {
						t.Fatalf("input data %v marshal fail: %v", item, err)
					}
					_, err = db.client.PutItemWithContext(context.TODO(), &dynamodb.PutItemInput{
						TableName: aws.String("test-1"),
						Item:      av,
					})
					if err != nil {
						t.Fatalf("input data %v create fail: %v", av, err)
					}
				}
			},
			cleanUpCmd: []string{
				`aws dynamodb --endpoint-url http://localhost:4566 delete-table --table test-1`,
			},
			wantOutFilePath: filepath.Join("testdata", "out1.json"),
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := db.Table(tt.tableName)
			testonly.CmdsExec(t, tt.inputDataCmd)
			if tt.inputDataFunc != nil {
				tt.inputDataFunc(t)
			}
			t.Cleanup(func() { testonly.CmdsExec(t, tt.cleanUpCmd) })

			start := time.Now()
			if err := table.Truncate(context.TODO()); (err != nil) != tt.wantErr {
				t.Errorf("Truncate() error = %v, wantErr %v", err, tt.wantErr)
			}
			elapsed := time.Since(start)
			t.Logf("Truncate() process time = %v", elapsed)

			want, err := ioutil.ReadFile(tt.wantOutFilePath)
			if err != nil {
				t.Errorf("read file error, path = %v: %v", tt.wantOutFilePath, err)
			}
			got := testonly.CmdExecCombinedOutput(t, fmt.Sprintf("aws dynamodb --endpoint-url http://localhost:4566 scan --table-name %s --no-cli-pager", tt.tableName))
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("items in table(%s) are mismatch (-want +got)\n%s", tt.tableName, diff)
			}
		})
	}
}

func TestTables_TruncateAll(t *testing.T) {
	tests := []struct {
		name                  string
		tableName             string
		inputDataCmd          []string
		cleanUpCmd            []string
		wantOutFilePathTables map[string]string
		wantErr               bool
	}{
		{
			name:      "all tables will be deleted",
			tableName: "test-2,test-3",
			inputDataCmd: []string{
				"aws dynamodb --endpoint-url http://localhost:4566 create-table --cli-input-json file://./testdata/table2.json",
				"aws dynamodb --endpoint-url http://localhost:4566 create-table --cli-input-json file://./testdata/table3.json",
				"aws dynamodb --endpoint-url http://localhost:4566 batch-write-item --request-items file://./testdata/in2.json",
				"aws dynamodb --endpoint-url http://localhost:4566 batch-write-item --request-items file://./testdata/in3.json",
			},
			cleanUpCmd: []string{
				`aws dynamodb --endpoint-url http://localhost:4566 delete-table --table test-2`,
				`aws dynamodb --endpoint-url http://localhost:4566 delete-table --table test-3`,
			},
			wantOutFilePathTables: map[string]string{
				filepath.Join("testdata", "out2.json"): "test-2",
				filepath.Join("testdata", "out3.json"): "test-3",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tables := db.Tables(tt.tableName)
			testonly.CmdsExec(t, tt.inputDataCmd)
			t.Cleanup(func() { testonly.CmdsExec(t, tt.cleanUpCmd) })

			if err := tables.TruncateAll(context.TODO()); (err != nil) != tt.wantErr {
				t.Errorf("Truncate() error = %v, wantErr %v", err, tt.wantErr)
			}

			for fpath, tableName := range tt.wantOutFilePathTables {
				want, err := ioutil.ReadFile(fpath)
				if err != nil {
					t.Errorf("read file error, path = %v: %v", fpath, err)
				}
				got := testonly.CmdExecCombinedOutput(t, fmt.Sprintf("aws dynamodb --endpoint-url http://localhost:4566 scan --table-name %s --no-cli-pager", tableName))
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("items in table(%s) are mismatch (-want +got):\n%s", tt.tableName, diff)
				}
			}
		})
	}
}
