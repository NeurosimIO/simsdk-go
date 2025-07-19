package simsdk

type FieldType string

const (
	FieldString    FieldType = "string"
	FieldInt       FieldType = "int"
	FieldUint      FieldType = "uint"
	FieldFloat     FieldType = "float"
	FieldBool      FieldType = "bool"
	FieldEnum      FieldType = "enum"
	FieldTimestamp FieldType = "timestamp"
)
