package models

type (
	// StoryProps is used to create reference to a story property for f.ex slack msg
	StoryProps struct {
		StoryID      string        `json:"storyId"`
		Assignment   Assignment    `json:"assignment"`
		Headline     string        `json:"headline"`
		ProfileProps *ProfileProps `json:"profileProps"`
	}
	// Assignment .
	Assignment struct {
		AssignmentLink string   `json:"assignmentLink"`
		AssignmentID   string   `json:"assignmentID"`
		Medias         []string `json:"medias"`
	}
	// ProfileProps .
	ProfileProps struct {
		DisplayName    string `json:"displayName"`
		ProfilePicture string `json:"profilePicture"`
	}
)
