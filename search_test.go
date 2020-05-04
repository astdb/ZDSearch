package main

import (
	"fmt"
	"testing"
)

func TestIndexOrgUsers(t *testing.T) {
	_, userDataFile, _, err := GetAppConfig()
	if err != nil {
		t.Error("Couln't read config file for TestIndexOrgUsers\n")
	}

	UserList, err := ReadUserData(userDataFile)
	if err != nil {
		t.Error("Couln't get user list for TestIndexOrgUsers\n")
	}

	OrgUserIndex := indexOrgUsers(UserList)

	if len(OrgUserIndex) != 25 {
		t.Error("TestIndexOrgUsers: indexing incorrect.\n")
	}

	// check a specific org/users mapping to check data integrity
	users, indexed := OrgUserIndex[104]
	if !indexed {
		t.Error("TestIndexOrgUsers: org not found in index.\n")
	}

	if len(users) != 4 {
		t.Error("TestIndexOrgUsers: users not mapped to org ID correctly.\n")
	}
}

func TestIndexOrgTickets(t *testing.T) {
	_, _, ticketDataFile, err := GetAppConfig()
	if err != nil {
		t.Error("TestIndexOrgTickets: cannot read config file.\n")
	}

	TicketList, err := ReadTicketData(ticketDataFile)
	if err != nil {
		t.Error("TestIndexOrgTickets: cannot get ticket list.\n")
	}

	orgTicketIndex := indexOrgTickets(TicketList)

	if len(orgTicketIndex) != 27 {
		t.Error("TestIndexOrgTickets: indexing incorrect.\n")
	}

	// check a specific org/ticket mapping to check data integrity
	tickets, indexed := orgTicketIndex[104]
	if !indexed {
		t.Error("TestIndexOrgTickets: org not found in index.\n")
	}

	if len(tickets) != 7 {
		t.Error("TestIndexOrgTickets: tickets not mapped to org ID correctly.\n")
	}
}

func TestIndexOrgs(t *testing.T) {
	orgDataFile, _, _, err := GetAppConfig()
	if err != nil {
		t.Error("TestIndexOrgs: cannot read config file.\n")
	}

	OrgList, err := ReadOrganizationData(orgDataFile)
	if err != nil {
		t.Error("TestIndexOrgs: cannot get org list.\n")
	}

	orgIndex := indexOrgs(OrgList)

	if len(orgIndex) != 25 {
		t.Error("TestIndexOrgs: indexing incorrect.\n")
	}

	// check a specific org mapping to check data integrity
	org, indexed := orgIndex[104]
	if !indexed {
		t.Error("TestIndexOrgs: org not found in index.\n")
	}

	if org.ID != 104 || org.URL != "http://initech.zendesk.com/api/v2/organizations/104.json" || org.Name != "Xylar" || org.Created_at != "2016-03-21T10:11:18 -11:00" || org.Details != "MegaCörp" || org.Shared_tickets != false {
		t.Error("TestIndexOrgs: incorrect org details in index.\n")
	}
}

func TestIndexUserSubmittedTickets(t *testing.T) {
	_, _, ticketDataFile, err := GetAppConfig()
	if err != nil {
		t.Error("TestIndexUserSubmittedTickets: cannot read config file.\n")
	}

	TicketList, err := ReadTicketData(ticketDataFile)
	if err != nil {
		t.Error("TestIndexUserSubmittedTickets: cannot get ticket list.\n")
	}

	UserSubmittedTixIndex := indexUserSubmittedTickets(TicketList)

	if len(UserSubmittedTixIndex) != 67 {
		t.Error("TestIndexUserSubmittedTickets: indexing incorrect.\n")
	}

	// check a specific org mapping to check data integrity
	tix, indexed := UserSubmittedTixIndex[49]
	if !indexed {
		t.Error("TestIndexUserSubmittedTickets: user not found in index.\n")
	}

	if len(tix) != 2 {
		t.Error("TestIndexUserSubmittedTickets: tickets not mapped correctly to user in index.\n")
	}
}

