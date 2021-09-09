package SimpleCsv

type finalOperationType interface {
	All(interface{}) error
	One(interface{}) error
}

type whereOperationType interface {
	WhereValue(...interface{}) *finalOperation
	WhereLineNumber(int) *finalOperation
}

type csvDataType interface {
	Insert(interface{}) error
	Replace(interface{}) error
	Select(...string) *whereOperation
	//Delete(interface{}) *whereOperation
	//Update(interface{}) *whereOperation
	//Delete(interface{}) error
	//Find(interface{}) interface{}
	//FindAll(interface{}) []interface{}
}
