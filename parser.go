package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func parseServers(filePath string) ([]Server, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var servers []Server
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Формат: IP:PORT@USER:PASSWORD
		parts := strings.Split(line, "@")
		if len(parts) != 2 {
			fmt.Printf("Неверный формат строки: %s\n", line)
			continue
		}

		hostPort := strings.Split(parts[0], ":")
		if len(hostPort) != 2 {
			fmt.Printf("Неверный формат хоста: %s\n", parts[0])
			continue
		}

		userPass := strings.Split(parts[1], ":")
		if len(userPass) != 2 {
			fmt.Printf("Неверный формат учетных данных: %s\n", parts[1])
			continue
		}

		servers = append(servers, Server{
			Host:     hostPort[0],
			Port:     hostPort[1],
			User:     userPass[0],
			Password: userPass[1],
		})
	}

	return servers, scanner.Err()
}
