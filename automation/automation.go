package automation

type AutomationSet struct {
	Automations []Automation
}

type Automation struct {
	Id          uint          `yaml:"id" json:"id"`
	Alias       string        `yaml:"alias" json:"alias"`
	Description string        `yaml:"description" json:"description"`
	Trigger     []BaseTrigger `yaml:"trigger" json:"trigger"`
	Condition   []Condition   `yaml:"condition" json:"condition"`
	Actions     []Action      `yaml:"action" json:"action"`
	Enabled     bool          `yaml:"active" json:"active"`
}
