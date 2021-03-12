package main
import (
	cli "./CLI"
	proxy "./Proxy"
	log "github.com/sirupsen/logrus"
	"os"
)
func main() {
	// Configure the logs to look better
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	// Ensure appropriate command line parameters
	if len(os.Args) < 2 {
		log.Fatal("Not enough command line arguments provided\n" +
			"\tTry \"go run main.go :8080\"\n ")
	} else if len(os.Args) > 2 {
		log.Fatal("Too many command line arguments provided\n" +
			"\tTry \"go run main.go :8080\"\n ")
	}
	// Create a proxy and start listening on the given port
	log.Infof("Starting web proxy...")
	p := proxy.New(os.Args[1])
	go cli.Cli(p)
	p.Start()
}