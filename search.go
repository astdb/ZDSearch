package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/oleiade/reflections.v1"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func main() {
	// parse app config and get data file locations for reading
	log.Println("Reading config..")
	orgDataFile, userDataFile, ticketDataFile, err := GetAppConfig()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error reading config file: %v", err))
	}

	// read organization data from file
	OrgList, err := ReadOrganizationData(orgDataFile)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error reading org data file: %v", err))
	}

	// read user data from file
	UserList, err := ReadUserData(userDataFile)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error reading user data file: %v", err))
	}

	// read ticket data from file
	TicketList, err := ReadTicketData(ticketDataFile)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error reading ticket data file: %v", err))
	}

	fmt.Println("Building indexes...")

	// build indexes
	OrgUserIndex := indexOrgUsers(UserList)
	OrgTicketIndex := indexOrgTickets(TicketList)
	OrgIndex := indexOrgs(OrgList)
	UserSubmittedTixIndex := indexUserSubmittedTickets(TicketList)
	UserAssignedTixIndex := indexUserAssignedTickets(TicketList)
	UserIndex := indexUsers(UserList)

	fmt.Printf("%d organizations.\n", len(OrgList))
	fmt.Printf("%d users.\n", len(UserList))
	fmt.Printf("%d tickets.\n\n", len(TicketList))

	// provide search prompt to user on command line, running REPL-style until keyboard interrupt
	buffReader := bufio.NewReader(os.Stdin) // buffered reader to read console input
	prompt := "search >>"                   // console prompt text

	for {
		// show prompt
		fmt.Print(prompt)

		// read input from user
		searchInput, _ := buffReader.ReadString('\n')

		if strings.TrimSpace(searchInput) == "" {
			// no input - continue showing prompt
			continue

		} else {
			// search input received - evaluate

			// input format expected: <searchtype> <searchfield> <search value> (search value can be empty)
			searchType, searchField, searchValue, err := parseSearchInput(searchInput)

			if err != nil {
				// show error and prompt for next search input
				fmt.Printf("Error: %v\n", err)
				continue
			}

			validSearchType := false

			// search organizations
			if strings.ToLower(searchType) == "org" {
				validSearchType = true

				// get list of organizations matching this search criteria
				orgs, err := SearchOrgs(searchField, searchValue, OrgList)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

				// add associated users and tickets to returned search results
				orgsAugmented := getAssociatedUsersAndTickets(orgs, OrgUserIndex, OrgTicketIndex)

				// print search result
				fmt.Println(FormatOrgResult(orgsAugmented))
			}

			// search users
			if strings.ToLower(searchType) == "user" {
				validSearchType = true

				// get list of users matching this search criteria
				users, err := SearchUsers(searchField, searchValue, UserList)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

				// add associated orgs and tickets to returned search results
				usersAugmented := getAssociatedOrgsAndTickets(users, OrgIndex, UserSubmittedTixIndex, UserAssignedTixIndex)

				// print search result
				fmt.Println(FormatUserResult(usersAugmented))
			}

			if strings.ToLower(searchType) == "ticket" {
				validSearchType = true

				

				// get list of tickets matching this search criteria
				tickets, err := SearchTickets(searchField, searchValue, TicketList)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

				// func getAssociatedUsersAndOrgs(tickets []Ticket, userIndex map[int]User, orgIndex map[int]Organization) []Ticket
				ticketsAugmented := getAssociatedUsersAndOrgs(tickets, UserIndex, OrgIndex)

				// print search result
				fmt.Println(FormatTicketResult(ticketsAugmented))
			}

			if !validSearchType {
				fmt.Printf("Invalid search type: %s\n", searchType)
			}
		}
	}
}

// --------------- struct types to store Org/User/Ticket data and read application config ---------------------------

type Organization struct {
	ID                int      `json:"_id"`
	Name              string   `json:"name"`
	URL               string   `json:"url"`
	External_id       string   `json:"external_id"`
	DomainNames       []string `json:"domain_names"`
	Created_at        string   `json:"created_at"`
	Details           string   `json:"details"`
	Shared_tickets    bool     `json:"shared_tickets"`
	Tags              []string `json:"tags"`
	AssociatedUsers   []User
	AssociatedTickets []Ticket
}

