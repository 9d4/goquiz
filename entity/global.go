package entity

var (
	globalBucketName = "global"
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
