package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

import (
	"github.com/codegangsta/cli"
)

import (
	"github.com/zpeters/speedtest/debug"
	"github.com/zpeters/speedtest/sthttp"
	"github.com/zpeters/speedtest/print"
	"github.com/zpeters/speedtest/tests"
	"github.com/zpeters/speedtest/settings"
)

// VERSION is the version of our software
var VERSION = "v0.8.1-1-gd861298"


func runTest(c *cli.Context) {
	// create our server object and load initial config
	var testServer sthttp.Server
	sthttp.CONFIG = sthttp.GetConfig()

	if debug.DEBUG {
		print.EnvironmentReport(c)
	}

	// get all possible servers
	if debug.DEBUG {
		log.Printf("Getting all servers for our test list")
	}
	allServers := sthttp.GetServers()

	// if they specified a specific server, test against that...
	if c.String("server") != "" {
		if debug.DEBUG {
			log.Printf("Server '%s' specified, getting info...", c.String("server"))
		}
		// find server and load latency report
		testServer = tests.FindServer(c.String("server"), allServers)
		// load latency
		testServer.Latency = sthttp.GetLatency(testServer)

		fmt.Printf("Selected server: %s\n", testServer)
		// ...otherwise get a list of all servers sorted by distance...
	} else {
		if debug.DEBUG {
			log.Printf("Getting closest servers...")
		}
		closestServers := sthttp.GetClosestServers(allServers)
		if debug.DEBUG {
			log.Printf("Getting the fastests of our closest servers...")
		}
		// ... and get the fastests NUMCLOSEST ones
		testServer = sthttp.GetFastestServer(closestServers)
	}

	// Start printing our report
	if !debug.REPORT {
		print.Server(testServer)
	} else {
		print.ServerReport(testServer)
	}

	// if ping only then just output latency results and exit nicely...
	if c.Bool("ping") {
		if c.Bool("report") {
			if settings.ALGOTYPE == "max" {
				fmt.Printf("%3.2f (Lowest)\n", testServer.Latency)
			} else {
				fmt.Printf("%3.2f (Avg)\n", testServer.Latency)
			}
		} else {
			if settings.ALGOTYPE == "max" {
				fmt.Printf("Ping (Lowest): %3.2f ms\n", testServer.Latency)
			} else {
				fmt.Printf("Ping (Avg): %3.2f ms\n", testServer.Latency)
			}
		}
		os.Exit(0)
		// ...otherwise run our full test
	} else {
		var dmbps float64
		var umbps float64
		
		if c.Bool("downloadonly") {
			dmbps = tests.DownloadTest(testServer)
		} else if c.Bool("uploadonly") {
			umbps = tests.UploadTest(testServer)
		} else {
			dmbps = tests.DownloadTest(testServer)
			umbps = tests.UploadTest(testServer)
		}

		if !debug.REPORT {
			if settings.ALGOTYPE == "max" {
				fmt.Printf("Ping (Lowest): %3.2f ms | Download (Max): %3.2f Mbps | Upload (Max): %3.2f Mbps\n", testServer.Latency, dmbps, umbps)
			} else {
				fmt.Printf("Ping (Avg): %3.2f ms | Download (Avg): %3.2f Mbps | Upload (Avg): %3.2f Mbps\n", testServer.Latency, dmbps, umbps)
			}
		} else {
			dkbps := dmbps * 1000
			ukbps := umbps * 1000
			fmt.Printf("%3.2f%s%d%s%d\n", testServer.Latency, settings.REPORTCHAR, int(dkbps), settings.REPORTCHAR, int(ukbps))
		}
	}

}

func main() {
	// seeding randomness
	rand.Seed(time.Now().UTC().UnixNano())

	// set logging to stdout for global logger
	log.SetOutput(os.Stdout)

	// setting up cli settings
	app := cli.NewApp()
	app.Name = "speedtest"
	app.Usage = "Unofficial command line interface to speedtest.net (https://github.com/zpeters/speedtest)"
	app.Author = "Zach Peters - zpeters@gmail.com - github.com/zpeters"
	app.Version = VERSION

	// setup cli flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "algo, a",
			Usage: "Specify the measurement method to use ('max', 'avg')",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "Turn on debugging",
		},
		cli.BoolFlag{
			Name:  "list, l",
			Usage: "List available servers",
		},
		cli.BoolFlag{
			Name:  "ping, p",
			Usage: "Ping only mode",
		},
		cli.BoolFlag{
			Name:  "quiet, q",
			Usage: "Quiet mode",
		},
		cli.BoolFlag{
			Name:  "report, r",
			Usage: "Reporting mode output, minimal output with '|' for separators, use '--rc' to change separator characters. Reports the following: Server ID, Server Name (Location), Ping time in ms, Download speed in kbps, Upload speed in kbps",
		},
		cli.BoolFlag{
			Name: "downloadonly, do",
			Usage: "Only perform download test",
		},
		cli.BoolFlag{
			Name: "uploadonly, uo",
			Usage: "Only perform upload test",
		},
		cli.StringFlag{
			Name:  "reportchar, rc",
			Usage: "Set the report separator. Example: --rc=','",
		},
		cli.StringFlag{
			Name:  "server, s",
			Usage: "Use a specific server",
		},
		cli.IntFlag{
			Name:  "numclosest, nc",
			Value: settings.NUMCLOSEST,
			Usage: "Number of 'closest' servers to find",
		},
		cli.IntFlag{
			Name:  "numlatency, nl",
			Value: settings.NUMLATENCYTESTS,
			Usage: "Number of latency tests to perform",
		},
	}

	// toggle our switches and setup variables
	app.Action = func(c *cli.Context) {
		// set our flags
		if c.Bool("debug") {
			debug.DEBUG = true
		}
		if c.Bool("quiet") {
			debug.QUIET = true
		}
		if c.Bool("report") {
			debug.REPORT = true
		}
		if c.String("algo") != "" {
			if c.String("algo") == "max" {
				settings.ALGOTYPE = "max"
			} else if c.String("algo") == "avg" {
				settings.ALGOTYPE = "avg"
			} else {
				fmt.Printf("** Invalid algorithm '%s'\n", c.String("algo"))
				os.Exit(1)
			}
		}
		settings.NUMCLOSEST = c.Int("numclosest")
		settings.NUMLATENCYTESTS = c.Int("numlatency")
		if c.String("reportchar") != "" {
			settings.REPORTCHAR = c.String("reportchar")
		}

		// run a oneshot list
		if c.Bool("list") {
			tests.ListServers()
			os.Exit(0)
		}

		// run our test
		runTest(c)
	}
	// run the app
	app.Run(os.Args)
}