func TestIndexUserAssignedTickets(t *testing.T) {
	_, _, ticketDataFile, err := GetAppConfig()
	if err != nil {
		t.Error("TestIndexUserAssignedTickets: cannot read config file.\n")
	}

	TicketList, err := ReadTicketData(ticketDataFile)
	if err != nil {
		t.Error("TestIndexUserAssignedTickets: cannot get ticket list.\n")
	}

	UserAssignedTixIndex := indexUserAssignedTickets(TicketList)

	if len(UserAssignedTixIndex) != 72 {
		t.Error("TestIndexUserAssignedTickets: indexing incorrect.\n")
	}

	// check a specific org mapping to check data integrity
	tix, indexed := UserAssignedTixIndex[49]
	if !indexed {
		t.Error("TestIndexUserAssignedTickets: user not found in index.\n")
	}

	if len(tix) != 2 {
		t.Error("TestIndexUserAssignedTickets: tickets not mapped correctly to user in index.\n")
	}
}

func TestSearchOrgs(t *testing.T) {
	orgDataFile, _, _, err := GetAppConfig()
	if err != nil {
		t.Error("TestSearchOrgs: cannot read config file.\n")
	}

	OrgList, err := ReadOrganizationData(orgDataFile)
	if err != nil {
		t.Error("TestSearchOrgs: cannot get org list.\n")
	}

	searchField := "ID"
	searchValue := "103"
	orgs, err := SearchOrgs(searchField, searchValue, OrgList)
	if err != nil {
		t.Error(fmt.Sprintf("TestSearchOrgs: SearchOrgs error - %v\n", err))
	}

	if len(orgs) <= 0 || orgs[0].ID != 103 || orgs[0].URL != "http://initech.zendesk.com/api/v2/organizations/103.json" || orgs[0].External_id != "e73240f3-8ecf-411d-ad0d-80ca8a84053d" || orgs[0].Name != "Plasmos" || orgs[0].Created_at != "2016-05-28T04:40:37 -10:00" || orgs[0].Details != "Non profit" || orgs[0].Shared_tickets != false {
		t.Error(fmt.Sprintf("TestSearchOrgs: SearchOrgs error - incorrect search result."))
	}

	// TODO: tests for searching on 'every' field

}

func TestSearchUsers(t *testing.T) {
	_, userDataFile, _, err := GetAppConfig()
	if err != nil {
		t.Error(fmt.Sprintf("TestSearchUsers: cannot read config file - %v\n", err))
	}

	UserList, err := ReadUserData(userDataFile)
	if err != nil {
		t.Error(fmt.Sprintf("TestSearchUsers: cannot get user list - %v\n", err))
	}

	searchField := "ID"
	searchValue := "49"
	users, err := SearchUsers(searchField, searchValue, UserList)
	if err != nil {
		t.Error(fmt.Sprintf("TestSearchUsers: error searching users - %v\n", err))
	}

	if len(users) <= 0 || users[0].ID != 49 || users[0].URL != "http://initech.zendesk.com/api/v2/users/49.json" || users[0].External_id != "4bd5e757-c0cd-445b-b702-ee3ed794f6c4" || users[0].Name != "Faulkner Holcomb" || users[0].Alias != "Miss Jody" || users[0].Created_at != "2016-05-12T08:39:30 -10:00" || users[0].Active != true || users[0].Verified != false || users[0].Shared != true || users[0].Locale != "zh-CN" || users[0].Timezone != "Antigua and Barbuda" || users[0].Last_login_at != "2014-12-04T12:51:36 -11:00" || users[0].Email != "jodyholcomb@flotonic.com" || users[0].Phone != "9255-943-719" || users[0].Signature != "Don't Worry Be Happy!" || users[0].Org != 118 || users[0].Suspended != true || users[0].Role != "end-user" {
		t.Error(fmt.Sprintf("TestSearchUsers: SearchUsers error - incorrect search result."))
	}

	// TODO: tests for searching on 'every' field

}

