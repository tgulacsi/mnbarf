/*
Copyright 2014 Tamás Gulácsi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"text/template"
	"time"

	gowsdl "github.com/cloudescape/gowsdl/generator"
	"github.com/tgulacsi/mnbarf/mnb"
	"gopkg.in/inconshreveable/log15.v2"
)

var Log = log15.New()

func main() {
	Log.SetHandler(log15.StderrHandler)

	flagWSDL := flag.String("wsdl", "http://www.mnb.hu/arfolyamok.asmx?WSDL", "MNB WSDL endpoint")
	flagGowsdl := flag.String("gowsdl", "gowsdl", "path of gowsdl (only needed for generation)")
	flagOutFormat := flag.String("format", "csv", `output format (possible: csv, json or template (go template: you can use Day, Currency, Unit and Rate - i.e. {{.Day}};{{.Currency}};{{.Unit}};{{.Rate}}{{print "\n"}})`)
	flagVerbose := flag.Bool("v", false, "verbose logging")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: mnbarf [options] <command>

List all the possible currencies:
	mnbarf currencies|currency|curr

Get the exchange rates for a specified period, for the specified currencies:
	mnbarf [options: -format] range <currencies> [<first day> [<last day>]]
The default first day is 30 days before yesterday,
and the default last day is yesterday.

for example to get USD and EUR for all days in the history (till yesterday):
	mnbarf -format=csv USD,EUR 1949-01-01

Get the actual exchange rates, for all currencies - this is the defallt:
	mnbarf [options: -format]

-format awaits
	csv for semicolon separated output in the column order of
		day;currency;unit;rate
	json to output JSON with Day, Currency, Unit and Rate fields
	or anything else, which will be treated as a Go text/template,
		with fields of Day, Currency, Unit and Rate.

Generate (and build) new webservice client
(you will need an installed Go and have gowsdl installed
 (go get github.com/cloudescape/gowsdl)):
	mnbarf [options: -wsdl, -gowsdl] gen|generate

Possible options:
`)
		flag.PrintDefaults()
	}
	flag.Parse()
	hndl := log15.StderrHandler
	if !*flagVerbose {
		hndl = log15.LvlFilterHandler(log15.LvlInfo, log15.StderrHandler)
	}
	gowsdl.Log.SetHandler(hndl)
	mnb.Log.SetHandler(hndl)
	Log.SetHandler(hndl)

	todo := flag.Arg(0)
	if todo == "" {
		todo = "current"
	}
	switch todo {
	case "gen", "generate":
		if err := exec.Command(*flagGowsdl, "-p", "mnb", "-o", "generated.go", *flagWSDL).Run(); err != nil {
			Log.Crit("generating mng/generated.go", "error", err)
			os.Exit(1)
		}
		if err := exec.Command("go", "build").Run(); err != nil {
			Log.Crit("building", "error", err)
			os.Exit(2)
		}
		os.Exit(0)
		return
	}

	ws := mnb.NewMNBArfolyamService()

	switch todo {
	case "currencies", "currency", "curr":
		currencies, err := ws.GetCurrencies()
		if err != nil {
			Log.Error("GetCurrencies", "error", err)
			os.Exit(2)
		}
		Log.Debug("GetCurrencies", "currencies", currencies)
		for _, curr := range currencies {
			fmt.Println(curr)
		}
		return

	case "range":
		curr := flag.Arg(1)
		if curr == "" {
			Log.Error("currency is needed")
			os.Exit(5)
		}
		var begin, end time.Time
		var err error
		s := flag.Arg(2)
		if s == "" {
			begin = time.Now().AddDate(0, 0, -30)
			end = time.Now().AddDate(0, 0, -1)
		} else {
			begin, err = time.Parse("2006-01-02", s)
			if err != nil {
				Log.Error("cannot parse first arg as 2006-01-02", "error", err)
				os.Exit(3)
			}
			s = flag.Arg(3)
			if s == "" {
				end = time.Now().AddDate(0, 0, -1)
			} else {
				end, err = time.Parse("2006-01-02", flag.Arg(2))
				if err != nil {
					Log.Error("cannot parse second arg as 2006-01-02", "error", err)
					os.Exit(3)
				}
			}
		}
		dayRates, err := ws.GetExchangeRates(curr, begin, end)
		if err != nil {
			Log.Error("GetExchangeRates", "error", err)
		}
		Log.Debug("GetExchangeRates", "dayRates", dayRates)
		printDayRates(dayRates, *flagOutFormat)
		return
	}

	// current
	day, err := ws.GetCurrentExchangeRates()
	if err != nil {
		Log.Error("GetCurrentExchangeRates", "error", err)
	}
	Log.Debug("GetCurrentExchangeRates", "day", day.Day, "rates", day.Rates)
	printDayRates([]mnb.DayRates{day}, *flagOutFormat)
}

func printDayRates(days []mnb.DayRates, outFormat string) error {
	type rowStruct struct {
		Day      string
		Currency string
		Unit     int
		Rate     string
	}

	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	switch outFormat {
	case "csv":
		for _, day := range days {
			dS := day.Day.String()
			for _, rate := range day.Rates {
				fmt.Fprintf(bw, dS+";"+rate.Currency+";")
				fmt.Fprintf(bw, "%d;%s\n", rate.Unit, rate.Rate.String())
			}
		}

	case "json":
		enc := json.NewEncoder(bw)

		row := rowStruct{}
		bw.WriteString("[")
		for _, day := range days {
			row.Day = day.Day.String()
			for _, rate := range day.Rates {
				row.Currency, row.Unit, row.Rate = rate.Currency, rate.Unit, rate.Rate.String()
				if err := enc.Encode(row); err != nil {
					Log.Error("encoding", "row", row, "error", err)
					return err
				}
			}
		}
		bw.WriteString("]")

	default: // template
		tmpl, err := template.New("row").Parse(outFormat)
		if err != nil {
			Log.Crit("template parse", "error", err)
			os.Exit(4)
		}
		row := rowStruct{}
		bw.WriteString("[")
		for _, day := range days {
			row.Day = day.Day.String()
			for _, rate := range day.Rates {
				row.Currency, row.Unit, row.Rate = rate.Currency, rate.Unit, rate.Rate.String()
				if err := tmpl.Execute(bw, row); err != nil {
					Log.Error("encoding", "row", row, "error", err)
					return err
				}
			}
		}
	}
	return nil
}
