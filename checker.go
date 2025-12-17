package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func checkOS(server Server, results chan<- Result) {
	config := &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(server.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	address := fmt.Sprintf("%s:%s", server.Host, server.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		results <- Result{
			Address: address,
			Status:  "ERROR",
			Error:   err.Error(),
		}
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		results <- Result{
			Address: address,
			Status:  "ERROR",
			Error:   "не удалось создать сессию: " + err.Error(),
		}
		return
	}
	defer session.Close()

	output, err := session.CombinedOutput("uname -a")

	// Если uname не работает, пробуем FortiGate команды
	if err != nil {
		session2, _ := client.NewSession()
		if session2 != nil {
			defer session2.Close()
			output2, err2 := session2.CombinedOutput("get system status")
			if err2 == nil && len(output2) > 0 {
				osInfo := cleanFortiOutput(string(output2))
				version := extractFortiVersion(osInfo)
				results <- Result{
					Address: address,
					Status:  "OK",
					OS:      "FortiOS",
					Info:    version,
				}
				return
			}
		}

		results <- Result{
			Address: address,
			Status:  "ERROR",
			Error:   "не удалось выполнить команду: " + err.Error(),
		}
		return
	}

	osInfo := strings.TrimSpace(string(output))

	// Проверяем, не FortiGate ли это (иногда они отвечают на uname)
	if strings.Contains(osInfo, "Unknown action") || strings.Contains(osInfo, "FortiGate") {
		session2, _ := client.NewSession()
		if session2 != nil {
			defer session2.Close()
			output2, err2 := session2.CombinedOutput("get system status")
			if err2 == nil && len(output2) > 0 {
				osInfo = cleanFortiOutput(string(output2))
				version := extractFortiVersion(osInfo)
				results <- Result{
					Address: address,
					Status:  "OK",
					OS:      "FortiOS",
					Info:    version,
				}
				return
			}
		}
		// Если не получилось, просто говорим что это FortiOS
		results <- Result{
			Address: address,
			Status:  "OK",
			OS:      "FortiOS",
			Info:    "FortiGate устройство",
		}
		return
	}

	osType := detectOS(osInfo)

	results <- Result{
		Address: address,
		Status:  "OK",
		OS:      osType,
		Info:    truncateString(osInfo, 60),
	}
}

func detectOS(info string) string {
	lower := strings.ToLower(info)

	if strings.Contains(lower, "linux") {
		if strings.Contains(lower, "ubuntu") {
			return "Ubuntu"
		} else if strings.Contains(lower, "debian") {
			return "Debian"
		} else if strings.Contains(lower, "centos") {
			return "CentOS"
		} else if strings.Contains(lower, "red hat") || strings.Contains(lower, "rhel") {
			return "Red Hat"
		} else if strings.Contains(lower, "fedora") {
			return "Fedora"
		} else if strings.Contains(lower, "alpine") {
			return "Alpine"
		}
		return "Linux"
	} else if strings.Contains(lower, "darwin") {
		return "macOS"
	} else if strings.Contains(lower, "freebsd") {
		return "FreeBSD"
	} else if strings.Contains(lower, "openbsd") {
		return "OpenBSD"
	} else if strings.Contains(lower, "fortigate") {
		return "FortiGate"
	}

	return "Unknown"
}

func cleanFortiOutput(output string) string {
	// Убираем мусор типа "Unknown action 0", hostname и прочее
	lines := strings.Split(output, "\n")
	var cleaned []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Пропускаем мусорные строки
		if line == "" ||
			strings.Contains(line, "Unknown action") ||
			strings.Contains(line, "$") ||
			strings.HasSuffix(line, "#") {
			continue
		}
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

func extractFortiVersion(info string) string {
	lines := strings.Split(info, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Ищем строку с версией
		if strings.Contains(line, "Version:") || strings.Contains(line, "FortiOS") {
			// Извлекаем только версию
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				version := strings.TrimSpace(parts[1])
				// Убираем лишнее
				version = strings.Split(version, ",")[0]
				return "FortiOS " + version
			}
		}
		// Или ищем Platform
		if strings.Contains(line, "Platform:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return "FortiGate устройство"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
