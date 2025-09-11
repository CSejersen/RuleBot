package client

type UpdateCommand[T any] struct {
	Update T `json:"update"`
}

type Page struct {
	Title   string    `yaml:"title" json:"title"`
	ID      string    `yaml:"id" json:"id"`
	Buttons []*Button `yaml:"buttons" json:"buttons"`
}

type Button struct {
	ID       string  `yaml:"id" json:"id"`
	Title    string  `yaml:"title" json:"title"`
	Subtitle string  `yaml:"subtitle" json:"subtitle"`
	Value    int     `yaml:"value" json:"value"`
	State    string  `yaml:"state" json:"state"`
	Content  Content `yaml:"content" json:"content"`
	Default  bool    `yaml:"default" json:"default"`
}

type Content struct {
	Text string `yaml:"text,omitempty" json:"text,omitempty"`
	Icon string `yaml:"icon,omitempty" json:"icon,omitempty"`
}
