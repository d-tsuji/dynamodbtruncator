package main

import (
	"errors"
	"log"
	"os"

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
				Name:  "region",
				Usage: "region name",
			},
			&cli.StringFlag{
				Name:  "table",
				Usage: "table name",
			},
		},
		Action: func(c *cli.Context) error {
			var cfg aws.Config
			regionName := c.String("region")
			if regionName != "" {
				cfg.Region= aws.String(regionName)
			}
			db := dynamodbtruncator.New(session.Must(session.NewSessionWithOptions(session.Options{
				Profile: c.String("profile"),
				Config: cfg,
				SharedConfigState: session.SharedConfigEnable,
			})))

			tableName := c.String("table")
			if tableName == "" {
				return errors.New(`table must be required. Please set "--table table1" or "--table table1,table2,...,table9"`)
			}
			return db.Tables(tableName).TruncateAll(c.Context)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
