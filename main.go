package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type quizResult struct {
	question      string
	userAnswer    string
	correctAnswer string
	isCorrect     bool
}

func generateQuizQuestion(difficulty string) (string, int) {
	signs := []string{"+", "-"}

	var num1 int
	var num2 int
	var sign string

	switch difficulty {
	case "Easy":
		num1 = r.Intn(10)
		num2 = r.Intn(10)
		sign = "+"
	case "Medium":
		num1 = r.Intn(90)
		num2 = r.Intn(20)
		sign = "+"
	case "Hard":
		num1 = r.Intn(90)
		num2 = r.Intn(30)
		sign = signs[r.Intn(len(signs))]
	}

	var correctRes int
	switch sign {
	case "+":
		correctRes = num1 + num2
	case "-":
		correctRes = num1 - num2
	default:
		log.Fatal("invalid sign: ", sign)
	}

	return strconv.Itoa(num1) + sign + strconv.Itoa(num2) + " = ", correctRes
}

func main() {
	timeoutAfter := flag.String("timeout", "30s", "set required timeout")
	questionsNumber := flag.Int("number", 10, "number of questions in a quiz")

	flag.Parse()

	prompt := promptui.Select{
		Label: "Select Difficulty",
		Items: []string{"Easy", "Medium", "Hard"},
	}

	_, difficulty, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}

	timeoutD, err := time.ParseDuration(*timeoutAfter)
	if err != nil {
		log.Fatal(err)
	}

	var counter atomic.Int64
	var m sync.RWMutex
	allCounter := *questionsNumber

	quizResultList := make([]quizResult, 0, allCounter)

	fmt.Println("Quiz will timeout after: " + timeoutD.String() + ". Press Enter...")

	inputReader := bufio.NewReader(os.Stdin)
	_, err = inputReader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-time.After(timeoutD)

		m.RLock()

		fmt.Println()
		for _, q := range quizResultList {
			printColoredResult(q.question, q.userAnswer, q.correctAnswer, q.isCorrect)
		}
		m.RUnlock()

		color.White("\nTimed out.")
		color.Yellow(resultWord(int(counter.Load()), allCounter))
		os.Exit(0)
	}()

	var isCorrect bool
	for range allCounter {
		isCorrect = false

		question, answer := generateQuizQuestion(difficulty)
		fmt.Print(question)

		userAnswer, err := inputReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		userAnswer = strings.TrimSpace(strings.ToLower(userAnswer))
		userAnswerI, err := strconv.Atoi(userAnswer)
		if err != nil {
			m.Lock()
			quizResultList = append(quizResultList, quizResult{
				question:      question,
				userAnswer:    userAnswer,
				correctAnswer: strconv.Itoa(answer),
				isCorrect:     false,
			})
			m.Unlock()
			continue
		}
		if userAnswerI == answer {
			isCorrect = true
			counter.Add(1)
		}
		m.Lock()
		quizResultList = append(quizResultList, quizResult{
			question:      question,
			userAnswer:    strconv.Itoa(userAnswerI),
			correctAnswer: strconv.Itoa(answer),
			isCorrect:     isCorrect,
		})
		m.Unlock()
	}

	for _, q := range quizResultList {
		printColoredResult(q.question, q.userAnswer, q.correctAnswer, q.isCorrect)
	}

	color.White("%d out of %d. \n", counter.Load(), allCounter)
	color.Yellow(resultWord(int(counter.Load()), allCounter))
	os.Exit(0)
}

func printColoredResult(question string, userAnswer, correctAnswer string, isCorrect bool) {
	switch isCorrect {
	case true:
		color.Green("%s%s ✓", question, userAnswer)
	case false:
		color.Red("%s%s (correct: %s)", question, userAnswer, correctAnswer)
	}
}

func resultWord(i, total int) string {
	if total == 0 {
		return "No questions answered."
	}
	p := float64(i) / float64(total)
	var res = "Great Job!"

	switch {
	case p <= float64(0.3):
		res = "Bad result... Try one more time!"
	case p <= float64(0.6):
		res = "Not Bad! But needs improvement... Try one more time!"
	case p < float64(1.0):
		res = "Very good. Keep going!"
	}
	return res
}
