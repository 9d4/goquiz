# GoQuiz
GoQuiz is a straightforward quiz application that has been ported from [BayuDC's quiz](https://github.com/bayudc/quiz) and completely rewritten using the Go programming language. It has been developed with a focus on simplicity, efficiency, and self-contained functionality. The front-end still the same.

## Run
The required seeders are `users.yml` and `quiz.yml` copy it from example.

```bash 
cp quiz.example.yml quiz.yml
cp users.example.yml users.yml
```

Build and run from Source:
```bash
go run . -s # -s to run seeder. Run once!

go run . # run without seeder
```

On the first run the app will automatically create the config `goquiz.yml`. Configure it if needed.

### Build with these in mind:
- Single Binary: GoQuiz is designed to be compiled into a single binary file, making it easy to distribute and run on various systems without any additional dependencies or installations.

- No Third-Party Database: GoQuiz eliminates the need for external databases. All quiz questions and answers are stored directly within the application itself, ensuring a seamless and self-contained experience.

- Run and Go: With GoQuiz, you can simply run the binary file and start using the application immediately. There's no need for complex setup or configuration processes.

**Credit to:**

- [BayuDC](https://github.com/bayudc)
- [Dimanda](https://github.com/9d4)
