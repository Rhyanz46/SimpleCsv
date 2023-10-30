Overview
this is ORM library for query data inside CSV file 

Installation :
```
go get github.com/Rhyanz46/SimpleCsv
```

Example :

- Create new data
```go
type ShareLocTime struct {
    Imei         string
    TimeInMinute string
    From         time.Time
    To           time.Time
}

csvData := SimpleCsv.NewCsvData(SimpleCsv.BaseData{
    Location: "./data.csv",
    Struct:   ShareLocTime{},
})
```

- Select Query

```go
var result ShareLocTime
now := time.Now().Format(time.RFC3339)
whereTo := SimpleCsv.Where{Field: "To", Operator: ">", Value: fmt.Sprintf("%v", now)}
whereImei := SimpleCsv.Where{Field: "Imei", Operator: "=", Value: imei}
err := csvData.Select("To").WhereValue(whereTo, "and", whereImei).One(&result)

if err != nil {
    if err.Error() == SimpleCsv.NO_RESULT {
        ....
    }
    return
}

fmt.println(result)
```
