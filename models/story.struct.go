package models

type (
	// StoryProps is used to create reference to a story property for f.ex slack msg
	StoryProps struct {
		StoryID       string `json:"storyId"`
		StoryHeadline string `json:"storyHeadline"`
	}
	// Assignment .
	Assignment struct {
		Headline     string `json:"assignmentHeadline"`
		AssignmentID string `json:"assignmentId"`
	}
	// ProfileProps .
	ProfileProps struct {
		DisplayName string `json:"displayName"`
		UserID      string `json:"userId"`
	}
)
