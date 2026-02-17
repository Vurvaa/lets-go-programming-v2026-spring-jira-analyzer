package main

import (
	"fmt"
	"time"
)

func main() {
	// Временный костыль, чтобы main работал и контейнер поддерживался живым
	fmt.Println("Hello Jira")
	time.Sleep(3 * time.Minute)
}
