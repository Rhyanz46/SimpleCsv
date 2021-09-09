package SimpleCsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

const NO_RESULT = "NO RESULT"

// roadmap
//
// - skip first line or not
// - in memory mode and put what is maximal time data in memory
// - get result data directly not using variable reference
// - Grouping
// - Join to another csv file
// - support uda and udf
// - Convert To XLSX
// - migrating to new schema
// - will using stream mode using channel to write and read
//
// One Function :
// if data get one it's stop scan

// WhereValue just accept Where struct and two logic operators 'and' and 'or'
// example :
// (where, 'and', where1)
// (where, 'or', where1)
// WhereValue checking all where value
// WhereValue checking for where data and return data
func (w *whereOperation) WhereValue(where ...interface{}) (result *finalOperation) {
	var operatorLogic []string
	var whereList []Where

	supportdLogicOptr := []string{"and", "or"}
	supportdCompOptr := []string{"<=", "<", ">", ">=", "<>", "=", "!="}
	for i, whereItem := range where {
		// struct
		if i%2 == 0 && reflect.TypeOf(whereItem).Kind() != reflect.Struct {
			msg := fmt.Sprintf("Parameter number %d should be using a struct", i+1)
			panic(msg)
		} else if i%2 == 0 {
			if reflect.TypeOf(Where{}).String() != reflect.TypeOf(whereItem).String() {
				msg := fmt.Sprintf("Parameter number %d should be using Where struct", i+1)
				panic(msg)
			}

			//check if field that exist in struct
			whereItemConverted := whereItem.(Where)
			notExist := FieldsNotInStruct([]string{whereItemConverted.Field}, w.baseData.Struct)
			notExistFieldsTotal := len(notExist)
			var fieldsNotExist string

			// looping all field that not exit and create message
			for j := 0; j < notExistFieldsTotal; j++ {
				if j+1 != notExistFieldsTotal {
					fieldsNotExist += fmt.Sprintf("%s, ", notExist[j])
				} else {
					fieldsNotExist += fmt.Sprintf("%s", notExist[j])
				}
			}

			// occur when the field is the one of we not expected to exist
			if notExist != nil {
				msg := fmt.Sprintf("Column '%v' is not in '%s' struct base data", fieldsNotExist, reflect.TypeOf(w.baseData.Struct))
				panic(msg)
			}

			var compOptrSupported bool
			for _, operator := range supportdCompOptr {
				if operator == whereItem.(Where).Operator {
					compOptrSupported = true
					break
				}
			}
			if !compOptrSupported {
				msg := fmt.Sprintf("'%s' is not supported as logic operator", whereItem.(Where).Operator)
				panic(msg)
			}

			whereList = append(whereList, whereItem.(Where))
		}

		// operator
		if i%2 == 1 && reflect.TypeOf(whereItem).Kind() != reflect.String {
			msg := fmt.Sprintf("Parameter number %d should be using string", i+1)
			panic(msg)
		} else if i%2 == 1 {
			var logicOptrSupported bool
			for _, operator := range supportdLogicOptr {
				if operator == whereItem {
					logicOptrSupported = true
					break
				}
			}
			if !logicOptrSupported {
				msg := fmt.Sprintf("at the moment, only support for two logic operator which is 'and' and 'or', so you cant use '%s'", whereItem)
				panic(msg)
			}
			operatorLogic = append(operatorLogic, whereItem.(string))
		}
		//break // at the moment not support for multiple where clause
	}
	w.whereData.List = whereList
	w.whereData.OperatorLogic = operatorLogic

	return &finalOperation{
		finalOperationState: w,
	}
}

func (w *whereOperation) WhereLineNumber(index int) (result *finalOperation) {
	return
}

//All will scan all line and
func (w *finalOperation) All(dist interface{}) error {
	fmt.Println("total field", len(w.finalOperationState.baseData.headers))
	//var lineFilter [][]string
	destObject := isCorrectObjectPointer(w.finalOperationState.baseData.Struct, dist)
	switch destObject.Now.Kind() {
	case reflect.Slice:
		goto checkSlice
	default:
		panic(fmt.Sprintf("in 'All' operation you should be use slice not %s", destObject.Now.Kind()))
	}
checkSlice:
	data := selectSearch(w, &dist)
	fmt.Println(data)
	return nil
}

