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

package mnb

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
	"speter.net/go/exp/math/dec/inf"
)

var Log = log15.New()

func init() {
	Log.SetHandler(log15.DiscardHandler())
}

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

func NewMNBArfolyamService() MNBArfolyamService {
	return MNBArfolyamService{srvc: NewMNBArfolyamServiceSoap("", false)}
}

type Date time.Time

func (d Date) String() string {
	return time.Time(d).String()
}

func (d Date) MarshalText() ([]byte, error) {
	return []byte(time.Time(d).Format("2006-01-02")), nil
}

func (d *Date) UnmarshalText(data []byte) error {
	t, err := time.Parse("2006-01-02", string(data))
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}

type MNBExchangeRatesQueryValues struct {
	FirstDate  Date
	LastDate   Date
	Currencies []string `xml:"Currencies>Curr"`
}

// GetCurrencies returns the list of currencies (3 letter codes)
func (ws MNBArfolyamService) GetCurrencies() ([]string, error) {
	resp, err := ws.srvc.GetInfo(&GetInfo{})
	if err != nil {
		return nil, err
	}
	Log.Debug("GetInfo", "resp", resp)

	var qv MNBExchangeRatesQueryValues
	err = xml.Unmarshal([]byte(resp.GetInfoResult), &qv)
	if err != nil {
		return nil, err
	}
	return qv.Currencies, nil
}

type MNBCurrentExchangeRates struct {
	Day DayRates
}

type DayRates struct {
	Day   *Date  `xml:"date,attr"`
	Rates []Rate `xml:"Rate"`
}

type Rate struct {
	Currency string `xml:"curr,attr"`
	Unit     int    `xml:"unit,attr"`
	Rate     Double `xml:",chardata"`
}

type Double inf.Dec

func (d *Double) String() string {
	return (*inf.Dec)(d).String()
}

func (d *Double) MarshalText() ([]byte, error) {
	return (*inf.Dec)(d).MarshalText()
}

func (d *Double) UnmarshalText(data []byte) error {
	i := bytes.IndexByte(data, ',')
	if i >= 0 {
		data[i] = '.'
	}
	if _, ok := (*inf.Dec)(d).SetString(string(data)); !ok {
		return fmt.Errorf("error parsing %q", data)
	}
	return nil
}

// GetCurrentExchangeRates returns the actual exchange rates.
func (ws MNBArfolyamService) GetCurrentExchangeRates() (DayRates, error) {
	resp, err := ws.srvc.GetCurrentExchangeRates(&GetCurrentExchangeRates{})
	if err != nil {
		return DayRates{}, err
	}
	Log.Debug("GetCurrentExchangeRates", "resp", resp)
	var rates MNBCurrentExchangeRates
	err = xml.Unmarshal([]byte(resp.GetCurrentExchangeRatesResult), &rates)
	return rates.Day, nil
}

// GetExchangeRates returns the exchange rates between the specified dates.
func (ws MNBArfolyamService) GetExchangeRates(begin, end time.Time) error {
	resp, err := ws.srvc.GetExchangeRates(&GetExchangeRates{})
	if err != nil {
		return err
	}
	Log.Debug("GetExchangeRates", "resp", resp)
	return nil
}
