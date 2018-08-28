package helpers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v2"
)

func GetLdapValue(url, base, uid, attr string, port int) (string, error) {
	attrValue := ""
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", url, port))
	if err != nil {
		log.Warn(err)
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		base,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(uid=%s))", uid),
		[]string{attr},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Warn(err)
	}
	for _, entry := range sr.Entries {
		attrValue = entry.GetAttributeValue(attr)
	}
	return attrValue, err
}
