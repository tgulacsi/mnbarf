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
	"text/template"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
	"github.com/tgulacsi/go/loghlp/kitloghlp"
	"github.com/tgulacsi/mnbarf/mnb"
)

//go:generate gowsdl -p mnb -o generated_arfolyamok.go "http://www.mnb.hu/arfolyamok.asmx?WSDL"
//go:generate gowsdl -p mnb -o generated_alapkamat.go "http://www.mnb.hu/alapkamat.asmx?WSDL"

var logger = kitloghlp.New(os.Stderr)

func main() {
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

Get the base rates:
	mnbarf rates|baserate|kamat|alapkamat

-format awaits
	csv for semicolon separated output in the column order of
		day;currency;unit;rate
	json to output JSON with Day, Currency, Unit and Rate fields
	or anything else, which will be treated as a Go text/template,
		with fields of Day, Currency, Unit and Rate.

Generate (and build) new webservice client
(you will need an installed Go and have gowsdl installed
 (go get github.com/hooklift/gowsdl)):
    go generate && go install

Possible options:
`)
		flag.PrintDefaults()
	}
	flag.Parse()
	hndl := log15.StderrHandler
	if !*flagVerbose {
		hndl = log15.LvlFilterHandler(log15.LvlInfo, log15.StderrHandler)
	}
	mnb.Log = logger.With("lib", "mnb").Log

	todo := flag.Arg(0)
	if todo == "" {
		todo = "current"
	}

	wsC := mnb.NewMNBArfolyamService()
	wsR := mnb.NewMNBAlapkamatService()

	switch todo {
	case "alapkamat", "kamat", "rate", "baserate":
		if flag.NArg() > 1 {
			begin, end, err := parseDates(flag.Arg(1), flag.Arg(2))
			if err != nil {
				Log("msg", "parse dates", "error", err)
				os.Exit(3)
			}
			rates, err := wsR.GetBaseRates(begin, end)
			if err != nil {
				Log("msg", "GetCentralBankBaseRates", "begin", begin, "end", end, "error", err)
				os.Exit(2)
			}
			//Log("msg","GetCentralBankBaseRates", "begin", begin, "end", end, "rates", rates)
			printBaseRates(rates, *flagOutFormat)
			return
		}
		rate, err := wsR.GetCurrentBaseRate()
		if err != nil {
			Log("msg", "GetCurrentCentralBankBaseRate", "error", err)
			os.Exit(2)
		}
		//Log("msg","GetCurrentCentralBankBaseRate", "rate", rate)
		fmt.Println(rate.Publication, rate.Rate)
		return

	case "currencies", "currency", "curr":
		currencies, err := wsC.GetCurrencies()
		if err != nil {
			Log("msg", "GetCurrencies", "error", err)
			os.Exit(2)
		}
		//Log("msg","GetCurrencies", "currencies", currencies)
		for _, curr := range currencies {
			fmt.Println(curr)
		}
		return

	case "range":
		curr := flag.Arg(1)
		if curr == "" {
			Log("msg", "currency is needed")
			os.Exit(5)
		}
		begin, end, err := parseDates(flag.Arg(2), flag.Arg(3))
		if err != nil {
			Log("msg", "parse dates", "error", err)
			os.Exit(3)
		}
		dayRates, err := wsC.GetExchangeRates(curr, begin, end)
		if err != nil {
			Log("msg", "GetExchangeRates", "error", err)
		}
		//Log("msg","GetExchangeRates", "dayRates", dayRates)
		printDayRates(dayRates, *flagOutFormat)
		return
	}

	// current
	day, err := wsC.GetCurrentExchangeRates()
	if err != nil {
		Log("msg", "GetCurrentExchangeRates", "error", err)
	}
	//Log("msg","GetCurrentExchangeRates", "day", day.Day, "rates", day.Rates)
	printDayRates([]mnb.DayRates{day}, *flagOutFormat)
}

func parseDates(beginS, endS string) (begin, end time.Time, err error) {
	if beginS == "" {
		begin = time.Now().AddDate(0, 0, -30)
		end = time.Now().AddDate(0, 0, -1)
		return
	}
	begin, err = time.Parse("2006-01-02", beginS)
	if err != nil {
		err = errors.Wrapf(err, "arg="+beginS)
		return
	}
	if endS == "" {
		end = time.Now().AddDate(0, 0, -1)
		return
	}
	end, err = time.Parse("2006-01-02", endS)
	if err != nil {
		err = errors.Wrapf(err, "arg="+endS)
	}
	return
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
					Log("msg", "encoding", "row", row, "error", err)
					return err
				}
			}
		}
		bw.WriteString("]")

	default: // template
		tmpl, err := template.New("row").Parse(outFormat)
		if err != nil {
			Log("msg", "template parse", "error", err)
			os.Exit(4)
		}
		row := rowStruct{}
		bw.WriteString("[")
		for _, day := range days {
			row.Day = day.Day.String()
			for _, rate := range day.Rates {
				row.Currency, row.Unit, row.Rate = rate.Currency, rate.Unit, rate.Rate.String()
				if err := tmpl.Execute(bw, row); err != nil {
					Log("msg", "encoding", "row", row, "error", err)
					return err
				}
			}
		}
	}
	return nil
}

func printBaseRates(rates []mnb.MNBBaseRate, outFormat string) error {
	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	switch outFormat {
	case "csv":
		for _, rate := range rates {
			fmt.Fprintf(bw, "%s;%s\n", rate.Publication, rate.Rate)
		}

	case "json":
		enc := json.NewEncoder(bw)

		bw.WriteString("[")
		for _, rate := range rates {
			if err := enc.Encode(rate); err != nil {
				Log("msg", "encoding", "rate", rate, "error", err)
				return err
			}
		}
		bw.WriteString("]")

	default: // template
		tmpl, err := template.New("row").Parse(outFormat)
		if err != nil {
			Log("msg", "template parse", "error", err)
			os.Exit(4)
		}
		bw.WriteString("[")
		for _, rate := range rates {
			if err := tmpl.Execute(bw, rate); err != nil {
				Log("msg", "encoding", "rate", rate, "error", err)
				return err
			}
		}
	}
	return nil
}
