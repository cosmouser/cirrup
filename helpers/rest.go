package helpers

import (
	"../data"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RestAuth is a subset of the toml config for passing auth into
// the deleltion and addition helpers
type RestAuth struct {
	JssUrl  string
	JssPort int
	ApiUser string
	ApiPass string
}

// generateAddition creates an xml body for a list of computers by jssID
func generateAddition(computers []int) (string, error) {
	if len(computers) < 1 {
		return "", errors.New("Cannot generate addition for 0 computers")
	}
	xml := `<computer_group><computer_additions>`
	for _, j := range computers {
		xml += `<computer><id>`
		xml += fmt.Sprintf("%d", j)
		xml += `</id></computer>`
	}
	xml += `</computer_additions></computer_group>`
	return xml, nil
}

// generateDeletion creates an xml body for a list of computers by jssID
func generateDeletion(computers []int) (string, error) {
	if len(computers) < 1 {
		return "", errors.New("Cannot generate deletion for 0 computers")
	}
	xml := `<computer_group><computer_deletions>`
	for _, j := range computers {
		xml += `<computer><id>`
		xml += fmt.Sprintf("%d", j)
		xml += `</id></computer>`
	}
	xml += `</computer_deletions></computer_group>`
	return xml, nil
}

// SendDeletion sends an http PUT to remove computers from a group in
// the JSS
func SendDeletion(computers []int, fsg_id int, aconfig RestAuth) error {
	xml, err := generateDeletion(computers)
	if err != nil {
		return err
	}
	xmlReader := strings.NewReader(xml)
	resourceURI := fmt.Sprintf("%s/JSSResource/computergroups/id/%d",
		fmt.Sprintf("%s:%d", aconfig.JssUrl, aconfig.JssPort),
		fsg_id,
	)
	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("PUT", resourceURI, xmlReader)
	req.SetBasicAuth(aconfig.ApiUser, aconfig.ApiPass)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 201 {
		data.Info.Printf("cirrup: removed %v from fsg %d\n",
			computers,
			fsg_id,
		)
	}
	// If the status code is 409 on a deletion, none of the
	// computers are actually removed from the group.
	// In order to remedy this, the server should set
	// whatever computers that were in the request that caused
	// the 409 to the fsg of 0.
	if resp.StatusCode == 409 {
		data.Info.Printf("cirrup: conflict encountered removing %v from fsg %d\n",
			computers,
			fsg_id,
		)
	}
	return nil
}

// SendAddition sends an http PUT to add computers to a group in
// the JSS
func SendAddition(computers []int, fsg_id int, aconfig RestAuth) error {
	xml, err := generateAddition(computers)
	if err != nil {
		return err
	}
	xmlReader := strings.NewReader(xml)
	resourceURI := fmt.Sprintf("%s/JSSResource/computergroups/id/%d",
		fmt.Sprintf("%s:%d", aconfig.JssUrl, aconfig.JssPort),
		fsg_id,
	)
	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("PUT", resourceURI, xmlReader)
	req.SetBasicAuth(aconfig.ApiUser, aconfig.ApiPass)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 201 {
		data.Info.Printf("cirrup: added %v to fsg %d\n",
			computers,
			fsg_id,
		)
	}
	return nil
}
