package xing

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"

	"github.com/machinebox/graphql"
)

const (
	empQuery string = `
query Employees(
	$id: SlugOrID!
	$first: Int
	$after: String
	$query: CompanyEmployeesQueryInput!
	$consumer: String! = ""
	$includeTotalQuery: Boolean = false
) {
	company(id: $id) {
		id
		totalEmployees: employees(first: 0, query: { consumer: $consumer })
			@include(if: $includeTotalQuery) {
			total
		}
		employees(first: $first, after: $after, query: $query) {
			total
			edges {
				node {
					contactDistance {
						distance
					}
					sharedContacts {
						total
					}
					profileDetails {
						id
						firstName
						lastName
						displayName
						gender
						pageName
						profileImage(size: SQUARE_256) {
							url
						}
						clickReasonProfileUrl(
							clickReasonId: CR_WEB_PUBLISHER_EMPLOYEES_MODULE
						) {
							profileUrl
						}
						userFlags {
							displayFlag
						}
						occupations {
							subline
						}
					}
				}
			}
		}
	}
}`
	/*
	   	empQuery string = `
	   query Employees($id: SlugOrID!, $first: Int, $after: String = "", $query: CompanyEmployeesQueryInput!, $consumer: String! = "", $includeTotalQuery: Boolean = true) {
	     company(id: $id) {
	       id
	       totalEmployees: employees(first: 0, query: {consumer: $consumer}) @include(if: $includeTotalQuery) {
	         total
	         __typename
	       }
	       employees(first: $first, after: $after, query: $query) {
	         total
	         edges {
	           node {
	             contactDistance {
	               distance
	               __typename
	             }
	             sharedContacts {
	               total
	               __typename
	             }
	             profileDetails {
	               id
	               firstName
	               lastName
	               displayName
	               gender
	               pageName
	               profileImage(size: SQUARE_256) {
	                 url
	                 __typename
	               }
	               clickReasonProfileUrl(clickReasonId: CR_WEB_PUBLISHER_EMPLOYEES_MODULE) {
	                 profileUrl
	                 __typename
	               }
	               userFlags {
	                 displayFlag
	                 __typename
	               }
	               occupations {
	                 subline
	                 __typename
	               }
	               __typename
	             }
	             networkRelationship {
	               id
	               relationship
	               permissions
	               error
	               __typename
	             }
	             __typename
	           }
	           __typename
	         }
	         pageInfo {
	           endCursor
	           hasNextPage
	           __typename
	         }
	         __typename
	       }
	       __typename
	     }
	   }`
	*/

	compQuery = `
query EntityPage($id: SlugOrID!) {
  entityPageEX(id: $id) {
    ... on EntityPage {
      context {
        companyId
      }
    }
  }
}`
)

type Data struct {
	Company Company `json:"Company"`
}

type Company struct {
	ID        string    `json:"id"`
	Employees Employees `json:"employees"`
}

type Employees struct {
	Total int    `json:"total"`
	Edges []Edge `json:"edges"`
}

type Edge struct {
	Node Node `json:"node"`
}

type Node struct {
	ContactDistance ContactDistance `json:"contactDistance"`
	SharedContacts  SharedContacts  `json:"sharedContacts"`
	ProfileDetails  ProfileDetails  `json:"profileDetails"`
}

type ContactDistance struct {
	Distance int `json:"distance"`
}

type SharedContacts struct {
	Total int `json:"total"`
}

type ProfileDetails struct {
	ID                    string                `json:"id"`
	FirstName             string                `json:"firstName"`
	LastName              string                `json:"lastName"`
	DisplayName           string                `json:"displayName"`
	Gender                string                `json:"gender"`
	PageName              string                `json:"pageName"`
	ProfileImage          []ProfileImage        `json:"profileImage"`
	ClickReasonProfileURL ClickReasonProfileURL `json:"clickReasonProfileUrl"`
	UserFlags             UserFlags             `json:"userFlags"`
	Occupations           []Occupations         `json:"occupations"`
}

type ProfileImage struct {
	URL string `json:"url"`
}

type ClickReasonProfileURL struct {
	ProfileURL string `json:"profileUrl"`
}

type UserFlags struct {
	DisplayFlag string `json:"displayFlag"`
}

