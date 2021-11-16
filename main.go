package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/patrickhener/xing-user-enum/xing"

	"golang.org/x/term"
)

const version string = "v0.0.4"

// args will setup connection details and target
// either from env or if not there prompt for it
func args() (string, string, string, string, error) {
	var username string
	var password string
	var company string
	var err error

	reader := bufio.NewReader(os.Stdin)

	if os.Getenv("XINGUSER") == "" {
		fmt.Print("Enter Username: ")
		username, err = reader.ReadString('\n')
		if err != nil {
			return "", "", "", "", err

		}
	} else {
		username = os.Getenv("XINGUSER")
	}

	if os.Getenv("XINGPASS") == "" {
		fmt.Print("Enter Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", "", "", "", err
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
			return "", "", "", "", err
		}
	} else {
		company = os.Getenv("XINGTARGET")
	}

	proxy := ""
	if os.Getenv("XINGPROXY") != "" {
		proxy = os.Getenv("XINGPROXY")
	}

	return strings.TrimSpace(username), strings.TrimSpace(password), strings.TrimSpace(company), proxy, nil
}

func main() {
	fmt.Printf("xing-user-enum %s\n\n", version)
	// fetch connection details
	user, pass, company, proxy, err := args()
	if err != nil {
		fmt.Printf("[-] Error setting the connection details: %+v", err)
		os.Exit(-1)
	}

	// init xing object
	x := xing.Xing{
		Username: user,
		Password: pass,
		Proxy:    proxy,
	}

	// Connect
	if err := x.Connect(); err != nil {
		fmt.Printf("[-] Error connecting to XING: %+v\n", err)
		os.Exit(-1)
	}

	slug, companyTitle, err := x.FindTargetSlug(company)
	if err != nil {
		fmt.Printf("[-] Error finding target company: %+v", err)
		os.Exit(-1)
	}
	fmt.Printf("[+] Target Company is: '%s'\n", companyTitle)

	fmt.Println("[*] Query company unique ID")

	id, err := x.GQLExtractCompID(slug)
	if err != nil {
		fmt.Printf("[-] Error fetching unique ID of company: %+v", err)
		os.Exit(-1)
	}

	if id == "" {
		fmt.Println("[-] Something went wrong fetching the unique ID - returned string is empty")
		os.Exit(-1)
	}

	total, err := x.GQLExtractEmployeesCount(id)
	if err != nil {
		fmt.Printf("[-] Error fetching the total employees count: %+v", err)
		os.Exit(-1)
	}

	fmt.Printf("[+] There are '%d' employees listed with '%s' on Xing\n", total, companyTitle)
	fmt.Println("[*] Extracting employees in badges of maximum 100")

	employees, err := x.GQLExtractEmployees(id)
	if err != nil {
		fmt.Printf("[-] Error fetching the employees list: %+v", err)
		os.Exit(-1)
	}

	fmt.Println("[*] Outputting in format patrick.hener")
	if err := xing.SaveFirstLast(employees); err != nil {
		fmt.Printf("[-] Error writing to file %+v", err)
		os.Exit(-1)
	}

	fmt.Println("[*] Outputting in format phener")
	if err := xing.SaveOneLetterFirstLast(employees); err != nil {
		fmt.Printf("[-] Error writing to file %+v", err)
		os.Exit(-1)
	}
}