func TestSearchTickets(t *testing.T) {
	_, _, ticketDataFile, err := GetAppConfig()
	if err != nil {
		t.Error("TestSearchTickets: cannot read config file.\n")
	}

	TicketList, err := ReadTicketData(ticketDataFile)
	if err != nil {
		t.Error("TestSearchTickets: cannot get ticket list.\n")
	}

	searchField := "ID"
	searchValue := "1a227508-9f39-427c-8f57-1b72f3fab87c"
	tickets, err := SearchTickets(searchField, searchValue, TicketList)
	if err != nil {
		t.Error(fmt.Sprintf("TestSearchTickets: error searching tickets - %v\n", err))
	}



	if len(tickets) <= 0 || tickets[0].ID != "1a227508-9f39-427c-8f57-1b72f3fab87c" || tickets[0].URL != "http://initech.zendesk.com/api/v2/tickets/1a227508-9f39-427c-8f57-1b72f3fab87c.json" || tickets[0].External_id != "3e5ca820-cd1f-4a02-a18f-11b18e7bb49a" || tickets[0].Created_at != "2016-04-14T08:32:31 -10:00" || tickets[0].Type != "incident" || tickets[0].Subject != "A Catastrophe in Micronesia" || tickets[0].Description != "Aliquip excepteur fugiat ex minim ea aute eu labore. Sunt eiusmod esse eu non commodo est veniam consequat." || tickets[0].Priority != "low" || tickets[0].Status != "hold" || tickets[0].Submitter != 71 || tickets[0].Assignee != 38 || tickets[0].Org != 112 || tickets[0].Has_incidents != false || tickets[0].Due_at != "2016-08-15T05:37:32 -10:00" || tickets[0].Via != "chat" {
		t.Error(fmt.Sprintf("TestSearchTickets: SearchTickets error - incorrect search result."))
	}
}

// test multiple ticket search
func TestMultipleSearchTickets(t *testing.T) {
	_, _, ticketDataFile, err := GetAppConfig()
	if err != nil {
		t.Error("TestSearchTickets: cannot read config file.\n")
	}

	TicketList, err := ReadTicketData(ticketDataFile)
	if err != nil {
		t.Error("TestSearchTickets: cannot get ticket list.\n")
	}

	searchField := "ID"
	searchValue := "1a227508-9f39-427c-8f57-1b72f3fab87c, 4cce7415-ef12-42b6-b7b5-fb00e24f9cc1"
	tickets, err := SearchTickets(searchField, searchValue, TicketList)
	if err != nil {
		t.Error(fmt.Sprintf("TestSearchTickets: error searching tickets - %v\n", err))
	}

	if len(tickets) != 2 {
		t.Error(fmt.Sprintf("TestSearchTickets: multiple ticket search returned wrong length: %d\n", len(tickets)))
	}

	// for _, ticket := range tickets {
	// 	if ticket.ID == 
	// }
}

func TestGetAssociatedUsersAndTickets(t *testing.T) {
	orgDataFile, userDataFile, ticketDataFile, err := GetAppConfig()
	if err != nil {
		t.Error("TestGetAssociatedUsersAndTickets: cannot read config file.\n")
	}

	OrgList, err := ReadOrganizationData(orgDataFile)
	if err != nil {
		t.Error("TestGetAssociatedUsersAndTickets: cannot get org list.\n")
	}

	UserList, err := ReadUserData(userDataFile)
	if err != nil {
		t.Error("TestGetAssociatedUsersAndTickets: cannot get user list\n")
	}

	TicketList, err := ReadTicketData(ticketDataFile)
	if err != nil {
		t.Error("TestGetAssociatedUsersAndTickets: cannot get ticket list.\n")
	}

	OrgTicketIndex := indexOrgTickets(TicketList)
	OrgUserIndex := indexOrgUsers(UserList)

	searchField := "ID"
	searchValue := "103"
	orgs, err := SearchOrgs(searchField, searchValue, OrgList)
	if err != nil {
		t.Error(fmt.Sprintf("TestGetAssociatedUsersAndTickets: SearchOrgs error - %v\n", err))
	}

	orgsAugmented := getAssociatedUsersAndTickets(orgs, OrgUserIndex, OrgTicketIndex)

	if len(orgsAugmented) <= 0 {
		t.Error("TestGetAssociatedUsersAndTickets: incorrect length for augmented org search results.\n")
	}

	assocUsers := orgsAugmented[0].AssociatedUsers
	assocTix := orgsAugmented[0].AssociatedTickets

	if len(assocUsers) != 3 || len(assocTix) != 6 {
		t.Error("TestGetAssociatedUsersAndTickets: org search results not augmented correctly with associated user/ticket results.\n")
	}
}