func (w *finalOperation) One(dist interface{}) error {
	destObject := isCorrectObjectPointer(w.finalOperationState.baseData.Struct, dist)
	switch destObject.Now.Kind() {
	case reflect.Struct:
		goto checkStruct
	default:
		panic(fmt.Sprintf("in 'One' operation you should be use struct not %s", destObject.Now.Kind()))
	}
checkStruct:
	err := selectSearch(w, &dist)
	if err != nil {
		return err
	}
	return nil
}

// Insert append new line
func (c *BaseData) Insert(data interface{}) (err error) {
	//var header []string
	dataValue := reflect.ValueOf(data)
	object := isCorrectObject(c.Struct, data)
	fieldNumber := object.Ori.NumField()

	// open file
	file, err := os.OpenFile(c.Location, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		panic(err)
	}

	// set file to csv object
	writer := *csv.NewWriter(file)
	reader := *csv.NewReader(file)

	// close
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)
	defer writer.Flush()

	// write header if not exist
	err = writeHeader(c.headers, &reader, &writer)
	if err != nil {
		panic(err)
	}

	switch object.Now.Kind() {
	// operation on slice or array
	case reflect.Slice, reflect.Array:

		// looping object on array or slice
		for i := 0; i < dataValue.Len(); i++ {

			// insert all data into string slice
			var rowForWrite []string
			for j := 0; j < fieldNumber; j++ {
				//fieldName := object.Ori.Field(j).Name
				//fmt.Print(fieldName, "(", dataValue.Index(i).Field(j).Type().String(), ") ")
				switch dataValue.Index(i).Field(j).Kind() {
				case reflect.Int:
					intValue := strconv.Itoa(int(dataValue.Index(i).Field(j).Int()))
					rowForWrite = append(rowForWrite, intValue)
				case reflect.Bool:
					if dataValue.Index(i).Field(j).Bool() {
						rowForWrite = append(rowForWrite, "true")
					} else {
						rowForWrite = append(rowForWrite, "false")
					}
				case reflect.String:
					rowForWrite = append(rowForWrite, dataValue.Index(i).Field(j).String())
				case reflect.Struct:
					if dataValue.Index(i).Field(j).Type().String() == "time.Time" {
						//date value in time.RFC3339
						timeValue := dataValue.Index(i).Field(j).Interface().(time.Time).Format(time.RFC3339)
						rowForWrite = append(rowForWrite, timeValue)
					}
				}
				//fmt.Print(dataValue.Index(i).Field(j).String())
			}
			//fmt.Println()

			// write csv from string slice
			err = writer.Write(rowForWrite)
			if err != nil {
				msg := fmt.Sprintf("%v", err)
				return errors.New(msg)
			}
		}

	// operation on struct
	default:

		// insert all data into string slice
		var rowForWrite []string
		for i := 0; i < fieldNumber; i++ {
			//fieldName := object.Ori.Field(i).Name
			//fmt.Print(fieldName, "(", dataValue.Field(i).Type().String(), ") ")
			switch dataValue.Field(i).Kind() {
			case reflect.Int:
				intValue := strconv.Itoa(int(dataValue.Field(i).Int()))
				rowForWrite = append(rowForWrite, intValue)
			case reflect.Bool:
				if dataValue.Field(i).Bool() {
					rowForWrite = append(rowForWrite, "true")
				} else {
					rowForWrite = append(rowForWrite, "false")
				}
			case reflect.String:
				rowForWrite = append(rowForWrite, dataValue.Field(i).String())
			case reflect.Struct:
				if dataValue.Field(i).Type().String() == "time.Time" {
					//date value in time.RFC3339
					timeValue := dataValue.Field(i).Interface().(time.Time).Format(time.RFC3339)
					rowForWrite = append(rowForWrite, timeValue)
				}
			}
			//fmt.Print(dataValue.Index(i).Field(j).String())
		}

		// write csv from string slice
		err = writer.Write(rowForWrite)
		if err != nil {
			msg := fmt.Sprintf("%v", err)
			return errors.New(msg)
		}
	}

	if !c.initialized {
		panic("You must declare this first")
	}
	return
}

