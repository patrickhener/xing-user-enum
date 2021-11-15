package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"xing-user-enum/xing"

	"golang.org/x/term"
)

// args will setup connection details and target
// either from env or if not there prompt for it
func args() (string, string, string, error) {
	var username string
	var password string
	var company string
	var err error

	reader := bufio.NewReader(os.Stdin)

	if os.Getenv("XINGUSER") == "" {
		fmt.Print("Enter Username: ")
		username, err = reader.ReadString('\n')
		if err != nil {
			return "", "", "", err

		}
	} else {
		username = os.Getenv("XINGUSER")
	}

	if os.Getenv("XINGPASS") == "" {
		fmt.Print("Enter Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", "", "", err
		}

		password = string(bytePassword)
		fmt.Printf("\n")
	} else {
		password = os.Getenv("XINGPASS")
	}

	if os.Getenv("XINGTARGET") == "" {
		fmt.Print("Enter Target: ")
		company, err = reader.ReadString('\n')
		if err != nil {
			return "", "", "", err
		}
	} else {
		company = os.Getenv("XINGTARGET")
	}

	return strings.TrimSpace(username), strings.TrimSpace(password), strings.TrimSpace(company), nil
}

func main() {
	// fetch connection details
	user, pass, company, err := args()
	if err != nil {
		log.Panicf("Error setting the connection details: %+v", err)
	}

	// init xing object
	x := xing.Xing{
		Username: user,
		Password: pass,
	}

	// Connect
	if err := x.Connect(); err != nil {
		log.Panicf("Error connecting to XING: %+v\n", err)
	}

	// Identify target by provided company string
	comp, err := x.FindTarget(company)
	if err != nil {
		log.Panicf("Error finding target company: %+v", err)
	}

	var target xing.QueryCompany

	// If there is more than one company throw error
	if len(comp.Collection) > 1 {
		log.Fatal("There were more than one hits on company")

		for _, c := range comp.Collection {
			fmt.Printf("[*] %d - %s (%s)\n", c.ID, c.Title, c.Image)
		}

		log.Panic("Please refine search by company title")
	} else {
		target = comp.Collection[0]
	}

	fmt.Printf("[+] Target Company is: %d - %s (%s)\n", target.ID, target.Title, target.Image)

	fmt.Println("[*] Query company unique ID")

	id, err := x.GQLExtractCompID(company)
	if err != nil {
		log.Panicf("Error fetching unique ID of company: %+v", err)
	}

	if id == "" {
		log.Panic("Something went wrong fetching the unique ID - returned string is empty")
	}

	data, err := x.GQLExtractEmployees(id)
	if err != nil {
		log.Panicf("Error fetching the employees list: %+v", err)
	}

	fmt.Printf("[+] There are '%d' employees listed with '%s' on Xing\n", data.Company.Employees.Total, target.Title)

	fmt.Println("[*] Outputting in format patrick.hener")
	if err := xing.SaveFirstLast(data); err != nil {
		log.Fatalf("Error writing to file %+v", err)
	}

	fmt.Println("[*] Outputting in format phener")
	if err := xing.SaveOneLetterFirstLast(data); err != nil {
		log.Fatalf("Error writing to file %+v", err)
	}
}
