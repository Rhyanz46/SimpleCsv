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

func NewCsvData(c BaseData) *BaseData {
	// if file not exist create new one
	if _, err := os.Stat(c.Location); err != nil && err.Error() == "stat "+c.Location+": no such file or directory" {
		f, e := os.Create(c.Location)
		if e != nil {
			panic(e)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}(f)
	}

	// field of struct just support string, int, and boolean
	if c.initialized {
		panic("Cannot use New twice")
	}
	if reflect.TypeOf(c.Struct).Kind() != reflect.Struct {
		panic("It's should be struct")
	}
	// get header
	for i := 0; i < reflect.TypeOf(c.Struct).NumField(); i++ {
		switch reflect.TypeOf(c.Struct).Field(i).Type.Kind() {
		case reflect.String, reflect.Int, reflect.Bool, reflect.Struct:
			fieldName := reflect.TypeOf(c.Struct).Field(i).Name
			c.headers = append(c.headers, fieldName)
			c.headersDetail = append(
				c.headersDetail,
				csvHead{
					Index: i,
					Type:  reflect.TypeOf(c.Struct).Field(i).Type,
					Head:  fieldName,
				},
			)
		default:
			msg := fmt.Sprintf("'%s' is not support for csv file row data", reflect.TypeOf(c.Struct).Field(i).Type.Kind())
			panic(msg)
		}
	}
	c.initialized = true
	return &c
}

func stringInSliceString(str string, sliceString []string) (exist bool) {
	for _, item := range sliceString {
		if item == str {
			return true
		}
	}
	return
}

func intInSliceInt(integer int, sliceInt []int) (exist bool) {
	for _, item := range sliceInt {
		if item == integer {
			return true
		}
	}
	return
}

func FieldsNotInStruct(fields []string, structObj interface{}) (notExist []string) {
	var fieldsInStruct []string
	for i := 0; i < reflect.TypeOf(structObj).NumField(); i++ {
		fieldsInStruct = append(fieldsInStruct, reflect.TypeOf(structObj).Field(i).Name)
	}
	for _, column := range fields {
		if !stringInSliceString(column, fieldsInStruct) {
			notExist = append(notExist, column)
		}
	}
	return
}

func writeHeader(header []string, reader *csv.Reader, writer *csv.Writer) (err error) {
	_, err = reader.Read()
	if err != nil {
		if err.Error() == "EOF" {
			err = writer.Write(header)
			if err != nil {
				msg := fmt.Sprintf("%v", err)
				return errors.New(msg)
			}
		} else {
			msg := fmt.Sprintf("%v", err)
			return errors.New(msg)
		}
	}
	return
}

// isCorrectObject to Correcting between two struct
func isCorrectObject(ori interface{}, now interface{}) compareObject {
	rvNow := reflect.TypeOf(now)
	rvFrom := reflect.TypeOf(ori)
	typeData := reflect.Slice

	nowValue := reflect.ValueOf(now)
	if nowValue.Kind() == reflect.Ptr {
		msg := fmt.Sprintf("expected value but got %T pointer", rvNow)
		panic(msg)
	}

	switch rvNow.Kind() {
	case reflect.Struct:
		if rvFrom.String() != rvNow.String() {
			msg := fmt.Sprintf("The data should be using %s struct", rvFrom.String())
			panic(msg)
		}
		return compareObject{
			Ori: rvFrom,
			Now: rvNow,
		}
	case reflect.Array:
		typeData = reflect.Array
		goto checkIterLocation
	case reflect.Slice:
		goto checkIterLocation
	default:
		panic(fmt.Sprintf("The data should be struct/Array/Slice not %s", rvNow))
	}
checkIterLocation:
	dataPkg := fmt.Sprintf("[]%s", rvFrom.String())
	dataValue := reflect.ValueOf(now)
	if dataValue.Len() == 0 {
		panic("data is empty")
	}
	// it's magic
	if typeData == reflect.Array {
		dataPkg = fmt.Sprintf("[%d]%s", dataValue.Len(), rvFrom.String())
	}
	if rvNow.String() != dataPkg {
		msg := fmt.Sprintf("The data iteration should be using %s struct", rvFrom.String())
		panic(msg)
	}
	return compareObject{
		Ori: rvFrom,
		Now: rvNow,
	}
}

