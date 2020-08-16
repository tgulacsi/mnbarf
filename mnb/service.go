/*
Copyright 2017 Tamás Gulácsi

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

package mnb

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

var Log = func(...interface{}) error { return nil }

//////////////////
// ExchangeRate //
//////////////////

// MNBArfolyamService - see
// http://www.biprojekt.hu/blog/MNB_arfolyamok_letoltese-Kozvetlenul_az_MNB-tol.htm
/*
   Az MNB web Service-e felé háromfajta lekérdezést tudunk küldeni.

	   GetInfo method: A GetInfo method visszaadja az MNB devizáinak kódját 1949-től napjainkig. Tulajdonképpen egy deviza törzs, ami nem tartalmaz mást, mint a devizák hárombetűs kódját. (mint például HUF, EUR, USD, de megtalálhatóak benne korábbi devizák is, mint pl. a DEM, vagy a jugoszláv dinár)
	   GetCurrentExchangeRates Method: A metódus visszaadja az éppen aktuális árfolyamokat, azaz a lekérdezés pillanatában érvényben lévő árfolyamokat.
	   GetExchangeRates Method. a GetExchangeRates Method visszaadja a napi árfolyamokat adott dátumtól dátumig intervallumra, adott devizára. Tehát a metódusnak létezik három bemenő paramétere: StartDate, EndDate, és Currency és ezek alapján visszaadja a hivatalos árfolyamokat.

   Nekünk az adattárház feltöltéséhez a GetInfo, és a GetExchangeRates metódusokra lesz leginkább szükségünk.
*/

type MNBArfolyamService struct {
	srvc *MNBArfolyamServiceSoap
}

func NewMNBArfolyamService(urls ...string) MNBArfolyamService {
	if len(urls) == 0 {
		urls = append(urls, "")
	}
	return MNBArfolyamService{srvc: NewMNBArfolyamServiceSoap(urls[0], strings.HasPrefix(urls[0], "https://"))}
}

type MNBExchangeRatesQueryValues struct {
	FirstDate  Date
	LastDate   Date
	Currencies []string `xml:"Currencies>Curr"`
}

// GetCurrencies returns the list of currencies (3 letter codes)
func (ws MNBArfolyamService) GetCurrencies() ([]string, error) {
	t := time.Now()
	resp, err := ws.srvc.GetCurrencies(&GetCurrencies{})
	if err != nil {
		return nil, err
	}
	dur := time.Since(t)
	Log("msg", "GetCurrencies", "duration", dur)
	//Log("msg","GetCurrencies", "resp", resp)

	var qv MNBExchangeRatesQueryValues
	err = xml.Unmarshal([]byte(resp.GetCurrenciesResult), &qv)
	if err != nil {
		return nil, err
	}
	return qv.Currencies, nil
}

// GetCurrentExchangeRates returns the actual exchange rates.
func (ws MNBArfolyamService) GetCurrentExchangeRates() (DayRates, error) {
	t := time.Now()
	resp, err := ws.srvc.GetCurrentExchangeRates(&GetCurrentExchangeRates{})
	if err != nil {
		return DayRates{}, err
	}
	dur := time.Since(t)
	Log("msg", "GetCurrentExchangeRates", "duration", dur)
	//Log("msg","GetCurrentExchangeRates", "resp", resp)
	var rates MNBCurrentExchangeRates
	err = xml.Unmarshal([]byte(resp.GetCurrentExchangeRatesResult), &rates)
	return rates.Day, err
}

type MNBExchangeRates struct {
	Days []DayRates `xml:"Day"`
}

// GetExchangeRates returns the exchange rates between the specified dates.
func (ws MNBArfolyamService) GetExchangeRates(currencyNames string, begin, end time.Time) ([]DayRates, error) {
	t := time.Now()
	resp, err := ws.srvc.GetExchangeRates(&GetExchangeRates{
		StartDate:     begin.Format("2006-01-02"),
		EndDate:       end.Format("2006-01-02"),
		CurrencyNames: currencyNames,
	})
	if err != nil {
		return nil, err
	}
	dur := time.Since(t)
	Log("msg", "GetExchangeRates", "duration", dur)
	//Log("msg","GetExchangeRates", "resp", resp)
	var rates MNBExchangeRates
	err = xml.Unmarshal([]byte(resp.GetExchangeRatesResult), &rates)
	return rates.Days, err
}

//////////////
// BaseRate //
//////////////

type MNBAlapkamatService struct {
	srvc *MNBAlapkamatServiceSoap
}

func NewMNBAlapkamatService(urls ...string) MNBAlapkamatService {
	if len(urls) == 0 {
		urls = append(urls, "")
	}
	return MNBAlapkamatService{srvc: NewMNBAlapkamatServiceSoap(urls[0], strings.HasPrefix(urls[0], "https://"))}
}

// GetCurrentBaseRate returns the current base rate.
func (ws MNBAlapkamatService) GetCurrentBaseRate() (MNBBaseRate, error) {
	t := time.Now()
	resp, err := ws.srvc.GetCurrentCentralBankBaseRate(&GetCurrentCentralBankBaseRate{})
	if err != nil {
		return MNBBaseRate{}, err
	}
	dur := time.Since(t)
	Log("msg", "GetCurrentCentralBankBaseRate", "duration", dur)
	//Log("msg","GetCurrentCentralBankBaseRate", "resp", resp)
	var rate MNBCurrentCentralBankBaseRate
	err = xml.Unmarshal([]byte(resp.GetCurrentCentralBankBaseRateResult), &rate)
	return rate.BaseRate, err
}

// <![CDATA[<MNBCentralBankBaseRate><BaseRates><BaseRate publicationDate="2015-03-25">1,9500</BaseRate><BaseRate publicationDate="2015-04-22">1,8000</BaseRate><BaseRate publicationDate="2015-05-27">1,6500</BaseRate><BaseRate publicationDate="2015-06-24">1,5000</BaseRate></BaseRates></MNBCentralBankBaseRate>]]>
type MNBCentralBankBaseRate struct {
	XMLName   xml.Name      `xml:"MNBCentralBankBaseRate" json:"-"`
	BaseRates []MNBBaseRate `xml:"BaseRates>BaseRate"`
}

// GetBaseRates returns the base rates between the specified dates.
func (ws MNBAlapkamatService) GetBaseRates(begin, end time.Time) ([]MNBBaseRate, error) {
	t := time.Now()
	resp, err := ws.srvc.GetCentralBankBaseRate(&GetCentralBankBaseRate{
		StartDate: begin.Format("2006-01-02"),
		EndDate:   end.Format("2006-01-02"),
	})
	if err != nil {
		return nil, err
	}
	dur := time.Since(t)
	Log("msg", "GetCentralBankBaseRate", "duration", dur)
	//Log("msg","GetCentralBankBaseRate", "resp", resp)
	var rates MNBCentralBankBaseRates
	if err = xml.Unmarshal([]byte(resp.GetCentralBankBaseRateResult), &rates); err != nil {
		return nil, fmt.Errorf("%v\n%s", err, resp.GetCentralBankBaseRateResult)
	}
	return rates.BaseRates, nil
}
