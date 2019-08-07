package models

// StoryProps is used to create reference to a story property for f.ex slack msg
type StoryProps struct {
	StoryID       string `json:"storyId"`
	StoryHeadline string `json:"storyHeadline"`
}

// Assignment .
type Assignment struct {
	Headline     string `json:"assignmentHeadline"`
	AssignmentID string `json:"assignmentId"`
}

// ProfileProps .
type ProfileProps struct {
	DisplayName string `json:"displayName"`
	UserID      string `json:"userId"`
	Email       string `json:"email"`
	Country     string `json:"country"`
}
