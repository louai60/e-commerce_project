package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
)

type Service struct {
    Name    string
    Path    string
    Port    string
    Command string
}

func main() {
    // Get current working directory
    wd, err := os.Getwd()
    if err != nil {
        log.Fatal(err)
    }

    // Define services
    services := []Service{
        {
            Name:    "API Gateway",
            Path:    filepath.Join(wd, "api-gateway"),
            Port:    "8080",
            Command: "air",
        },
        {
            Name:    "Product Service",
            Path:    filepath.Join(wd, "product-service"),
            Port:    "50051",
            Command: "air",
        },
        {
            Name:    "User Service",
            Path:    filepath.Join(wd, "user-service"),
            Port:    "50052",
            Command: "air",
        },
        {
            Name:    "Frontend",
            Path:    filepath.Join(wd, "..", "frontend"),
            Port:    "3000",
            Command: "npm run dev",
        },
    }

    // Find Git Bash path
    gitBashPath := findGitBash()
    if gitBashPath == "" {
        log.Fatal("Git Bash not found")
    }

    // Launch each service
    for _, service := range services {
        launchServiceInGitBash(gitBashPath, service)
    }

    fmt.Println("\nAll services are starting:")
    for _, service := range services {
        if service.Name == "Frontend" || service.Port == "8080" {
            fmt.Printf("%s ➔ http://localhost:%s\n", service.Name, service.Port)
        } else {
            fmt.Printf("%s ➔ localhost:%s\n", service.Name, service.Port)
        }
    }
}

func findGitBash() string {
    commonPaths := []string{
        `C:\Program Files\Git\bin\bash.exe`,
        `C:\Program Files (x86)\Git\bin\bash.exe`,
    }

    for _, path := range commonPaths {
        if _, err := os.Stat(path); err == nil {
            return path
        }
    }

    return ""
}

func launchServiceInGitBash(gitBashPath string, service Service) {
    // Wrap the commands
    bashCommand := fmt.Sprintf(`cd "%s" && %s; exec bash`, service.Path, service.Command)

    // Build the full command to open Git Bash and run inside new window
    cmd := exec.Command("cmd", "/C", "start", "", gitBashPath, "-c", bashCommand)

    // Start the command
    err := cmd.Start()
    if err != nil {
        log.Printf("Failed to start %s: %v\n", service.Name, err)
    }
}