func TestGetAssociatedUsersAndOrgs(t *testing.T) {
	orgDataFile, userDataFile, ticketDataFile, err := GetAppConfig()
	if err != nil {
		t.Error("TestGetAssociatedUsersAndOrgs: cannot read config file.\n")
	}

	OrgList, err := ReadOrganizationData(orgDataFile)
	if err != nil {
		t.Error("TestGetAssociatedUsersAndOrgs: cannot get org list.\n")
	}

	UserList, err := ReadUserData(userDataFile)
	if err != nil {
		t.Error("TestGetAssociatedUsersAndOrgs: cannot get user list\n")
	}

	TicketList, err := ReadTicketData(ticketDataFile)
	if err != nil {
		t.Error("TestGetAssociatedUsersAndTickets: cannot get ticket list.\n")
	}

	OrgIndex := indexOrgs(OrgList)
	UserIndex := indexUsers(UserList)

	searchField := "ID"
	searchValue := "3584e2c9-ccd4-4acb-9419-9245891cf398"
	tickets, err := SearchTickets(searchField, searchValue, TicketList)
	if err != nil {
		t.Error(fmt.Sprintf("TestGetAssociatedUsersAndOrgs: error searching tickets - %v\n", err))
	}

	// func getAssociatedUsersAndOrgs(tickets []Ticket, userIndex map[int]User, orgIndex map[int]Organization) []Ticket
	ticketsAugmented := getAssociatedUsersAndOrgs(tickets, UserIndex, OrgIndex)

	if len(ticketsAugmented) <= 0 {
		t.Error("TestGetAssociatedUsersAndOrgs: incorrect length for augmented ticket search results.\n")
	}

	submitter := ticketsAugmented[0].SubmitterObj
	assignee := ticketsAugmented[0].AssigneeObj
	org := ticketsAugmented[0].OrgObj

	if submitter.ID != 10 || assignee.ID != 47 || org.ID != 103 {
		t.Error("TestGetAssociatedUsersAndOrgs: org search results not augmented correctly with associated user/ticket results.\n")
	}
}

func TestGetAssociatedOrgsAndTickets(t *testing.T) {
	orgDataFile, userDataFile, ticketDataFile, err := GetAppConfig()
	if err != nil {
		t.Error("TestgetAssociatedOrgsAndTickets: cannot read config file.\n")
	}

	OrgList, err := ReadOrganizationData(orgDataFile)
	if err != nil {
		t.Error("TestgetAssociatedOrgsAndTickets: cannot get org list.\n")
	}

	UserList, err := ReadUserData(userDataFile)
	if err != nil {
		t.Error("TestgetAssociatedOrgsAndTickets: cannot get org list.\n")
	}

	TicketList, err := ReadTicketData(ticketDataFile)
	if err != nil {
		t.Error("TestgetAssociatedOrgsAndTickets: cannot get ticket list.\n")
	}

	OrgIndex := indexOrgs(OrgList)
	UserSubmittedTixIndex := indexUserSubmittedTickets(TicketList)
	UserAssignedTixIndex := indexUserAssignedTickets(TicketList)

	searchField := "ID"
	searchValue := "36"
	users, err := SearchUsers(searchField, searchValue, UserList)
	if err != nil {
		t.Error("TestgetAssociatedOrgsAndTickets: error with user search.\n")
	}

	usersAugmented := getAssociatedOrgsAndTickets(users, OrgIndex, UserSubmittedTixIndex, UserAssignedTixIndex)

	if len(usersAugmented) <= 0 {
		t.Error("TestgetAssociatedOrgsAndTickets: error with augmented user search results.\n")
	}

	userOrg := usersAugmented[0].OrgObject
	userTixSubmitted := usersAugmented[0].TicketsSubmitted
	userTixAssigned := usersAugmented[0].TicketsAssigned

	if userOrg.ID != 115 || len(userTixSubmitted) != 3 || len(userTixAssigned) != 3 {
		t.Error("TestgetAssociatedOrgsAndTickets: incorrect augmented user search results.\n")
	}
}
