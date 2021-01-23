package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/d-tsuji/dynamodbtruncator"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "dynamodbtruncator",
		Usage:     "Truncate table for DynamoDB",
		UsageText: "dynamodbtruncator [global options]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "profile",
				Usage: "profile name",
				Value: "default",
			},
			&cli.StringFlag{
				Name:  "table",
				Usage: "table name",
			},
		},
		Action: func(c *cli.Context) error {
			db := dynamodbtruncator.New(session.Must(session.NewSessionWithOptions(session.Options{
				Profile: c.String("profile"),
				Config: aws.Config{
					Region: aws.String(os.Getenv("AWS_REGION")),
				},
			})))

			tableStr := c.String("table")
			if tableStr == "" {
				return errors.New("table must be required. Please set --table [table1] or --table [table1,table2,...,table9]")
			}
			tables := dynamodbtruncator.Tables{}
			ts := strings.Split(tableStr, ",")
			for _, t := range ts {
				tables = append(tables, db.Table(strings.TrimSpace(t)))
			}
			return tables.TruncateAll(c.Context)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
