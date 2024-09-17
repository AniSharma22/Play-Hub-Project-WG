package ui

import (
	"context"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"project2/internal/domain/entities"
	"project2/pkg/utils"
	"project2/pkg/validation"
	"strings"
	"syscall"
)

func (ui *UI) ShowSignupPage() {
	var email, password, phoneNo, gender string

	// Get valid email
	for {

		fmt.Print("Enter your WatchGuard email: ")
		email, _ = ui.reader.ReadString('\n')
		email = strings.TrimSpace(email)
		if validation.IsValidEmail(email) && !ui.userService.EmailAlreadyRegistered(context.Background(), email) {
			break
		} else if !validation.IsValidEmail(email) {
			fmt.Println("Invalid email. Please try again.")
		} else {
			fmt.Println("Email already exists. Please enter a new email")
		}
	}

	// Get and confirm password
	for {
		fmt.Println("(1 Capital, 1 small, 1 special character with min 8 length)")
		fmt.Print("Enter your password: ")
		bytePassword1, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("Error reading password:", err)
		}
		fmt.Println()

		// Check password complexity before asking for confirmation
		if !validation.IsValidPassword(string(bytePassword1)) {
			fmt.Println("Password complexity not met! Please try again.")
			continue
		}

		// Password is valid, ask for confirmation
		fmt.Print("Confirm your password: ")
		bytePassword2, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("Error reading password:", err)
		}
		fmt.Println()

		if string(bytePassword1) != string(bytePassword2) {
			fmt.Println("Passwords did not match. Please try again.")
		} else {
			password, _ = utils.GetHashedPassword(bytePassword1)
			break
		}
	}

	// Get valid phone number
	for {
		fmt.Print("Enter your phone number: ")
		phoneNo, _ = ui.reader.ReadString('\n')
		phoneNo = strings.TrimSpace(phoneNo)
		if validation.IsValidPhoneNumber(phoneNo) {
			break
		} else {
			fmt.Println("Invalid phone number. Please try again.")
		}
	}

	// Get valid gender
	for {
		fmt.Print("Enter your gender (Male/Female/Other): ")
		gender, _ = ui.reader.ReadString('\n')
		gender = strings.TrimSpace(gender)
		if validation.IsValidGender(gender) {
			break
		} else {
			fmt.Println("Invalid gender. Please try again.")
		}
	}

	// Create a user entity
	user := entities.User{
		Email:        email,
		Password:     password,
		MobileNumber: phoneNo,
		Gender:       gender,
	}

	// Sign up the user
	if err := ui.userService.Signup(context.Background(), &user); err != nil {
		fmt.Println(err)
		return
	} // Redirect to User dashboard
	fmt.Println("Signup successful!")
	ui.ShowUserDashboard()
}
