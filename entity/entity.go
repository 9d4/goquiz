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
