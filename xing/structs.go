package xing

import "net/http"

type Data struct {
	Company Company `json:"Company"`
}

type Company struct {
	ID        string    `json:"id"`
	Employees Employees `json:"employees"`
}

type Employees struct {
	Total    int      `json:"total"`
	Edges    []Edge   `json:"edges"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Edge struct {
	Node Node `json:"node"`
}

type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
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
	Proxy    string
}

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Perm     string `json:"perm"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Target  string `json:"target"`
}

// type QueryCompany struct {
// 	ID    int    `json:"id"`
// 	Title string `json:"title"`
// 	Image string `json:"image"`
// }

// type Collection struct {
// 	Collection []QueryCompany `json:"collection"`
// }

type RespCompany struct {
	Count string `json:"count"`
	Items []Item `json:"items"`
}

type Item struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`
	Slug  string
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

type Query struct {
	Consumer string `json:"consumer"`
	Sort     string `json:"sort"`
}