func isCorrectObjectPointer(ori interface{}, now interface{}) compareObject {
	rvNow := reflect.Indirect(reflect.ValueOf(now))
	rvFrom := reflect.TypeOf(ori)
	typeData := reflect.Slice

	// it's should be pointer
	ptrValueNow := reflect.ValueOf(now)
	if ptrValueNow.Kind() != reflect.Ptr {
		msg := fmt.Sprintf("Expected %s struct but got %s", rvFrom.String(), rvNow.String())
		panic(msg)
	}

	switch rvNow.Kind() {
	case reflect.Struct:
		if rvFrom.String() != rvNow.Type().String() {
			msg := fmt.Sprintf("Expected %s struct but got %s", rvFrom.String(), rvNow.String())
			panic(msg)
		}
		return compareObject{
			Ori: rvFrom,
			Now: rvNow.Type(),
		}
	case reflect.Array:
		typeData = reflect.Array
		goto checkIterLocation
	case reflect.Slice:
		goto checkIterLocation
	default:
		panic(fmt.Sprintf("Expected struct/Array/Slice but got %s", rvNow.Type().String()))
	}
checkIterLocation:
	dataPkg := fmt.Sprintf("[]%s", rvFrom.String())
	// it's magic
	if typeData == reflect.Array {
		dataPkg = fmt.Sprintf("[%d]%s", rvNow.Len(), rvFrom.String())
	}

	if rvNow.Type().String() != dataPkg {
		msg := fmt.Sprintf("expected '%s' but got '%s'", dataPkg, rvNow.Type().String())
		panic(msg)
	}
	return compareObject{
		Ori: rvFrom,
		Now: rvNow.Type(),
	}
}

func setValue(selectedIndex []int, data []string, dist interface{}) {
	destObject := reflect.Indirect(reflect.ValueOf(dist))
	for i, index := range selectedIndex {
		switch destObject.Elem().Elem().Field(index).Kind() {
		case reflect.String:
			destObject.Elem().Elem().Field(index).SetString(data[i])
		case reflect.Int:
			val, _ := strconv.Atoi(data[i])
			destObject.Elem().Elem().Field(index).SetInt(int64(val))
		case reflect.Bool:
			if data[i] == "true" {
				destObject.Elem().Elem().Field(index).SetBool(true)
			} else {
				destObject.Elem().Elem().Field(index).SetBool(false)
			}
		case reflect.Struct:
			// just support date.Time
			//if destObject.Elem().Elem().Field(index).Type().Name()
			dateResult, err := time.Parse(time.RFC3339, data[i])
			if err != nil {
				msg := fmt.Sprintf("Error while parsing date : %v", data[i])
				panic(msg)
			}
			v := reflect.ValueOf(dateResult)
			destObject.Elem().Elem().Field(index).Set(v)
		}
	}
}

