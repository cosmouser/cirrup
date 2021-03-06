package main

import (
	"./data"
	"./helpers"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Config struct {
	JssUrl         string
	JssPort        int
	ApiUser        string
	ApiPass        string
	LdapUrl        string
	LdapPort       int
	LdapSearchBase string
	LdapAttribute  string
	DaysPerRefresh int
	CirrupPort     int
	ValueGidMap    map[string]int
}

var config Config

type ComputerInventoryReport struct {
	Webhook struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		WebhookEvent string `json:"webhookEvent"`
	} `json:"webhook"`
	Event struct {
		Udid                string `json:"udid"`
		DeviceName          string `json:"deviceName"`
		Model               string `json:"model"`
		MacAddress          string `json:"macAddress"`
		AlternateMacAddress string `json:"alternateMacAddress"`
		SerialNumber        string `json:"serialNumber"`
		OsVersion           string `json:"osVersion"`
		OsBuild             string `json:"osBuild"`
		UserDirectoryID     string `json:"userDirectoryID"`
		Username            string `json:"username"`
		RealName            string `json:"realName"`
		EmailAddress        string `json:"emailAddress"`
		Phone               string `json:"phone"`
		Position            string `json:"position"`
		Department          string `json:"department"`
		Building            string `json:"building"`
		Room                string `json:"room"`
		JssID               int    `json:"jssID"`
	} `json:"event"`
}

type ComputerRecord struct {
	ComputerID int
	Username   string
	FsgID      int
}

// handleCirrup is the main handler, it takes the incoming webhooks
// and it determines whether the computer should be assigned
// to one of the specified functional smart groups
func handleCirrup(w http.ResponseWriter, r *http.Request) {
	var c ComputerInventoryReport
	if r.URL.Path != "/handle_cirrup" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case "GET":
		w.Write([]byte("This is the cirrup handler."))
	case "POST":
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Body.Close()
		hooksReceived.Inc()

		var userEmpty, userInCache bool
		var affiliation string
		userEmpty = c.Event.Username == ""

		log.WithFields(log.Fields{
			"jssID":    c.Event.JssID,
			"username": c.Event.Username,
		}).Info("webhook received")

		if !userEmpty {
			userInCache = data.LookupUser(c.Event.Username)
		}

		// check if the user is in the cache already
		// if they aren't, then add them to the cache
		// after getting their affiliation from ldap
		if !userInCache && !userEmpty {
			result, err := helpers.GetLdapValue(config.LdapUrl,
				config.LdapSearchBase,
				c.Event.Username,
				config.LdapAttribute,
				config.LdapPort,
			)
			if err != nil {
				log.Fatal(err)
			}
			err = data.InsertUser(c.Event.Username, result)
			if err != nil {
				log.Warn(err)
			}
		}

		// Show that the insertion was successfull
		if !userEmpty {
			userInCache = data.LookupUser(c.Event.Username)
			if !userInCache {
				log.WithFields(log.Fields{
					"jssID":    c.Event.JssID,
					"username": c.Event.Username,
				}).Warn("user was not added to cache")
				return
			}

			affiliation = data.GetUserAff(c.Event.Username)

		} else {
			affiliation = ""
		}

		// Check to see if the computer is in the cache
		// If not, put it in the cache with an fsg_id of 0
		computerInCache := data.LookupComputer(c.Event.JssID)

		if !computerInCache {
			err = data.InsertComputer(c.Event.JssID, 0, c.Event.Username)
			if err != nil {
				log.WithFields(log.Fields{
					"jssID":    c.Event.JssID,
					"username": c.Event.Username,
				}).Warn(err)
			}
		} else {
			if data.GetComputerUser(c.Event.JssID) != c.Event.Username {
				data.UpdateComputerUser(c.Event.JssID, c.Event.Username)
			}
		}

		// Show that the insertion was successfull
		computerInCache = data.LookupComputer(c.Event.JssID)
		if !computerInCache {
			log.WithFields(log.Fields{
				"jssID":    c.Event.JssID,
				"username": c.Event.Username,
			}).Warn("computer was not added to cache")
			return
		}

		// Get the computer group id number that the username maps to
		d_fsg := config.ValueGidMap[affiliation]

		// Find computers from the cache that have the same username but don't
		// have the correct computergroup
		computers, err := data.FindUnmatchedComputers(c.Event.Username, d_fsg)
		if err != nil {
			return
		}

		authconfig := helpers.RestAuth{
			JssUrl:  config.JssUrl,
			JssPort: config.JssPort,
			ApiUser: config.ApiUser,
			ApiPass: config.ApiPass,
		}
		cids := []int{}
		for _, j := range computers {
			cids = append(cids, j.ComputerID)
		}

		// group unmatched computers by fsg for removal from cached group
		unmatched := make(map[int][]int)
		for _, j := range computers {
			if j.FsgID != 0 {
				unmatched[j.FsgID] = append(unmatched[j.FsgID], j.ComputerID)
			}
		}
		if len(unmatched) > 0 {
			for i, j := range unmatched {
				err = helpers.SendDeletion(j, i, authconfig)
				if err != nil {
					for _, cid := range j {
						data.UpdateComputer(cid, 0, c.Event.Username)
					}
				}
				rpcsMade.With(prometheus.Labels{"method": "delete"}).Inc()
			}
		}
		err = helpers.SendAddition(cids, d_fsg, authconfig)
		if err != nil {
			log.Warn(err)
			return
		}
		for _, cid := range cids {
			data.UpdateComputer(cid, d_fsg, c.Event.Username)
		}
		rpcsMade.With(prometheus.Labels{"method": "add"}).Inc()

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented) + "\n"))
	}
}

var (
	hooksReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cirrup_hooks_received_total",
			Help: "Total number of hooks received from the JSS.",
		},
	)
	rpcsMade = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cirrup_rpcs_made_total",
			Help: "Total number of rpcs made to the JSS.",
		},
		[]string{"method"},
	)
	dbSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cirrup_db_size_bytes",
		Help: "Current size of the Cirrup db in bytes",
	})
)

func init() {
	// Register the counters and gauges with Prometheus's default registry.
	prometheus.MustRegister(hooksReceived)
	prometheus.MustRegister(rpcsMade)
	prometheus.MustRegister(dbSize)
}

func main() {
	var err error
	_, err = toml.DecodeFile(data.ConfigPath, &config)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			data.CullUsers(config.DaysPerRefresh)
			time.Sleep(time.Hour * 24)
		}
	}()

	go func() {
		for {
			dbSize.Set(data.GetDBSize())
			time.Sleep(time.Second * 60)
		}
	}()
	log.WithFields(log.Fields{
		"port": config.CirrupPort,
	}).Info("The cirrup has been poured")
	http.HandleFunc("/handle_cirrup", handleCirrup)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", config.CirrupPort), nil)
}
