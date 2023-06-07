package entity

type User struct {
	ID       int `storm:"id,increment"`
	Fullname string
	Username string `storm:"unique"`
	Password string
}

type Question struct {
	ID   int `storm:"id,increment"`
	Body string
}

type Choice struct {
	ID         int `storm:"id,increment"`
	QuestionID int
	Body       string
	Correct    bool
}
