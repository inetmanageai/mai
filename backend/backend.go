package backend

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/inetmanageai/mai/spinner"
)

var (
	projectNameDefault = "structure-golang"
	envDefault         = `ENV=dev

DB_URI=<mongodb://<username>:<password>@<host>>
DB_NAME=<DB_name>
`
	envEvent = `
ELASTIC_HOST=<https://<username>:<password>@<host>>
ELASTIC_INDEX=<elastic_index>
`
)

type Backend struct {
}

type OptionBackend struct {
}

func NewBackend(opt *OptionBackend) Backend {
	return Backend{}
}

func (b Backend) Create() error {
	// Input project name
	reader := bufio.NewReader(os.Stdin)
	projectName := ""
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter project name: ")
		projectName, _ = reader.ReadString('\n')
		projectName = strings.TrimSpace(projectName)
		if regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`).MatchString(projectName) {
			break
		} else {
			fmt.Println("Invalid format project_name")
		}
	}

	// Input feature events
	fmt.Print("Add event (Kafka) [y/n]: ")
	event, _ := reader.ReadString('\n')
	event = strings.TrimSpace(event)

	// Input gitlab-ci file
	fmt.Print("Add gitlab-ci [y/n]: ")
	gitci, _ := reader.ReadString('\n')
	gitci = strings.TrimSpace(gitci)

	spin := make(chan bool)
	go spinner.Spinner(spin)

	// Clone the repository
	cmd := exec.Command("git", "clone", "https://github.com/inetmanageai/structure-golang.git")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error cloning repository: %v", err)
	}

	// Check if user choose to add events
	if strings.ToLower(event) != "y" {
		os.RemoveAll("./structure-golang/common/events")
		os.Remove("./structure-golang/core/handlers/consumer_event.go")
		err := deleteLineWithMatch("./structure-golang/main.go", `[e,E]vent|select\s\{\}|[c,C]onsumer`)
		if err != nil {
			return err
		}
		err = replaceInFile("./structure-golang/main.go", "go ", "")
		if err != nil {
			return err
		}
	}

	// Check if user choose to add a gitlab-ci file
	if strings.ToLower(gitci) == "y" {
		spin <- false
		fmt.Print("\rEnter runner (default: uranus-01): ")
		runner, _ := reader.ReadString('\n')
		go spinner.Spinner(spin)
		runner = strings.TrimSpace(runner)
		if runner != "" {
			err := replaceInFile("./"+projectNameDefault+"/.gitlab-ci.yml", "uranus-01", runner)
			if err != nil {
				return err
			}
		}
	} else {
		os.Remove("./" + projectNameDefault + "/.gitlab-ci.yml")
	}

	// Remove .git file, go.mod, go.sum
	os.RemoveAll("./" + projectNameDefault + "/.git")
	os.Remove("./" + projectNameDefault + "/go.mod")
	os.Remove("./" + projectNameDefault + "/go.sum")
	os.RemoveAll("./" + projectNameDefault + "/.github")
	os.Remove("./" + projectNameDefault + "/LICENSE")

	// Replace project name
	err := replaceInFiles("./"+projectNameDefault, projectNameDefault, projectName)
	if err != nil {
		return err
	}

	// Rename the directory
	maxRetries := 3
	time.Sleep(1 * time.Second)
	for i := 0; i < maxRetries; i++ {
		err := os.Rename("./"+projectNameDefault, "./"+projectName)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Init module project
	os.Chdir("./" + projectName)
	exec.Command("go", "mod", "init", projectName).Run()
	exec.Command("go", "mod", "tidy").Run()

	// Reset README file
	err = os.WriteFile("README.md", []byte("# "+projectName), 0644)
	if err != nil {
		return fmt.Errorf("error write to README.md: %v", err)
	}

	// Create default ENV file
	env := envDefault
	if event == "y" {
		env += envEvent
	}
	err = os.WriteFile(".env", []byte(env), 0644)
	if err != nil {
		return fmt.Errorf("error write to .env: %v", err)
	}

	close(spin)
	fmt.Println("\rProject creation finished.")
	return nil
}

func replaceInFile(filePath, old, new string) error {
	input, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v %v", filePath, err)
	}

	output := strings.ReplaceAll(string(input), old, new)
	if err = os.WriteFile(filePath, []byte(output), 0644); err != nil {
		return fmt.Errorf("error writing file: %v %v", filePath, err)
	}

	return nil
}

func replaceInFiles(rootPath, old, new string) error {
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			err = replaceInFile(path, old, new)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func deleteLineWithMatch(pathFile, reg string) error {
	// Open the Go source file
	file, err := os.Open(pathFile)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)

	// Read the file line by line and keep all lines except the one to delete
	for scanner.Scan() {
		if !regexp.MustCompile(reg).MatchString(scanner.Text()) {
			lines = append(lines, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Write the modified content back to the file
	output := strings.Join(removeAdjacentEmpty(lines), "\n")
	err = os.WriteFile(pathFile, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func removeAdjacentEmpty(data []string) (result []string) {
	seeEmpty := false
	for _, v := range data {
		if v == "" {
			if seeEmpty {
				continue
			}
			result = append(result, v)
			seeEmpty = true
		} else {
			result = append(result, v)
			seeEmpty = false
		}
	}

	return
}
