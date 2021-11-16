package xing

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/machinebox/graphql"
)

func (x *Xing) Connect() error {
	ld := LoginData{
		Username: x.Username,
		Password: x.Password,
		Perm:     "1",
	}

	// Init jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("not able to create cookie jar: %+v", err)
	}

	// Init session (http client using the jar)
	x.Session = &http.Client{
		Jar: jar,
	}

	// Set proxy if not empty
	if x.Proxy != "" {
		proxyURL, err := url.Parse(x.Proxy)
		if err != nil {
			return fmt.Errorf("not able to craft proxy url: %+v", err)
		}
		x.Session.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	// Fetch initial cookies by getting the homepage and login page
	fmt.Println("[*] Fetching cookies in www.xing.com")
	req, err := http.NewRequest("GET", "https://www.xing.com", nil)
	if err != nil {
		return fmt.Errorf("not able to craft request to www.xing.com: %+v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:94.0) Gecko/20100101 Firefox/94.0")

	resp, err := x.Session.Do(req)
	if err != nil {
		return fmt.Errorf("not able to request www.xing.com: %+v", err)
	}
	defer resp.Body.Close()

	fmt.Println("[*] Fetching cookies in login.xing.com")
	req, err = http.NewRequest("GET", "https://login.xing.com", nil)
	if err != nil {
		return fmt.Errorf("not able to craft request to login.xing.com: %+v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:94.0) Gecko/20100101 Firefox/94.0")

	resp, err = x.Session.Do(req)
	if err != nil {
		return fmt.Errorf("not able to request login.xing.com: %+v", err)
	}
	defer resp.Body.Close()

	fmt.Println("[*] Fetching XSRF Token")
	// Getting xsrf token from response
	xsrf := ""
	for _, c := range resp.Cookies() {
		if c.Name == "xing_csrf_token" {
			xsrf = c.Value
		}
	}

	// Only continue if xsrf is present
	if xsrf != "" {
		fmt.Println("[+] XSRF Token found")
		// Actual login
		js, err := json.Marshal(&ld)
		if err != nil {
			return fmt.Errorf("unable to marshal login data: %+v", err)
		}

		req, err = http.NewRequest("POST", "https://login.xing.com/login/api/login", bytes.NewBuffer(js))
		if err != nil {
			return fmt.Errorf("not able to craft login request to login.xing.com: %+v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Csrf-Token", xsrf)
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:94.0) Gecko/20100101 Firefox/94.0")

		fmt.Println("[*] Trigger login request")
		resp, err = x.Session.Do(req)
		if err != nil {
			return fmt.Errorf("login request did not complete: %+v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			fmt.Println("[+] Login was successful")
		} else {
			return fmt.Errorf("Login was not successful: Status - %d", resp.StatusCode)
		}

		return nil
	} else {
		return fmt.Errorf("%s", "XSRF token was not retrieved from server response")
	}
}

// func (x *Xing) FindTarget(company string) (Collection, error) {
// 	var suggestions Collection

// 	base := "https://www.xing.com/dsc/suggestions/companies/name.json?consumer=loggedin.web.search.companies.name&query="

// 	url := fmt.Sprintf("%s%s", base, url.PathEscape(company))

// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return suggestions, fmt.Errorf("unable to craft request: %+v", err)
// 	}

// 	resp, err := x.Session.Do(req)
// 	if err != nil {
// 		return suggestions, fmt.Errorf("unable to request company: %+v", err)
// 	}
// 	defer resp.Body.Close()

// 	data, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return suggestions, fmt.Errorf("unable to read server response: %+v", err)
// 	}

// 	if err := json.Unmarshal(data, &suggestions); err != nil {
// 		return suggestions, fmt.Errorf("unable to unmarshal received data into struct: %+v", err)
// 	}

// 	return suggestions, nil
// }

func (x *Xing) FindTargetSlug(company string) (string, string, error) {
	var data RespCompany
	slug := ""
	companyTitle := ""

	base := "https://www.xing.com/search/api/results/companies?sc_o=navigation_search_advanced_click&keywords="
	url := fmt.Sprintf("%s%s", base, url.PathEscape(company))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("unable to craft request: %+v", err)
	}

	resp, err := x.Session.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("unable to request company: %+v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("unable to read server response: %+v", err)
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return "", "", fmt.Errorf("unable to unmarshal received data into struct: %+v", err)
	}

	iCount, err := strconv.Atoi(data.Count)
	if err != nil {
		return "", "", fmt.Errorf("unable to convert count to int: %+v", err)
	}

	if iCount > 1 {
		// Ask which to use interactively
		for i, e := range data.Items {
			slugSplit := strings.Split(e.Link, "/")
			data.Items[i].Slug = slugSplit[len(slugSplit)-1]
		}
		fmt.Println("[*] There were more than one hits on company")
		fmt.Println("[*] Those are:")
		for i, e := range data.Items {
			fmt.Printf("[%d] %s (%s)\n", i, e.Title, e.Link)
		}
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("[!] Please choose company to use: ")
		id, err := reader.ReadString('\n')
		if err != nil {
			return "", "", fmt.Errorf("error when choosing company: %+v", err)

		}
		iId, err := strconv.Atoi(strings.Split(id, "\n")[0])
		if err != nil {
			return "", "", fmt.Errorf("unable to convert id to int: %+v", err)
		}
		slug = data.Items[iId].Slug
		companyTitle = data.Items[iId].Title

	} else if iCount < 1 {
		return "", "", fmt.Errorf("%s", "no company with this name was found")
	} else {
		slugSplit := strings.Split(data.Items[0].Link, "/")
		slug = slugSplit[len(slugSplit)-1]
		companyTitle = data.Items[0].Title
	}

	return slug, companyTitle, nil
}

func (x *Xing) GQLExtractCompID(compname string) (string, error) {
	var id string
	var resp CompData
	gqlClient := graphql.NewClient("https://www.xing.com/xing-one/api")
	req := graphql.NewRequest(compQuery)

	req.Var("id", compname)

	if err := gqlClient.Run(context.Background(), req, &resp); err != nil {
		return "", fmt.Errorf("unable to get company id: %+v", err)
	}

	id = resp.EntityPageEx.Context.CompanyID

	return id, nil
}

func (x *Xing) GQLExtractEmployeesCount(id string) (int, error) {
	var data Data
	limit := 1
	gqlClient := graphql.NewClient("https://www.xing.com/xing-one/api", graphql.WithHTTPClient(x.Session))
	req := graphql.NewRequest(empQuery)

	req.Var("consumer", "")
	req.Var("includeTotalQuery", false)
	req.Var("id", id)
	req.Var("first", limit)

	q := Query{
		Consumer: "web.entity_pages.employees_subpage",
		Sort:     "LAST_NAME",
	}
	req.Var("query", q)

	if err := gqlClient.Run(context.Background(), req, &data); err != nil {
		return 0, fmt.Errorf("unable to get employ count: %+v", err)
	}

	return data.Company.Employees.Total, nil
}

func (x *Xing) GQLExtractEmployees(id string) ([]Edge, error) {
	var data Data
	var employees []Edge
	limit := 100
	gqlClient := graphql.NewClient("https://www.xing.com/xing-one/api", graphql.WithHTTPClient(x.Session))
	req := graphql.NewRequest(empQuery)

	req.Var("consumer", "")
	req.Var("includeTotalQuery", false)
	req.Var("id", id)
	req.Var("first", limit)

	q := Query{
		Consumer: "web.entity_pages.employees_subpage",
		Sort:     "LAST_NAME",
	}
	req.Var("query", q)

	// Initial query
	fmt.Println("[*] Initial badge")
	if err := gqlClient.Run(context.Background(), req, &data); err != nil {
		return nil, fmt.Errorf("unable to query first badge: %+v", err)
	}
	employees = append(employees, data.Company.Employees.Edges...)

	// Extract the rest via badges
	for data.Company.Employees.PageInfo.HasNextPage {
		fmt.Println("[*] Running another badge")
		req.Var("after", data.Company.Employees.PageInfo.EndCursor)
		if err := gqlClient.Run(context.Background(), req, &data); err != nil {
			return nil, fmt.Errorf("unable to query a following badge: %+v", err)
		}
		employees = append(employees, data.Company.Employees.Edges...)
	}

	return employees, nil
}

func SaveFirstLast(employees []Edge) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable getting current working path: %+v", err)
	}

	file, err := os.OpenFile(path.Join(cwd, "first.last.users.txt"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open file: %+v", err)
	}

	datawriter := bufio.NewWriter(file)

	var sorted []string
	for _, e := range employees {
		if e.Node.ProfileDetails.ID != "" {
			format := fmt.Sprintf("%s.%s\n", strings.ToLower(e.Node.ProfileDetails.FirstName), strings.ToLower(e.Node.ProfileDetails.LastName))
			sorted = append(sorted, format)
		}
	}

	// actually sort the slice
	sort.Strings(sorted)

	for _, e := range sorted {
		_, err := datawriter.WriteString(e)
		if err != nil {
			return fmt.Errorf("unable to write line in datawriter: %+v", err)
		}
	}

	if err := datawriter.Flush(); err != nil {
		return fmt.Errorf("unable to flush datawriter: %+v", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("unable to close file properly: %+v", err)
	}

	fmt.Printf("[+] %s has been written\n", path.Join(cwd, "first.last.users.txt"))

	return nil
}

func SaveOneLetterFirstLast(employees []Edge) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable getting current working path: %+v", err)
	}

	file, err := os.OpenFile(path.Join(cwd, "flast.users.txt"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open file: %+v", err)
	}

	datawriter := bufio.NewWriter(file)

	var sorted []string
	for _, e := range employees {
		if e.Node.ProfileDetails.ID != "" {
			format := fmt.Sprintf("%s%s\n", strings.ToLower(string(e.Node.ProfileDetails.FirstName[0])), strings.ToLower(e.Node.ProfileDetails.LastName))
			sorted = append(sorted, format)
		}
	}

	// actually sort the slice
	sort.Strings(sorted)

	for _, e := range sorted {
		_, err := datawriter.WriteString(e)
		if err != nil {
			return fmt.Errorf("unable to write line in datawriter: %+v", err)
		}
	}

	if err := datawriter.Flush(); err != nil {
		return fmt.Errorf("unable to flush datawriter: %+v", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("unable to close file properly: %+v", err)
	}

	fmt.Printf("[+] %s has been written\n", path.Join(cwd, "flast.users.txt"))

	return nil
}