func selectSearch(w *finalOperation, dist interface{}) error {
	file, err := os.OpenFile(w.finalOperationState.baseData.Location, os.O_APPEND|os.O_RDONLY, os.ModeAppend)
	var indexProcess, indexSelected, index []int
	var exist bool
	if err != nil {
		panic(err)
	}

	// close file
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	// set file to csv object
	reader := *csv.NewReader(file)

	// set index selected
	for i, head := range w.finalOperationState.baseData.headersDetail {
		if stringInSliceString(head.Head, w.finalOperationState.selectColumns) {
			indexProcess = append(indexProcess, head.Index)
			indexSelected = append(indexSelected, head.Index)
		} else {
			for _, wheredata := range w.finalOperationState.whereData.List {
				if head.Head == wheredata.Field {
					indexProcess = append(indexProcess, head.Index)
				}
			}
		}
		index = append(index, i)
	}

	totalWhereProcess := len(w.finalOperationState.whereData.List)
	indexWhereProcess := 0
	// read all file
	all, err := reader.ReadAll()
	if err != nil {
		return err
	}

processWhere:
	var result [][]string
	where := w.finalOperationState.whereData.List[indexWhereProcess]
	//"<=", "<", ">", ">=", "<>", "=", "!="

	// debug
	//fmt.Println()
	//fmt.Println("proses ke ", indexWhereProcess, " - ", where)
	//fmt.Println("dengan data : ")
	//for _, strings := range all {
	//	fmt.Println(strings)
	//}
	//fmt.Println("\nhasilnya adalah : ")
	// debug

	for i, line := range all {
		// skip header
		if i == 0 && indexWhereProcess == 0 {
			continue
		}

		// one where clause
		indexWhere := func() int {
			for j, head := range w.finalOperationState.baseData.headersDetail {
				if head.Head == where.Field {
					return j
				}
			}
			return -1
		}()
		var dateResultField, dateResultValue time.Time
		var intResultField, intResultValue int
		var boolField, boolValue bool

		// when field is date.Time
		if w.finalOperationState.baseData.headersDetail[indexWhere].Type.String() == "time.Time" {
			dateResultField, err = time.Parse(time.RFC3339, line[indexWhere])
			if err != nil {
				msg := fmt.Sprintf("Error while parsing date : %v", err)
				return errors.New(msg)
			}
			dateResultValue, err = time.Parse(time.RFC3339, where.Value)
			if err != nil {
				msg := fmt.Sprintf("Error while parsing date : %v", err)
				return errors.New(msg)
			}
		}

		// when field type is int,string,boolean
		switch w.finalOperationState.baseData.headersDetail[indexWhere].Type.Kind() {
		case reflect.Int:
			intResultField, err = strconv.Atoi(line[indexWhere])
			if err != nil {
				msg := fmt.Sprintf("Error while parsing int : %v", err)
				return errors.New(msg)
			}
			intResultValue, err = strconv.Atoi(where.Value)
			if err != nil {
				msg := fmt.Sprintf("Error while parsing int : %v", err)
				return errors.New(msg)
			}
		case reflect.Bool:
			if line[indexWhere] == "true" {
				boolField = true
			}
			if where.Value == "true" {
				boolValue = true
			}
		}

		// will skip if no match any value on line
		switch where.Operator {
		case "<=":
			panic("On Development")
		case "<":
			switch w.finalOperationState.baseData.headersDetail[indexWhere].Type.String() {
			case "time.Time":
				if dateResultField.String() == dateResultValue.String() {
					continue
				}
				if dateResultField.After(dateResultValue) {
					continue
				}
			case "string":
				if len(line[indexWhere]) < len(where.Value) {
					continue
				}
			case "int":
				if intResultField >= intResultValue {
					continue
				}
			default:
				panic("On Development")
			}
		case ">":
			switch w.finalOperationState.baseData.headersDetail[indexWhere].Type.String() {
			case "time.Time":
				if dateResultField.String() == dateResultValue.String() {
					continue
				}
				if dateResultField.Before(dateResultValue) {
					continue
				}
			case "string":
				if len(line[indexWhere]) < len(where.Value) {
					continue
				}
			case "int":
				if intResultField <= intResultValue {
					continue
				}
			default:
				panic("On Development")
			}
		case ">=":
			panic("On Development")
		case "<>", "!=":
			if line[indexWhere] == where.Value {
				continue
			}
		case "=":
			switch w.finalOperationState.baseData.headersDetail[indexWhere].Type.String() {
			case "time.Time":
				if dateResultField.String() != dateResultValue.String() {
					continue
				}
			case "string":
				if line[indexWhere] != where.Value {
					continue
				}
			case "int":
				if intResultField != intResultValue {
					continue
				}
			case "bool":
				if boolField != boolValue {
					continue
				}
			default:
				panic("On Development")
			}
		}
		result = append(result, line)
	}

	if totalWhereProcess-1 == 0 {
		goto returnProcess
	}
	indexWhereProcess++
	totalWhereProcess--
	all = result
	goto processWhere
returnProcess:
	// select field
	//var rowData []string // to select data
	////var headSelected []interface{}
	//for _,selected := range indexProcess{
	//	rowData = append(rowData, line[selected])
	//	//headSelected = append(headSelected, w.finalOperationState.baseData.headers[selected])
	//}
	//result = append(result, rowData)
	// select field

	for _, rowData := range result {
		fmt.Println("rowData", rowData)
		exist = true
		setValue(index, rowData, dist)
	}

	if !exist {
		return errors.New(NO_RESULT)
	}

	return nil
}
