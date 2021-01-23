dynamodbtruncator (Truncate table for DynamoDB)
===============================================

[![Test Status](https://github.com/d-tsuji/dynamodbtruncator/workflows/test/badge.svg?branch=master)][actions]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![Go Report Card](https://goreportcard.com/badge/github.com/d-tsuji/dynamodbtruncator)][report]

[actions]: https://github.com/d-tsuji/dynamodbtruncator/actions?workflow=test
[license]: https://github.com/d-tsuji/dynamodbtruncator/blob/main/LICENSE
[report]: https://goreportcard.com/report/github.com/d-tsuji/dynamodbtruncator

`dynamodbtruncator` truncate tables for DynamoDB.

## Usage

```console
$ dynamodbtruncator [options]
```

### Options

```
--profile string
	The name of the profile from which the session can be obtained (default `default`)

--table string
	Trucated tables. Multiple tables can be specified separated by commas.
	e.g. table or table1,table2,table3
```

### Example

```
$ dynamodbtruncator --table hoge-table --profile my-profile
```

## Installation

- From source code

```
# go get
$ go get github.com/d-tsuji/dynamodbtruncator/cmd/dynamodbtruncator
```

- From binary

```
# binary
$ curl -sfL https://raw.githubusercontent.com/d-tsuji/dynamodbtruncator/master/install.sh | sudo sh -s -- -b /usr/local/bin
```
