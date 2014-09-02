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

	gowsdl "github.com/tgulacsi/gowsdl/generator"
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
	case "currencies", "currency":
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
		begin, err := time.Parse("2006-01-02", flag.Arg(1))
		if err != nil {
			Log.Error("cannot parse first arg as 2006-01-02", "error", err)
			os.Exit(3)
		}
		end, err := time.Parse("2006-01-02", flag.Arg(2))
		if err != nil {
			Log.Error("cannot parse second arg as 2006-01-02", "error", err)
			os.Exit(3)
		}
		curr := flag.Arg(3)
		if curr == "" {
			curr = "USD"
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
		tmpl := template.Must(template.New("row").Parse(outFormat))
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
