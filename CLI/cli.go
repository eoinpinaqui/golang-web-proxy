package CLI

import (
	proxy "../Proxy"
	"bufio"
	"fmt"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	BLOCK     = "block"
	UNBLOCK   = "unblock"
	TIMING    = "timing"
	BANDWIDTH = "bandwidth"
	LIST      = "list"
	BLOCKED   = "blocked"
	CACHED    = "cached"
)

// This function handles input from the user in the management console
func Cli(p *proxy.Proxy) {
	reader := bufio.NewReader(os.Stdin)
	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			log.Error(err.Error())
		} else {
			// Extract the command and execute it
			c := strings.Fields(command)
			if len(c) > 0 {
				switch strings.ToLower(c[0]) {
				case BLOCK:
					if len(c) == 2 {
						err = p.BlockedSites.Add(c[1])
						if err != nil {
							log.Error(err)
						} else {
							log.Infof("%s has been blocked successfully\n", c[1])
						}
					} else {
						log.Errorf("Invalid block command\n")
					}
				case UNBLOCK:
					if len(c) == 2 {
						err = p.BlockedSites.Remove(c[1])
						if err != nil {
							log.Error(err)
						} else {
							log.Infof("%s has been unblocked successfully\n", c[1])
						}
					} else {
						log.Errorf("Invalid unblock command\n")
					}
				case LIST:
					if len(c) == 2 {
						switch c[1] {
						case BLOCKED:
							fmt.Println("\nBLOCKED SITES:")
							for _, v := range p.BlockedSites.List() {
								fmt.Println(v)
							}
						case CACHED:
							fmt.Println("\nCACHED SITES:")
							fmt.Println(p.CachedSites.List())
						case TIMING:
							data := p.PerformanceData.GetAverageTimes()
							table := tablewriter.NewWriter(os.Stdout)
							table.SetHeader([]string{"Host", "Average Uncached Response Time", "Average Cached Response Time"})

							for _, v := range data {
								table.Append(v)
							}
							table.Render()
						case BANDWIDTH:
							data := p.PerformanceData.GetAverageBandwidths()
							table := tablewriter.NewWriter(os.Stdout)
							table.SetHeader([]string{"Host", "Average Uncached Bandwidth", "Average Cached Bandwidth"})

							for _, v := range data {
								table.Append(v)
							}
							table.Render()
						default:
							log.Errorf("Invalid list specification \"%s\"\n", c[1])
						}
					} else {
						log.Errorf("Invalid list command\n")
					}
				default:
					log.Error("Unrecognised command\n")
				}
			}
		}
	}
}
