package automation

type Condition struct {
	Entity string `yaml:"entity" json:"entity"` // Integration.Typ.Entity_name
	Field  string `yaml:"field" json:"field"`   // "brightness"

	// Comparison operators
	Equals      interface{} `yaml:"equals,omitempty" json:"equals,omitempty"`
	NotEquals   interface{} `yaml:"not_equals,omitempty" json:"not_equals,omitempty"`
	GreaterThan *float64    `yaml:"gt,omitempty" json:"gt,omitempty"`
	LessThan    *float64    `yaml:"lt,omitempty" json:"lt,omitempty"`
}
