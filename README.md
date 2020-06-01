# README

This search app allows a user to search across a given repository of organization, user, and ticket data. The user can search on any attribute of any of those entities, and the search will automatically show the primary entities relevant to the search conducted augmented with other associated entities (e.g. a search for an organization will also show associated users and tickets for each organization appearing in the search results). 


# Installation/Configuration

This app is written using the Go programming language. A working Go installation is required to build and run the search application. Instructions to download, install and configure Go for all major operating system platforms are available at https://golang.org/dl/ (for instance, Linux-specific information is available at https://golang.org/doc/install?download=go1.12.1.linux-amd64.tar.gz). 

Once installed, newer versions of Go now provide automatic access to the base go command on Windows CLI (if it doesn't, you will need to add the Go binary path to the Windows Path environment variable). If you're on Linux, ensure that the Go binary path (usually /usr/local/go/bin) is added to $PATH and the $HOME/.profile file and go command is available on command line (type 'go version' to check). 

Create a workspace directory at $HOME/go. Clone the contents of this repo into $HOME/go/src. Change to the zd_search directory on a command line. The application uses an augmented reflections package (https://gopkg.in/oleiade/reflections.v1), to help with primary entity searches to be run on arbitrary fields. Install this package by running `go get gopkg.in/oleiade/reflections.v1`. Then run `go build search.go` to build the app and `./search` to run it. Alternatively, the application can be run directly using Go's interpret mode, by running `go run search.go`


# Application Design/Implementation

Upon initial invocation by (i.e. running `./search`) the app would refer to a config file to get primary data file locations (organization/user/ticket JSON files). It would then read this data and build a number of indexes to help search efficiently across datasets. After indexing, the application would interact with the user by providing a REPL-style recurring command prompt where users can enter search queries of a pre-defined format. Search results will be printed to the terminal and the users can keep entering further search queries. The REPL can be exited using Ctrl+C.

For each primary entity type (i.e. Org/User/Ticket), the app will perform a linear search on the relevant dataset to find results. Once a primary result set is obtained, it will then call a relevant result augmentation function (e.g. if the primary search was for organizations, getAssociatedUsersAndTickets() would augment it by populating associated Users and Tickets for each Org found in the primary search). Result augmentation uses indexes built at the app initialization, and can augment results in constant time. Finally, the augmented result set is input to a formatting function to output the results to terminal in a human-readable format. 

The time complexity of the search would therefore be the `O(size of primary dataset) + O(size of the search result set)`, and is largely linear. If indexing was not utilized this would have been close to quadratic. However, this increases the space requirements of the app as indexing utilizes extra memory space. 


# Usage

Run `./search` (or `go run search`) to initialize and start up the search app on a command line. It will display a search prompt to enter queries. 

The search query is expected to be of the format `$> searchtype searchfield searchvalues`

SearchType has to be one of the following literals: `org`, `user`, `ticket`. It denotes if the search is for organizations, users, or tickets, respectively.

SearchField denotes the attribute field the search is conducted over, and depends on the SearchType's value. The accepted values are as follows:

* if the SearchType is `org`, SearchField must be one of the following: ID, Name, URL, External_id , DomainNames, Created_at, Details, Shared_tickets, Tags
* if the SearchType is `user`, SearchField must be one of the following: ID, Name, URL, External_id, Alias, Created_at, Active, Verified, Shared, Locale, Timezone, Last_login_at, Email, Phone, Signature, Tags, Suspended, Role, Org
* if the SearchType is `ticket`, SearchField must be one of the following: ID, URL, External_id, Created_at, Priority, Status, Type, Subject, Description, Tags, Org, Has_incidents, Due_at, Submitter, Assignee, Via

SearchValues can take any number or string form, and the app will look for values exactly matching the input. It can have spaces (while SearchType or SearchField cannot), and strings must not be entered within quotes (unless the target value includes quotes). It can also be empty (i.e. only SearchType and SearchField entered in the query), and the app will search for results with the specified field being empty.


# Testing

Tests are included in the `search_test.go` file within the repository. They can be invoked by running `go test` within the repository on a command line. 


# Future Improvements

A number of further improvements could be made to the application given time permits, summarized as below:

* More granular tests
* A special command to rebuild indexes while running the search REPL-style prompt
* Improve primary search implementation to perform better than O(n)
* Support wildcard searches 


# License

Feel free to use any of this code as needed, without commercial/non-commercial limitations. It is however provided as-is, without specific warranties. 
