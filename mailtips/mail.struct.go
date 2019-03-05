package mailtips

// Content asd
type Content struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// Email asd
type Email struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"email,omitempty"`
}

// MailBody asd
type MailBody struct {
	To      []*Email   `json:"to,omitempty"`
	Subject string     `json:"subject,omitempty"`
	From    *Email     `json:"from,omitempty"`
	Content []*Content `json:"content,omitempty"`
}
