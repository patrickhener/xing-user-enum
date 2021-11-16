package xing

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
			pageInfo {
				endCursor
				hasNextPage
			}
		}
	}
}`

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
