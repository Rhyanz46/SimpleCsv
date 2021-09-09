package SimpleCsv

import "reflect"

type csvHead struct {
	Index int
	Head  string
	Type  reflect.Type
}

type whereData struct {
	OperatorLogic []string
	List          []Where
}

type whereOperation struct {
	baseData         BaseData
	selectColumns    []string
	whereInitialized bool
	data             interface{}
	operationType    string
	whereData        whereData
}

type finalOperation struct {
	finalOperationState *whereOperation
}

type BaseData struct {
	Location      string
	Struct        interface{}
	headers       []string
	headersDetail []csvHead
	initialized   bool
}

type compareObject struct {
	Ori reflect.Type
	Now reflect.Type
}

type Where struct {
	Field string
	//FirstField   string
	// Operator it's only support for comparison operators and 'in', 'not in'
	Operator string
	Value    string
	//SecondField	 string
	//SecondFieldType string // optional default is 'value' and you just can use 'value' or 'field'
}