type Occupations struct {
	Subline string `json:"subline"`
}

type Xing struct {
	Username string
	Password string
	Session  *http.Client
}

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Perm     string `json:"perm"`
}

func (x *Xing) Connect() error {
	url := "https://login.xing.com/login/api/login"

	ld := LoginData{
		Username: x.Username,
		Password: x.Password,
		Perm:     "1",
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	x.Session = &http.Client{
		Jar: jar,
	}

	json, err := json.Marshal(&ld)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := x.Session.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("\nLogin status: %+v\n", resp.Status)

	return nil
}

type QueryCompany struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Image string `json:"image"`
}

type Collection struct {
	Collection []QueryCompany `json:"collection"`
}

func (x *Xing) FindTarget(company string) (Collection, error) {
	var suggestions Collection

	base := "https://www.xing.com/dsc/suggestions/companies/name.json?consumer=loggedin.web.search.companies.name&query="

	url := fmt.Sprintf("%s%s", base, company)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return suggestions, err
	}

	resp, err := x.Session.Do(req)
	if err != nil {
		return suggestions, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return suggestions, err
	}

	if err := json.Unmarshal(data, &suggestions); err != nil {
		return suggestions, err
	}

	return suggestions, nil
}

type CompResponse struct {
	Data CompData `json:"data"`
}

type CompData struct {
	EntityPageEx EntityPageEx `json:"entityPageEX"`
}

type EntityPageEx struct {
	Context Context `json:"context"`
}

type Context struct {
	CompanyID string `json:"companyId"`
}

func (x *Xing) GQLExtractCompID(compname string) (string, error) {
	var id string
	var resp CompData
	gqlClient := graphql.NewClient("https://www.xing.com/xing-one/api")
	req := graphql.NewRequest(compQuery)

	req.Var("id", compname)

	if err := gqlClient.Run(context.Background(), req, &resp); err != nil {
		return "", err
	}

	id = resp.EntityPageEx.Context.CompanyID

	return id, nil
}

type Query struct {
	Consumer string `json:"consumer"`
	Sort     string `json:"sort"`
}

func (x *Xing) GQLExtractEmployees(id string) (Data, error) {
	var data Data
	limit := 100
	gqlClient := graphql.NewClient("https://www.xing.com/xing-one/api")
	req := graphql.NewRequest(empQuery)

	req.Var("consumer", "")
	req.Var("includeTotalQuery", true)
	req.Var("id", id)
	req.Var("first", limit)

	q := Query{
		Consumer: "web.entity_pages.employees_subpage",
		Sort:     "LAST_NAME",
	}
	req.Var("query", q)

	if err := gqlClient.Run(context.Background(), req, &data); err != nil {
		return data, err
	}

	if data.Company.Employees.Total > limit {
		// Redo with the actual limit before returning to fetch all employees at once
		req.Var("first", data.Company.Employees.Total)
		if err := gqlClient.Run(context.Background(), req, &data); err != nil {
			return data, err
		}
	}

	return data, nil
}

func SaveFirstLast(data Data) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path.Join(cwd, "first.last.users.txt"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	datawriter := bufio.NewWriter(file)

	for _, e := range data.Company.Employees.Edges {
		if e.Node.ProfileDetails.ID != "" {
			format := fmt.Sprintf("%s.%s\n", e.Node.ProfileDetails.FirstName, e.Node.ProfileDetails.LastName)
			_, err := datawriter.WriteString(format)
			if err != nil {
				return err
			}
		}
	}

	if err := datawriter.Flush(); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	fmt.Printf("%s has been written\n", path.Join(cwd, "first.last.users.txt"))

	return nil
}

func SaveOneLetterFirstLast(data Data) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path.Join(cwd, "flast.users.txt"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	datawriter := bufio.NewWriter(file)

	for _, e := range data.Company.Employees.Edges {
		if e.Node.ProfileDetails.ID != "" {
			format := fmt.Sprintf("%s.%s\n", string(e.Node.ProfileDetails.FirstName[0]), e.Node.ProfileDetails.LastName)
			_, err := datawriter.WriteString(format)
			if err != nil {
				return err
			}
		}
	}

	if err := datawriter.Flush(); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	fmt.Printf("%s has been written\n", path.Join(cwd, "flast.users.txt"))

	return nil
}
