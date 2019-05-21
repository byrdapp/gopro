package storage

// Profile for all types of profiles
type Profile struct {
	UserID         string `json:"userId,omitempty"`
	DisplayName    string `json:"displayName"`
	FirstName      string `json:"firstName,omitempty"`
	LastName       string `json:"lastName,omitempty"`
	Address        string `json:"address,omitempty"`
	Email          string `json:"email,omitempty"`
	IsMedia        bool   `json:"isMedia,omitempty"`
	IsProfessional bool   `json:"isProfessional,omitempty"`
	IsPress        bool   `json:"isPress,omitempty"`
	// CreateDate         *time.Time `json:"createDate"`
	SalesQuantity      int64  `json:"salesQuantity,omitempty"`
	SalesAmount        int64  `json:"salesAmount,omitempty"`
	WithdrawableAmount int64  `json:"withdrawableAmount,omitempty"`
	UserPicture        string `json:"userPicture,omitempty"`
}

// Media struct
type Media struct {
	ID                  string `sql:"id"`
	ProfileData         *Profile
	UserID              string `json:"userId"`
	Country             string `json:"country,omitempty"`
	City                string `json:"city,omitempty"`
	GoCredits           byte   `json:"goCredits,omitempty"`
	SubscriptionCredits byte   `json:"subscriptionCredits,omitempty"`
}

// Transaction struct
type Transaction struct {
	PaymentDate              uint64 `json:"paymentDate,omitempty"`
	StoryID                  string `json:"storyId,omitempty"`
	PaymentSellerDisplayName string `json:"paymentSellerDisplayName,omitempty"`
	PaymentSeller            string `json:"paymentSeller,omitempty"`
	PaymentBuyer             string `json:"paymentBuyer,omitempty"`
	PaymentBuyerDisplayName  string `json:"paymentBuyerDisplayName,omitempty"`
}

// Withdrawals ..
type Withdrawals struct {
	// CashoutDetails       string `json:"cashoutDetails,omitempty"`
	RequestAmount        int64  `json:"requestAmount,omitempty"`
	RequestCompleted     bool   `json:"requestCompleted,omitempty"`
	RequestCompletedDate int64  `json:"requestCompletedDate,omitempty"`
	RequestUserID        string `json:"requestUser,omitempty"`
	RequestDate          int64  `json:"requestDate,omitempty"`
}
