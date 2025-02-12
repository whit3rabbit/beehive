package models

type TaskParameter struct {
	Type  string      `json:"type"`  // "string", "number", "boolean"
	Value interface{} `json:"value"` // Actual value, type depends on "Type"
}

type TaskOutput struct {
	Type  string      `json:"type"`  // "string", "number", "boolean"
	Value interface{} `json:"value"` // Actual value, type depends on "Type"
}
