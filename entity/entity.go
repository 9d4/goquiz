package entity

type User struct {
	ID       int    `storm:"id,increment" json:"id"`
	Fullname string `json:"fullname"`
	Username string `storm:"unique" json:"username"`
	Password string `json:"password"`
}

type Question struct {
	ID     int    `storm:"id,increment" json:"id"`
	Body   string `json:"body"`
	Number int    `json:"number"`
}

type Choice struct {
	ID         int    `storm:"id,increment" json:"id"`
	QuestionID int    `json:"questionId"`
	Body       string `json:"body"`
	Correct    bool   `json:"correct"`
}

type Answer struct {
	ID         int  `storm:"id,increment" json:"id"`
	UserID     int  `storm:"index" json:"userId"`
	QuestionID int  `storm:"index" json:"questionId"`
	ChoiceID   int  `storm:"index" json:"choiceId"`
	Correct    bool `json:"correct"`
}

type Score struct {
	ID     int     `storm:"id,increment" json:"id"`
	UserID int     `storm:"index,unique" json:"userId"`
	Value  float64 `json:"value"`
}
