package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func LoadEnv(filename string) {

	data, err := os.ReadFile(filename)

	if err != nil {
		log.Println("No .env file found")
		return
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)

		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			os.Setenv(key, value)
		}

	}

	fmt.Println("Successfully load env file", filename)

}