type User struct {
	ID               int      `json:"_id"`
	Name             string   `json:"name"`
	URL              string   `json:"url"`
	External_id      string   `json:"external_id"`
	Alias            string   `json:"alias"`
	Created_at       string   `json:"created_at"`
	Active           bool     `json:"active"`
	Verified         bool     `json:"verified"`
	Shared           bool     `json:"shared"`
	Locale           string   `json:"locale"`
	Timezone         string   `json:"timezone"`
	Last_login_at    string   `json:"last_login_at"`
	Email            string   `json:"email"`
	Phone            string   `json:"phone"`
	Signature        string   `json:"signature"`
	Tags             []string `json:"tags"`
	Suspended        bool     `json:"suspended"`
	Role             string   `json:"role"`
	Org              int      `json:"organization_id"`
	OrgObject        Organization
	TicketsSubmitted []Ticket
	TicketsAssigned  []Ticket
}

type Ticket struct {
	ID            string   `json:"_id"`
	URL           string   `json:"url"`
	External_id   string   `json:"external_id"`
	Created_at    string   `json:"created_at"`
	Priority      string   `json:"priority"`
	Status        string   `json:"status"`
	Type          string   `json:"type"`
	Subject       string   `json:"subject"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Org           int      `json:"organization_id"`
	Has_incidents bool     `json:"has_incidents"`
	Due_at        string   `json:"due_at"`
	Submitter     int      `json:"submitter_id"`
	Assignee      int      `json:"assignee_id"`
	Via           string   `json:"via"`
	SubmitterObj  User
	AssigneeObj   User
	OrgObj        Organization
}

// struct to read in an application config file with locations of input data files
type AppConfig struct {
	OrgFileLocation    string `json:"OrgDataFileLocation"`
	UserFileLocation   string `json:"UserDataFileLocation"`
	TicketFileLocation string `json:"TicketDataileLocation"`
}

// -------------------- data indexing functions --------------------

// indexOrgUsers builds a map mapping organization ID's to users, enabling fast retrieval of all users belonging to a specific org
func indexOrgUsers(UserList []User) map[int][]User {
	orgUserIndex := map[int][]User{}

	for _, user := range UserList {
		_, indexed := orgUserIndex[user.Org]
		if !indexed {
			orgUserIndex[user.Org] = []User{user}
		} else {
			orgUserIndex[user.Org] = append(orgUserIndex[user.Org], user)
		}
	}

	return orgUserIndex
}

// indexOrgTickets builds a map mapping organization ID's to tickets, enabling fast retrieval of all tickets belonging to a specific org
func indexOrgTickets(TicketList []Ticket) map[int][]Ticket {
	ticketUserIndex := map[int][]Ticket{}

	for _, ticket := range TicketList {
		_, indexed := ticketUserIndex[ticket.Org]
		if !indexed {
			ticketUserIndex[ticket.Org] = []Ticket{ticket}
		} else {
			ticketUserIndex[ticket.Org] = append(ticketUserIndex[ticket.Org], ticket)
		}
	}

	return ticketUserIndex
}

// indexOrgs builds a map mapping Org IDs to Org objects, enabling efficient access to the parent Org objects via an Org ID
func indexOrgs(OrgList []Organization) map[int]Organization {
	orgIndex := map[int]Organization{}

	for _, org := range OrgList {
		_, indexed := orgIndex[org.ID]
		if !indexed {
			orgIndex[org.ID] = org
		}
	}

	return orgIndex
}

// indexUserSubmittedTickets builds a map mapping User ID's to submitted tickets, enabling efficient access to a list of tickets submitted by any user
func indexUserSubmittedTickets(TicketList []Ticket) map[int][]Ticket {
	userSubmittedIndex := map[int][]Ticket{}

	for _, ticket := range TicketList {
		userSubmittedIndex[ticket.Submitter] = append(userSubmittedIndex[ticket.Submitter], ticket)
	}

	return userSubmittedIndex
}

// indexUserAssignedTickets builds a map mapping User ID's to assigned tickets, enabling efficient access to a list of tickets assigned to any user
func indexUserAssignedTickets(TicketList []Ticket) map[int][]Ticket {
	userAssignedIndex := map[int][]Ticket{}

	for _, ticket := range TicketList {
		userAssignedIndex[ticket.Assignee] = append(userAssignedIndex[ticket.Assignee], ticket)
	}

	return userAssignedIndex
}

// indexUsers builds amap mapping user ID's to User objects, enabling efficient access to full user objects from a User ID
func indexUsers(UserList []User) map[int]User {
	userIndex := map[int]User{}

	for _, user := range UserList {
		_, indexed := userIndex[user.ID]

		if !indexed {
			userIndex[user.ID] = user
		}
	}

	return userIndex
}

// -------------------- primary type search functions (orgs/users/tickets) --------------------

func SearchOrgs(searchField, searchValue string, OrgList []Organization) ([]Organization, error) {
	results := []Organization{}
	for _, org := range OrgList {
		val, err := reflections.GetField(org, searchField)
		if err != nil {
			return results, err
		}

		searchFieldType, err := GetFieldType(org, searchField)
		if err != nil {
			return results, err
		}

		// depending on the type of the field to be searched on, check whether the search value matches

		// searching string fields
		if searchFieldType == "string" {
			if val == searchValue {
				results = append(results, org)
			}
		}

		// searching string array fields
		if searchFieldType == "[]string" {
			// Convert value into a string, strip off the opening and trailing square brackets, split on whitespace and iterate through it, but that would mean that
			// Tag field values cannot contain whitespaces.

			// get field value as a string
			valStr := fmt.Sprintf("%v", val)

			// strip off leading/trailing square brackets
			valStr = valStr[1 : len(valStr)-1]

			// split on whitespaec to get separate tags/domain names etc
			valStrComps := strings.Split(valStr, " ")

			for _, v := range valStrComps {
				if v == searchValue {
					results = append(results, org)
				}
			}
		}

		// searching boolean fields
		if searchFieldType == "bool" {
			searchBoolVal := false

			if strings.ToLower(searchValue) == "true" {
				searchBoolVal = true
			} else if strings.ToLower(searchValue) == "false" {
				searchBoolVal = false
			} else {
				// invalid search value
				return results, errors.New(fmt.Sprintf("Invalid search value for boolean field: Organization.%s is a %s field and search value (%s) must be boolean (true/false)", searchField, searchFieldType, searchValue))
			}

			if val == searchBoolVal {
				results = append(results, org)
			}
		}

		// searching an int field
		if searchFieldType == "int" {
			valInt, err := strconv.Atoi(fmt.Sprintf("%v", val))
			if err != nil {
				return results, err
			}

			searchInt, err := strconv.Atoi(searchValue)
			if err != nil {
				return results, err
			}

			if valInt == searchInt {
				results = append(results, org)
			}
		}

	}

	return results, nil
}

func SearchUsers(searchField, searchValue string, UserList []User) ([]User, error) {
	results := []User{}
	for _, user := range UserList {
		val, err := reflections.GetField(user, searchField)
		if err != nil {
			return results, err
		}

		searchFieldType, err := GetFieldType(user, searchField)
		if err != nil {
			return results, err
		}

		// depending on the type of the field to be searched on, check whether the search value matches

		// searching string fields
		if searchFieldType == "string" {
			if val == searchValue {
				results = append(results, user)
			}
		}

		// searching string array fields
		if searchFieldType == "[]string" {
			// Convert value into a string, strip off the opening and trailing square brackets, split on whitespace and iterate through it, but that would mean that
			// Tag field values cannot contain whitespaces.

			// get field value as a string
			valStr := fmt.Sprintf("%v", val)

			// strip off leading/trailing square brackets
			valStr = valStr[1 : len(valStr)-1]

			// split on whitespaec to get separate tags/domain names etc
			valStrComps := strings.Split(valStr, " ")

			for _, v := range valStrComps {
				if v == searchValue {
					results = append(results, user)
				}
			}
		}

		// searching boolean fields
		if searchFieldType == "bool" {
			searchBoolVal := false

			if strings.ToLower(searchValue) == "true" {
				searchBoolVal = true
			} else if strings.ToLower(searchValue) == "false" {
				searchBoolVal = false
			} else {
				// invalid search value
				return results, errors.New(fmt.Sprintf("Invalid search value for boolean field: User.%s is a %s field and search value (%s) must be boolean (true/false)", searchField, searchFieldType, searchValue))
			}

			if val == searchBoolVal {
				results = append(results, user)
			}
		}

		// searching an int field
		if searchFieldType == "int" {
			valInt, err := strconv.Atoi(fmt.Sprintf("%v", val))
			if err != nil {
				return results, err
			}

			searchInt, err := strconv.Atoi(searchValue)
			if err != nil {
				return results, err
			}

			if valInt == searchInt {
				results = append(results, user)
			}
		}
	}

	return results, nil
}

func SearchTickets(searchField, searchValue string, TicketList []Ticket) ([]Ticket, error) {
	results := []Ticket{}
	for _, ticket := range TicketList {
		val, err := reflections.GetField(ticket, searchField)
		if err != nil {
			return results, err
		}

		searchFieldType, err := GetFieldType(ticket, searchField)
		if err != nil {
			return results, err
		}

		// depending on the type of the field to be searched on, check whether the search value matches

		// searching string fields
		if searchFieldType == "string" {
			if val == searchValue {
				results = append(results, ticket)
			}
		}

		// searching string array fields
		if searchFieldType == "[]string" {
			// Convert value into a string, strip off the opening and trailing square brackets, split on whitespace and iterate through it, but that would mean that
			// Tag field values cannot contain whitespaces.

			// get field value as a string
			valStr := fmt.Sprintf("%v", val)

			// strip off leading/trailing square brackets
			valStr = valStr[1 : len(valStr)-1]

			// split on whitespace to get separate tags/domain names etc
			// Assumption: tags don't contain whitespace
			valStrComps := strings.Split(valStr, " ")

			for _, v := range valStrComps {
				if v == searchValue {
					results = append(results, ticket)
				}
			}
		}

		// searching boolean fields
		if searchFieldType == "bool" {
			searchBoolVal := false

			if strings.ToLower(searchValue) == "true" {
				searchBoolVal = true
			} else if strings.ToLower(searchValue) == "false" {
				searchBoolVal = false
			} else {
				// invalid search value
				return results, errors.New(fmt.Sprintf("Invalid search value for boolean field: Ticket.%s is a %s field and search value (%s) must be boolean (true/false)", searchField, searchFieldType, searchValue))
			}

			if val == searchBoolVal {
				results = append(results, ticket)
			}
		}

		// searching an int field
		if searchFieldType == "int" {
			valInt, err := strconv.Atoi(fmt.Sprintf("%v", val))
			if err != nil {
				return results, err
			}

			searchInt, err := strconv.Atoi(searchValue)
			if err != nil {
				return results, err
			}

			if valInt == searchInt {
				results = append(results, ticket)
			}
		}
	}

	return results, nil
}

// -------------------- associated entity search functions -----------------------------

// getAssociatedUsersAndTickets accepts a list of organization objects and for each, populates the associated user and ticket fields
// func getAssociatedUsersAndTickets(orgs []Organization, UserList []User, TicketList []Ticket) []Organization {
func getAssociatedUsersAndTickets(orgs []Organization, orgUserIndex map[int][]User, indexOrgTickets map[int][]Ticket) []Organization {
	orgResults := []Organization{}
	for _, org := range orgs {
		// get users associated with this org
		assocUsers, indexed := orgUserIndex[org.ID]
		if indexed {
			org.AssociatedUsers = assocUsers
		} else {
			org.AssociatedUsers = []User{}
		}

		// get tickets associated with this org
		assocTickets, indexed := indexOrgTickets[org.ID]
		if indexed {
			org.AssociatedTickets = assocTickets
		} else {
			org.AssociatedTickets = []Ticket{}
		}

		orgResults = append(orgResults, org)
	}

	// return orgs
	return orgResults
}

// getAssociatedOrgsAndTickets accepts a list of user objects and for each, populates the associated org and ticket fields
func getAssociatedOrgsAndTickets(users []User, orgIndex map[int]Organization, userSubmittedIndex map[int][]Ticket, userAssignedIndex map[int][]Ticket) []User {
	userResults := []User{}
	for _, user := range users {
		org, indexed := orgIndex[user.Org]
		if indexed {
			user.OrgObject = org
		}

		submittedTix, indexed := userSubmittedIndex[user.ID]
		if indexed {
			user.TicketsSubmitted = submittedTix
		}

		assignedTix, indexed := userAssignedIndex[user.ID]
		if indexed {
			user.TicketsAssigned = assignedTix
		}

		userResults = append(userResults, user)
	}

	// return orgs
	return userResults
}

// getAssociatedUsersAndOrgs accepts a list of ticket objects and for each, populates the associated user and org fields
func getAssociatedUsersAndOrgs(tickets []Ticket, userIndex map[int]User, orgIndex map[int]Organization) []Ticket {
	ticketResults := []Ticket{}

	for _, ticket := range tickets {
		userSubmitted, indexed := userIndex[ticket.Submitter]
		if indexed {
			ticket.SubmitterObj = userSubmitted
		}

		userAssigned, indexed := userIndex[ticket.Assignee]
		if indexed {
			ticket.AssigneeObj = userAssigned
		}

		org, indexed := orgIndex[ticket.Org]
		if indexed {
			ticket.OrgObj = org
		}

		ticketResults = append(ticketResults, ticket)
	}

	// return orgs
	return ticketResults
}

// ----------------------- result formatting functions -----------------------------

func FormatOrgResult(orgs []Organization) string {
	var formattedResult strings.Builder
	formattedResult.WriteString("\nORGS\n----\n")

	if len(orgs) <= 0 {
		formattedResult.WriteString("<No results found>\n")
		return formattedResult.String()
	}

	for _, org := range orgs {
		formattedResult.WriteString(fmt.Sprintf("\nOrganization ID: %d\nName: %s\nURLs: %s\nExternal_ID: %s\nDomain Names: %s\nCreated At: %s\nDetails:  %s\nShared Tickets: %v\nTags: %v\n\n", org.ID, org.Name, org.URL, org.External_id, org.DomainNames, org.Created_at, org.Details, org.Shared_tickets, org.Tags))

		formattedResult.WriteString("\tASSOCIATED USERS\n\t----------------\n")
		if len(org.AssociatedUsers) > 0 {
			for _, user := range org.AssociatedUsers {
				formattedResult.WriteString(fmt.Sprintf("\n\tID: %d\n\tName: %s\n\tURL: %s\n\tExternal ID: %s\n\tAlias: %s\n\tCreated At: %s\n\tActive: %v\n\tVerified: %v\n\tShared: %v\n\tLocale: %s\n\tTime Zone: %s\n\tLast Login At: %s\n\tEmail: %s\n\tPhone: %s\n\tSignature: %s\n\tTags: %v\n\tSuspended: %v\n\tRole: %s\n\tOrganization: %d\n\n", user.ID, user.Name, user.URL, user.External_id, user.Alias, user.Created_at, user.Active, user.Verified, user.Shared, user.Locale, user.Timezone, user.Last_login_at, user.Email, user.Phone, user.Signature, user.Tags, user.Suspended, user.Role, user.Org))
			}
		} else {
			formattedResult.WriteString("\t<No associated users found for this organization>\n")
		}

		formattedResult.WriteString("\n\tASSOCIATED TICKETS\n\t------------------\n")
		if len(org.AssociatedTickets) > 0 {
			for _, ticket := range org.AssociatedTickets {
				formattedResult.WriteString(fmt.Sprintf("\n\tTicket ID: %s\n\tURL: %s\n\tExternal ID: %s\n\tCreated At: %s\n\tPriority: %s\n\tStatus: %s\n\tType: %s\n\tSubject: %s\n\tDescription: %s\n\tTags: %v\n\tOrganization: %d\n\tHas Incidents: %v\n\tDue At: %s\n\tSubmitter: %d\n\tAssignee: %d\n\tVia: %s\n\n", ticket.ID, ticket.URL, ticket.External_id, ticket.Created_at, ticket.Priority, ticket.Status, ticket.Type, ticket.Subject, ticket.Description, ticket.Tags, ticket.Org, ticket.Has_incidents, ticket.Due_at, ticket.Submitter, ticket.Assignee, ticket.Via))
			}
		} else {
			formattedResult.WriteString("\t<No associated tickets found for this organization>\n")
		}
	}

	return formattedResult.String()
}

func FormatUserResult(users []User) string {
	var formattedResult strings.Builder
	formattedResult.WriteString("\nUSERS\n-----\n")

	if len(users) <= 0 {
		formattedResult.WriteString("<No results found>\n")
		return formattedResult.String()
	}

	for _, user := range users {
		formattedResult.WriteString(fmt.Sprintf("\nID: %d\nName: %s\nURL: %s\nExternal ID: %s\nAlias: %s\nCreated At: %s\nActive: %v\nVerified: %v\nShared: %v\nLocale: %s\nTime Zone: %s\nLast Login At: %s\nEmail: %s\nPhone: %s\nSignature: %s\nTags: %v\nSuspended: %v\nRole: %s\nOrganization: %d\n\n", user.ID, user.Name, user.URL, user.External_id, user.Alias, user.Created_at, user.Active, user.Verified, user.Shared, user.Locale, user.Timezone, user.Last_login_at, user.Email, user.Phone, user.Signature, user.Tags, user.Suspended, user.Role, user.Org))

		formattedResult.WriteString("\tASSOCIATED ORGS\n\t----------------\n")

		formattedResult.WriteString(fmt.Sprintf("\n\tOrganization ID: %d\n\tName: %s\n\tURLs: %s\n\tExternal_ID: %s\n\tDomain Names: %s\n\tCreated At: %s\n\tDetails:  %s\n\tShared Tickets: %v\n\tTags: %v\n\n", user.OrgObject.ID, user.OrgObject.Name, user.OrgObject.URL, user.OrgObject.External_id, user.OrgObject.DomainNames, user.OrgObject.Created_at, user.OrgObject.Details, user.OrgObject.Shared_tickets, user.OrgObject.Tags))

		formattedResult.WriteString("\n\tTICKETS (SUBMITTED)\n\t-------------------\n")
		if len(user.TicketsSubmitted) > 0 {
			for _, ticket := range user.TicketsSubmitted {
				formattedResult.WriteString(fmt.Sprintf("\n\tTicket ID: %s\n\tURL: %s\n\tExternal ID: %s\n\tCreated At: %s\n\tPriority: %s\n\tStatus: %s\n\tType: %s\n\tSubject: %s\n\tDescription: %s\n\tTags: %v\n\tOrganization: %d\n\tHas Incidents: %v\n\tDue At: %s\n\tSubmitter: %d\n\tAssignee: %d\n\tVia: %s\n\n", ticket.ID, ticket.URL, ticket.External_id, ticket.Created_at, ticket.Priority, ticket.Status, ticket.Type, ticket.Subject, ticket.Description, ticket.Tags, ticket.Org, ticket.Has_incidents, ticket.Due_at, ticket.Submitter, ticket.Assignee, ticket.Via))
			}
		} else {
			formattedResult.WriteString("\t<No submitted tickets found for this user>\n")
		}

		formattedResult.WriteString("\n\tTICKETS (ASSIGNED)\n\t------------------\n")
		if len(user.TicketsAssigned) > 0 {
			for _, ticket := range user.TicketsAssigned {
				formattedResult.WriteString(fmt.Sprintf("\n\tTicket ID: %s\n\tURL: %s\n\tExternal ID: %s\n\tCreated At: %s\n\tPriority: %s\n\tStatus: %s\n\tType: %s\n\tSubject: %s\n\tDescription: %s\n\tTags: %v\n\tOrganization: %d\n\tHas Incidents: %v\n\tDue At: %s\n\tSubmitter: %d\n\tAssignee: %d\n\tVia: %s\n\n", ticket.ID, ticket.URL, ticket.External_id, ticket.Created_at, ticket.Priority, ticket.Status, ticket.Type, ticket.Subject, ticket.Description, ticket.Tags, ticket.Org, ticket.Has_incidents, ticket.Due_at, ticket.Submitter, ticket.Assignee, ticket.Via))
			}
		} else {
			formattedResult.WriteString("\t<No assigned tickets found for this user>\n")
		}
	}

	return formattedResult.String()
}

func FormatTicketResult(tickets []Ticket) string {
	var formattedResult strings.Builder
	formattedResult.WriteString("\nTICKETS\n-------\n")

	if len(tickets) <= 0 {
		formattedResult.WriteString("<No results found>\n")
		return formattedResult.String()
	}

	for _, ticket := range tickets {
		formattedResult.WriteString(fmt.Sprintf("\nTicket ID: %s\nURL: %s\nExternal ID: %s\nCreated At: %s\nPriority: %s\nStatus: %s\nType: %s\nSubject: %s\nDescription: %s\nTags: %v\nOrganization: %d\nHas Incidents: %v\nDue At: %s\nSubmitter: %d\nAssignee: %d\nVia: %s\n\n", ticket.ID, ticket.URL, ticket.External_id, ticket.Created_at, ticket.Priority, ticket.Status, ticket.Type, ticket.Subject, ticket.Description, ticket.Tags, ticket.Org, ticket.Has_incidents, ticket.Due_at, ticket.Submitter, ticket.Assignee, ticket.Via))

		formattedResult.WriteString("\tASSOCIATED ORGS\n\t----------------\n")
		formattedResult.WriteString(fmt.Sprintf("\n\tOrganization ID: %d\n\tName: %s\n\tURLs: %s\n\tExternal ID: %s\n\tDomain Names: %s\n\tCreated At: %s\n\tDetails:  %s\n\tShared Tickets: %v\n\tTags: %v\n\n", ticket.OrgObj.ID, ticket.OrgObj.Name, ticket.OrgObj.URL, ticket.OrgObj.External_id, ticket.OrgObj.DomainNames, ticket.OrgObj.Created_at, ticket.OrgObj.Details, ticket.OrgObj.Shared_tickets, ticket.OrgObj.Tags))

		formattedResult.WriteString("\n\tASSOCIATED USERS (SUBMITTER)\n\t----------------------------\n")
		formattedResult.WriteString(fmt.Sprintf("\n\tID: %d\n\tName: %s\n\tURL: %s\n\tExternal ID: %s\n\tAlias: %s\n\tCreated At: %s\n\tActive: %v\n\tVerified: %v\n\tShared: %v\n\tLocale: %s\n\tTime Zone: %s\n\tLast Login At: %s\n\tEmail: %s\n\tPhone: %s\n\tSignature: %s\n\tTags: %v\n\tSuspended: %v\n\tRole: %s\n\tOrganization: %d\n\n", ticket.SubmitterObj.ID, ticket.SubmitterObj.Name, ticket.SubmitterObj.URL, ticket.SubmitterObj.External_id, ticket.SubmitterObj.Alias, ticket.SubmitterObj.Created_at, ticket.SubmitterObj.Active, ticket.SubmitterObj.Verified, ticket.SubmitterObj.Shared, ticket.SubmitterObj.Locale, ticket.SubmitterObj.Timezone, ticket.SubmitterObj.Last_login_at, ticket.SubmitterObj.Email, ticket.SubmitterObj.Phone, ticket.SubmitterObj.Signature, ticket.SubmitterObj.Tags, ticket.SubmitterObj.Suspended, ticket.SubmitterObj.Role, ticket.SubmitterObj.Org))

		formattedResult.WriteString("\n\tASSOCIATED USERS (ASSIGNEE)\n\t---------------------------\n")
		formattedResult.WriteString(fmt.Sprintf("\n\tID: %d\n\tName: %s\n\tURL: %s\n\tExternal ID: %s\n\tAlias: %s\n\tCreated At: %s\n\tActive: %v\n\tVerified: %v\n\tShared: %v\n\tLocale: %s\n\tTime Zone: %s\n\tLast Login At: %s\n\tEmail: %s\n\tPhone: %s\n\tSignature: %s\n\tTags: %v\n\tSuspended: %v\n\tRole: %s\n\tOrganization: %d\n\n", ticket.AssigneeObj.ID, ticket.AssigneeObj.Name, ticket.AssigneeObj.URL, ticket.AssigneeObj.External_id, ticket.AssigneeObj.Alias, ticket.AssigneeObj.Created_at, ticket.AssigneeObj.Active, ticket.AssigneeObj.Verified, ticket.AssigneeObj.Shared, ticket.AssigneeObj.Locale, ticket.AssigneeObj.Timezone, ticket.AssigneeObj.Last_login_at, ticket.AssigneeObj.Email, ticket.AssigneeObj.Phone, ticket.AssigneeObj.Signature, ticket.AssigneeObj.Tags, ticket.AssigneeObj.Suspended, ticket.AssigneeObj.Role, ticket.AssigneeObj.Org))
	}

	return formattedResult.String()
}

// -------------------- data loader functions --------------------------

// ReadUserData reads in user data from a given file, and returns a list of User objects (and an error if required)
func ReadUserData(fileName string) ([]User, error) {
	userData, _ := ioutil.ReadFile(fileName)

	var userList []User
	err := json.Unmarshal(userData, &userList)
	if err != nil {
		return userList, err
	}

	return userList, nil
}

// ReadTicketData reads in ticket data from a given file, and returns a list of Ticket objects (and an error if required)
func ReadTicketData(fileName string) ([]Ticket, error) {
	ticketData, _ := ioutil.ReadFile(fileName)

	var ticketList []Ticket
	err := json.Unmarshal(ticketData, &ticketList)
	if err != nil {
		return ticketList, err
	}

	return ticketList, nil
}

// ReadOrganizationData reads in organization data from a given file, and returns a list of Organization objects (and an error if required)
func ReadOrganizationData(fileName string) ([]Organization, error) {
	orgData, _ := ioutil.ReadFile(fileName)

	var orgList []Organization
	err := json.Unmarshal(orgData, &orgList)
	if err != nil {
		return orgList, err
	}

	return orgList, nil
}

// ---------------------- command line input capture ----------------------------

func parseSearchInput(searchInput string) (string, string, string, error) {
	// start by removing the trailing newline and splitting search input into tokens by whitespace
	// input format expected: <searchtype> <searchfield> <search value> (search value can be empty)

	searchInputComps := strings.Split(searchInput[:len(searchInput)-1], " ")

	// search must at least supply a non-empty search type, and search field
	if len(searchInputComps) >= 2 {
		searchType := strings.TrimSpace(searchInputComps[0])
		searchField := strings.TrimSpace(searchInputComps[1])

		if len(searchInputComps) > 2 {
			var searchValue strings.Builder

			for i := 2; i < len(searchInputComps); i++ {
				searchValue.WriteString(fmt.Sprintf("%s ", searchInputComps[i]))
			}

			return searchType, searchField, strings.TrimSpace(searchValue.String()), nil
		} else {
			// empty search term
			return searchType, searchField, "", nil
		}
	} else {

		return "", "", "", errors.New("Invalid search format. Search format: $> <searchtype> <searchfield> <search values>")
	}
}

// ------------------------- App config ---------------------------------
// read application config file and return locations of org/user/data files
func GetAppConfig() (string, string, string, error) {
	file, _ := os.Open("config.json")
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := AppConfig{}
	err := decoder.Decode(&config)
	if err != nil {
		return "", "", "", err
	}

	return config.OrgFileLocation, config.UserFileLocation, config.TicketFileLocation, nil
}

// ----------------- reflect functions to get struct field types at runtime --------------------

func GetFieldType(obj interface{}, name string) (string, error) {
	if !hasValidType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return "", errors.New("Cannot use GetField on a non-struct interface")
	}

	objValue := reflectValue(obj)
	field := objValue.FieldByName(name)

	if !field.IsValid() {
		return "", fmt.Errorf("No such field: %s in obj", name)
	}

	return field.Type().String(), nil
}

func hasValidType(obj interface{}, types []reflect.Kind) bool {
	for _, t := range types {
		if reflect.TypeOf(obj).Kind() == t {
			return true
		}
	}

	return false
}

func reflectValue(obj interface{}) reflect.Value {
	var val reflect.Value

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}

	return val
}
