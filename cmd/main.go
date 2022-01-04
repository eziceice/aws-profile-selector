package main

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"gopkg.in/ini.v1"
	"os"
	"os/exec"
)

type Credentials struct {
	AccessKeyId          string
	AccessKey            string
	SessionToken         string
	SecurityToken        string
	PrincipalArn         string
	SecurityTokenExpires string
}

const (
	AccessKeyId          = "aws_access_key_id"
	AccessKey            = "aws_secret_access_key"
	SessionToken         = "aws_session_token"
	SecurityToken        = "aws_security_token"
	PrincipalArn         = "x_principal_arn"
	SecurityTokenExpires = "x_security_token_expires"
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
		Label:     "Select the AWS account you want to login",
		Items:     profiles,
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

	_, err = overwriteDefaultCredentials(targetProfile)
	if err != nil {
		fmt.Printf("Cannot overwrite default AWS credentials %v\n", err)
		return
	}
}

func overwriteDefaultCredentials(targetProfile string) (*Credentials, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cfg, err := ini.Load(fmt.Sprintf("%v/.aws/credentials", dir))
	if err != nil {
		return nil, err
	}

	targetSection := cfg.Section(targetProfile)
	credentials := &Credentials{
		AccessKeyId:          targetSection.Key(AccessKeyId).String(),
		AccessKey:            targetSection.Key(AccessKey).String(),
		SessionToken:         targetSection.Key(SessionToken).String(),
		SecurityToken:        targetSection.Key(SecurityToken).String(),
		PrincipalArn:         targetSection.Key(PrincipalArn).String(),
		SecurityTokenExpires: targetSection.Key(SecurityTokenExpires).String(),
	}

	defaultSection := cfg.Section("default")
	defaultSection.Key(AccessKeyId).SetValue(credentials.AccessKeyId)
	defaultSection.Key(AccessKey).SetValue(credentials.AccessKey)
	defaultSection.Key(SessionToken).SetValue(credentials.SessionToken)
	defaultSection.Key(SecurityToken).SetValue(credentials.SecurityToken)
	defaultSection.Key(PrincipalArn).SetValue(credentials.PrincipalArn)
	defaultSection.Key(SecurityTokenExpires).SetValue(credentials.SecurityTokenExpires)

	err = cfg.SaveTo(fmt.Sprintf("%v/.aws/credentials", dir))
	if err != nil {
		return nil, err
	}

	return credentials, nil
}

func readProfiles() ([]string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cfg, err := ini.Load(fmt.Sprintf("%v/.saml2aws", dir))
	if err != nil {
		return nil, err
	}
	names := cfg.SectionStrings()

	return names[1:], nil
}
