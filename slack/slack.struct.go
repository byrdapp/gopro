package slack

type (
	// Message .
	Message struct {
		Text        string        `json:"text"`
		Attachments []Attachments `json:"attachments"`
	}
	// Attachments -
	Attachments struct {
		// https://api.slack.com/docs/messages/builder
		Fallback   string    `json:"fallback"` //Required!
		Text       string    `json:"text"`     //Within attachment
		Pretext    string    `json:"pretext"`  //Outside attachment
		Color      string    `json:"color"`
		Authorname string    `json:"author_name"`
		Authoricon string    `json:"author_icon,omitempty"`
		Title      string    `json:"title"`
		Titlelink  string    `json:"title_link"`
		Fields     []*Fields `json:"fields"`
		Timestamp  int64     `json:"ts"`
	}
	// Fields .
	Fields struct {
		Title string `json:"title"`
		Value string `json:"value"`
		Short bool   `json:"short"`
	}
)
