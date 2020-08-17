/*
Copyright 2020 Tamás Gulácsi

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
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pkg/errors"
	"github.com/tgulacsi/mnbarf/mnb"
)

//go:generate gowsdl -p mnb -o generated_arfolyamok.go "http://www.mnb.hu/arfolyamok.asmx?WSDL"
//go:generate gowsdl -p mnb -o generated_alapkamat.go "http://www.mnb.hu/alapkamat.asmx?WSDL"

var logger = log.NewLogfmtLogger(os.Stderr)
var Log = logger.Log

func main() {
	if err := Main(); err != nil {
		logger.Log("ERROR", fmt.Sprintf("%+v", err))
	}
}

func Main() error {
	var wsC mnb.MNBArfolyamService
	var wsR mnb.MNBAlapkamatService
	fs := flag.NewFlagSet("mnbarf", flag.ContinueOnError)
	flagOutFormat := fs.String("format", "csv", `output format (possible: csv, json or template (go template: you can use Day, Currency, Unit and Rate - i.e. {{.Day}};{{.Currency}};{{.Unit}};{{.Rate}}{{print "\n"}})`)
	flagVerbose := fs.Bool("v", false, "verbose logging")
	flagURL := fs.String("url", "", "URL to use")

	baserateCmd := ffcli.Command{
		Name: "baserate",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				if len(args) == 1 {
					args = append(args, "")
				}
				begin, end, err := parseDates(args[0], args[1])
				if err != nil {
					Log("msg", "parse dates", "error", err)
					return err
				}
				rates, err := wsR.GetBaseRates(ctx, begin, end)
				if err != nil {
					Log("msg", "GetCentralBankBaseRates", "begin", begin, "end", end, "error", err)
					return err
				}
				//Log("msg","GetCentralBankBaseRates", "begin", begin, "end", end, "rates", rates)
				printBaseRates(rates, *flagOutFormat)
				return nil
			}
			rate, err := wsR.GetCurrentBaseRate(ctx)
			if err != nil {
				Log("msg", "GetCurrentCentralBankBaseRate", "error", err)
				return err
			}
			//Log("msg","GetCurrentCentralBankBaseRate", "rate", rate)
			fmt.Println(rate.Publication, rate.Rate)
			return nil
		},
	}

	currenciesCmd := ffcli.Command{
		Name: "currencies",
		Exec: func(ctx context.Context, args []string) error {
			currencies, err := wsC.GetCurrencies(ctx)
			if err != nil {
				Log("msg", "GetCurrencies", "error", err)
				return err
			}
			//Log("msg","GetCurrencies", "currencies", currencies)
			for _, curr := range currencies {
				fmt.Println(curr)
			}
			return nil
		},
	}

	ratesCmd := ffcli.Command{
		Name: "rates",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("begin, end and at least one currency is needed")
			}
			begin, end, err := parseDates(args[0], args[1])
			if err != nil {
				return err
			}
			dayRates, err := wsC.GetExchangeRates(ctx, begin, end, args[2:]...)
			if err != nil {
				Log("msg", "GetExchangeRates", "error", err)
			}
			//Log("msg","GetExchangeRates", "dayRates", dayRates)
			printDayRates(dayRates, *flagOutFormat)
			return err
		},
	}
	currentCmd := ffcli.Command{
		Name: "current",
		Exec: func(ctx context.Context, args []string) error {
			// current
			day, err := wsC.GetCurrentExchangeRates(ctx)
			if err != nil {
				Log("msg", "GetCurrentExchangeRates", "error", err)
			}
			//Log("msg","GetCurrentExchangeRates", "day", day.Day, "rates", day.Rates)
			printDayRates([]mnb.DayRates{day}, *flagOutFormat)
			return err
		},
	}
	infoCmd := ffcli.Command{
		Name: "info",
		Exec: func(ctx context.Context, args []string) error {
			info, err := wsC.GetInfo(ctx)
			if err != nil {
				Log("msg", "GetInfo", "error", err)
				return err
			}
			fmt.Printf("%s - %s:\n", info.FirstDate, info.LastDate)
			for _, curr := range info.Currencies {
				fmt.Printf("\t%s\n", curr)
			}
			return nil
		},
	}

	app := ffcli.Command{FlagSet: fs,
		LongHelp: `Usage: mnbarf [options] <command>

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

-url http://www.mnb.hu/arfolyamok.asmx

Generate (and build) new webservice client
(you will need an installed Go and have gowsdl installed
 (go get github.com/hooklift/gowsdl)):
    go generate && go install

`,
		Subcommands: append(append(append(append(make([]*ffcli.Command, 0, 16),
			&currentCmd, &infoCmd),
			alias(&baserateCmd, "alapkamat", "kamat", "rate")...),
			alias(&currenciesCmd, "currency", "curr")...),
			alias(&ratesCmd, "rates")...),

		Exec: func(ctx context.Context, args []string) error {
			return currentCmd.Exec(ctx, args)
		},
	}
	if err := app.Parse(os.Args[1:]); err != nil {
		return err
	}
	var mnbLog func(...interface{}) error
	if *flagVerbose {
		mnbLog = log.With(logger, "lib", "mnb").Log
	}

	wsC = mnb.NewMNBArfolyamService(*flagURL, nil, mnbLog)
	wsR = mnb.NewMNBAlapkamatService(*flagURL, nil, mnbLog)

	ctx, cancel := wrap(context.Background())
	defer cancel()
	return app.Run(ctx)
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

func alias(cmd *ffcli.Command, names ...string) []*ffcli.Command {
	commands := make([]*ffcli.Command, 1+len(names))
	commands[0] = cmd
	for i, nm := range names {
		cmd2 := *cmd
		cmd2.Name = nm
		commands[i+1] = &cmd2
	}
	return commands
}

// wrap returns a new context with cancel that is canceled on interrupts.
func wrap(ctx context.Context) (context.Context, context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		<-sigCh
		cancel()
		signal.Stop(sigCh)
	}()
	return ctx, cancel
}