// Replace for replace all content
func (c *BaseData) Replace(data interface{}) (err error) {
	dataValue := reflect.ValueOf(data)
	object := isCorrectObject(c.Struct, data)
	fieldNumber := object.Ori.NumField()

	// replace file
	file, err := os.OpenFile(c.Location, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		panic(err)
	}

	// set file to csv object
	writer := *csv.NewWriter(file)
	reader := *csv.NewReader(file)

	// close
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)
	defer writer.Flush()

	// write header if not exist
	err = writeHeader(c.headers, &reader, &writer)
	if err != nil {
		panic(err)
	}

	switch object.Now.Kind() {
	// operation on slice or array
	case reflect.Slice, reflect.Array:

		// looping object on array or slice
		for i := 0; i < dataValue.Len(); i++ {

			// insert all data into string slice
			var rowForWrite []string
			for j := 0; j < fieldNumber; j++ {
				//fieldName := object.Ori.Field(j).Name
				//fmt.Print(fieldName, "(", dataValue.Index(i).Field(j).Type().String(), ") ")
				switch dataValue.Index(i).Field(j).Kind() {
				case reflect.Int:
					intValue := strconv.Itoa(int(dataValue.Index(i).Field(j).Int()))
					rowForWrite = append(rowForWrite, intValue)
				case reflect.Bool:
					if dataValue.Index(i).Field(j).Bool() {
						rowForWrite = append(rowForWrite, "true")
					} else {
						rowForWrite = append(rowForWrite, "false")
					}
				case reflect.String:
					rowForWrite = append(rowForWrite, dataValue.Index(i).Field(j).String())
				case reflect.Struct:
					if dataValue.Index(i).Field(j).Type().String() == "time.Time" {
						//date value in time.RFC3339
						timeValue := dataValue.Index(i).Field(j).Interface().(time.Time).Format(time.RFC3339)
						rowForWrite = append(rowForWrite, timeValue)
					}
				}
				//fmt.Print(dataValue.Index(i).Field(j).String())
			}
			//fmt.Println()

			// write csv from string slice
			err = writer.Write(rowForWrite)
			if err != nil {
				msg := fmt.Sprintf("%v", err)
				return errors.New(msg)
			}
		}

	// operation on struct
	default:

		// insert all data into string slice
		var rowForWrite []string
		for i := 0; i < fieldNumber; i++ {
			//fieldName := object.Ori.Field(i).Name
			//fmt.Print(fieldName, "(", dataValue.Field(i).Type().String(), ") ")
			switch dataValue.Field(i).Kind() {
			case reflect.Int:
				intValue := strconv.Itoa(int(dataValue.Field(i).Int()))
				rowForWrite = append(rowForWrite, intValue)
			case reflect.Bool:
				if dataValue.Field(i).Bool() {
					rowForWrite = append(rowForWrite, "true")
				} else {
					rowForWrite = append(rowForWrite, "false")
				}
			case reflect.String:
				rowForWrite = append(rowForWrite, dataValue.Field(i).String())
			case reflect.Struct:
				if dataValue.Field(i).Type().String() == "time.Time" {
					//date value in time.RFC3339
					timeValue := dataValue.Field(i).Interface().(time.Time).Format(time.RFC3339)
					rowForWrite = append(rowForWrite, timeValue)
				}
			}
			//fmt.Print(dataValue.Index(i).Field(j).String())
		}

		// write csv from string slice
		err = writer.Write(rowForWrite)
		if err != nil {
			msg := fmt.Sprintf("%v", err)
			return errors.New(msg)
		}
	}

	if !c.initialized {
		panic("You must declare this first")
	}
	return
}

// Select is checking list of columns it's match
func (c *BaseData) Select(columns ...string) *whereOperation {
	// *, udf, and uda will support in future
	if columns[0] == "*" && len(columns) > 1 {
		panic("if you using `*` you cant put others column")
	}
	if columns[0] == "*" {
		return &whereOperation{
			selectColumns:    c.headers,
			baseData:         *c,
			whereInitialized: true,
		}
	}
	if !c.initialized {
		panic("You must declare this first")
	}
	var fields string
	notExistFields := FieldsNotInStruct(columns, c.Struct)
	notExistFieldsTotal := len(notExistFields)

	for i := 0; i < notExistFieldsTotal; i++ {
		if i+1 != notExistFieldsTotal {
			fields += fmt.Sprintf("%s, ", notExistFields[i])
		} else {
			fields += fmt.Sprintf("%s", notExistFields[i])
		}
	}
	if notExistFields != nil {
		msg := fmt.Sprintf("Column '%v' is not in '%s' struct base data", fields, reflect.TypeOf(c.Struct))
		panic(msg)
	}
	return &whereOperation{
		selectColumns:    columns,
		baseData:         *c,
		whereInitialized: true,
	}
}

//func (c *BaseData) Update(data interface{}) *whereOperation {
//	if !c.initialized {
//		panic("You must declare this first")
//	}
//	return &whereOperation{
//		baseData:              *c,
//		data:             data,
//		whereInitialized: true,
//	}
//}
