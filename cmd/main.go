package main

import (
	"bufio"
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"os/exec"
	"strings"
)

func main() {
	profiles, err := readProfiles()
	if err != nil {
		fmt.Printf("Cannot load any AWS profiles %v\n", err)
		return
	}

	templates := promptui.SelectTemplates{
		Active:   `🍕 {{ . | red | bold }}`,
		Inactive: `   {{ . | cyan }}`,
		Selected: `{{ "✔" | green | bold }}: {{ . | red }}`,
	}

	prompt := promptui.Select{
		Label: "Select AWS Profile",
		Items: profiles,
		Templates: &templates,
	}
	_, targetProfile, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	fmt.Printf("You choose %q\n", targetProfile)

	app := "saml2aws"
	arg0 := "login"
	arg1 := "-a"
	arg2 := targetProfile
	arg3 := "--force"
	cmd := exec.Command(app, arg0, arg1, arg2, arg3)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	err = overwriteDefaultCredentials(targetProfile)
	if err != nil {
		fmt.Printf("Cannot overwrite default AWS credentials %v\n", err)
		return
	}

	arg0 = "script"
	cmd = exec.Command(app, arg0, arg1, arg2)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func overwriteDefaultCredentials(targetProfile string) error {
	results := make([]string, 0)

	dir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	targetFile := fmt.Sprintf("%v/.aws/credentials", dir)
	f, err := os.Open(targetFile)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	newDefault := make([]string, 0)
	newDefault = append(newDefault, "[default]")

	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "default") {
			for scanner.Scan() {
				text = scanner.Text()
				if strings.Contains(text, "[") || strings.Contains(text, "]") {
					break
				}
			}
		}

		if strings.HasPrefix(text, "[") && strings.Contains(text, targetProfile) {
			results = append(results, text)
			for scanner.Scan() {
				text = scanner.Text()
				if strings.Contains(text, "[") || strings.Contains(text, "]") {
					break
				}
				newDefault = append(newDefault, text)
				results = append(results, text)
			}
		}

		results = append(results, text)
	}

	results = append(newDefault, results...)

	newFile, err := os.OpenFile(targetFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)

	defer newFile.Close()

	for i := range results {
		_, err := newFile.WriteString(results[i] + "\n")
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func readProfiles() ([]string, error) {
	profiles := make([]string, 0)

	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	targetFile := fmt.Sprintf("%v/.saml2aws", dir)
	f, err := os.Open(targetFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
			profiles = append(profiles, strings.ReplaceAll(strings.ReplaceAll(text, "[", ""), "]", ""))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}
