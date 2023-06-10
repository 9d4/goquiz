package entity

type User struct {
	ID       int `storm:"id,increment"`
	Fullname string
	Username string `storm:"unique"`
	Password string
}

type Question struct {
	ID   int    `storm:"id,increment" json:"id"`
	Body string `json:"body"`
}

type Choice struct {
	ID         int `storm:"id,increment" json:"id"`
	QuestionID int
	Body       string `json:"body"`
	Correct    bool
}

type Answer struct {
	ID         int `storm:"id,increment"`
	UserID     int `storm:"index"`
	QuestionID int `storm:"index"`
	ChoiceID   int `storm:"index"`
	Correct    bool
}

type Score struct {
	ID     int `storm:"id,increment"`
	UserID int `storm:"index,unique"`
	Value  float64
}
