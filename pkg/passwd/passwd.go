package passwd

import (
	"os"
	"strings"
)

func Shell(username, fallback string) string {
	passwd, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return fallback
	}

	for _, line := range strings.Split(string(passwd), "\n") {
		parts := strings.Split(line, ":")

		if len(parts) != 7 {
			continue
		}

		if parts[0] == username {
			return parts[6]
		}
	}

	return fallback
}
