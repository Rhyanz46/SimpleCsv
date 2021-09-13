Example :

- Select Query

```go
now := time.Now().Format(time.RFC3339)
whereTo := SimpleCsv.Where{Field: "To", Operator: ">", Value: fmt.Sprintf("%v", now)}
whereImei := SimpleCsv.Where{Field: "Imei", Operator: "=", Value: imei}
err := csvData.Select("To").WhereValue(whereTo, "and", whereImei).One(&result)
```