package entity

import "github.com/asdine/storm"

var (
	globalBucketName = "global"
	quizBucketName   = "quiz"
	quizNameKey      = "quizName"
)

func SaveQuizName(name string) {
	DB().Set(globalBucketName, quizNameKey, name)
}

func GetQuizName() (name string) {
	DB().Get(globalBucketName, quizNameKey, &name)

	return
}

func CountQuestions() (count int) {
	count, _ = DB().Count(&Question{})
	return
}

func CountUsers() (count int) {
	count, _ = DB().Count(&User{})
	return
}

var ErrNotFound = storm.ErrNotFound

func QuizSet(key, data interface{}) error {
	return DB().Set(quizBucketName, key, data)
}

func QuizGet(key, to interface{}) (err error) {
	err = DB().Get(quizBucketName, key, to)
	return
}

func QuizDelete(key interface{}) error {
	return DB().Delete(quizBucketName, key)
}
