package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Host     string
	Port     string
	User     string
	Password string
}

type Result struct {
	Address string
	Status  string
	OS      string
	Info    string
	Error   string
}

func main() {
	printBanner()

	fmt.Print("📁 Введите путь к файлу с серверами: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	filePath := strings.TrimSpace(scanner.Text())

	servers, err := parseServers(filePath)
	if err != nil {
		fmt.Printf("❌ Ошибка чтения файла: %v\n", err)
		waitAndExit()
		return
	}

	fmt.Printf("\n🔍 Найдено серверов: %d\n", len(servers))
	fmt.Println("⏳ Начинаю проверку...\n")
	fmt.Println(strings.Repeat("─", 80))

	var wg sync.WaitGroup
	results := make(chan Result, len(servers))
	var allResults []Result

	startTime := time.Now()

	for _, server := range servers {
		wg.Add(1)
		go func(s Server) {
			defer wg.Done()
			checkOS(s, results)
		}(server)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	successCount := 0
	errorCount := 0

	for result := range results {
		allResults = append(allResults, result)
		if result.Status == "OK" {
			successCount++
			fmt.Printf("✅ %-21s │ %-15s │ %s\n", result.Address, result.OS, result.Info)
		} else {
			errorCount++
			fmt.Printf("❌ %-21s │ %s\n", result.Address, result.Error)
		}
	}

	elapsed := time.Since(startTime)

	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("\n📊 Статистика:\n")
	fmt.Printf("   ✅ Успешно: %d\n", successCount)
	fmt.Printf("   ❌ Ошибок: %d\n", errorCount)
	fmt.Printf("   ⏱️  Время: %.2f сек\n\n", elapsed.Seconds())

	// Сохранение в файл
	saveResults(allResults)

	waitAndExit()
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                    OS CHECKER v1.0                        ║
║              Проверка ОС на удаленных серверах            ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func saveResults(results []Result) {
	file, err := os.Create("os.txt")
	if err != nil {
		fmt.Printf("⚠️  Не удалось создать файл os.txt: %v\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString("OS CHECKER - Результаты проверки\n")
	writer.WriteString("Дата: " + time.Now().Format("2006-01-02 15:04:05") + "\n")
	writer.WriteString(strings.Repeat("=", 80) + "\n\n")

	for _, result := range results {
		if result.Status == "OK" {
			writer.WriteString(fmt.Sprintf("[OK] %s\n", result.Address))
			writer.WriteString(fmt.Sprintf("  ОС: %s\n", result.OS))
			writer.WriteString(fmt.Sprintf("  Инфо: %s\n\n", result.Info))
		} else {
			writer.WriteString(fmt.Sprintf("[ERROR] %s\n", result.Address))
			writer.WriteString(fmt.Sprintf("  Ошибка: %s\n\n", result.Error))
		}
	}

	writer.Flush()
	fmt.Printf("💾 Результаты сохранены в файл: os.txt\n\n")
}

func waitAndExit() {
	fmt.Println("⏳ Программа закроется через 2 минуты...")
	fmt.Println("   (Нажмите Enter для немедленного выхода)")

	done := make(chan bool)

	go func() {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("👋 Выход...")
	}
}
