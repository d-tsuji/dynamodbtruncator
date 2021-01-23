package testonly

type TestTable struct {
	Hkey string `dynamodbav:"hkey"`
	Skey string `dynamodbav:"skey"`
}
