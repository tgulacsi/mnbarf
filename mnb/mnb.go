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

package mnb

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rogpeppe/retry"
)

const (
	ArfolyamokURL = "http://www.mnb.hu/arfolyamok.asmx"
	AlapkamatURL  = "http://www.mnb.hu/alapkamat.asmx"
)

func NewMNBArfolyamService(URL string, client *http.Client, Log func(...interface{}) error) MNBArfolyamService {
	return MNBArfolyamService{MNB: NewMNB(URL, client, Log)}
}
func NewMNBAlapkamatService(URL string, client *http.Client, Log func(...interface{}) error) MNBAlapkamatService {
	return MNBAlapkamatService{MNB: NewMNB(URL, client, Log)}
}
func NewMNB(URL string, client *http.Client, Log func(...interface{}) error) MNB {
	return MNB{URL: URL, Log: Log, Client: client}
}

type MNB struct {
	URL string
	Log func(...interface{}) error
	*http.Client
}
type MNBAlapkamatService struct {
	MNB
}

/*
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
   <s:Body>
      <GetCurrentCentralBankBaseRateResponse xmlns="http://www.mnb.hu/webservices/" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
         <GetCurrentCentralBankBaseRateResult>&lt;MNBCurrentCentralBankBaseRate>&lt;BaseRate publicationDate="2020-07-21">0,60&lt;/BaseRate>&lt;/MNBCurrentCentralBankBaseRate></GetCurrentCentralBankBaseRateResult>
      </GetCurrentCentralBankBaseRateResponse>
   </s:Body>
</s:Envelope>
*/
// &lt;MNBCurrentCentralBankBaseRate&gt;&lt;BaseRate publicationDate="2015-06-24"&gt;1,5000&lt;/BaseRate&gt;&lt;/MNBCurrentCentralBankBaseRate&gt
type MNBCurrentCentralBankBaseRate struct {
	BaseRate MNBBaseRate
}
type MNBBaseRate struct {
	XMLName     xml.Name `xml:"BaseRate" json:"-"`
	Publication Date     `xml:"publicationDate,attr"`
	Rate        Double   `xml:",chardata"`
}

func (m MNBAlapkamatService) GetCurrentCentralBankBaseRate(ctx context.Context) (MNBBaseRate, error) {
	b, err := m.call(ctx,
		AlapkamatURL,
		"http://www.mnb.hu/webservices/MNBAlapkamatServiceSoap/GetCurrentCentralBankBaseRate",
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:web="http://www.mnb.hu/webservices/">
   <soapenv:Header/>
   <soapenv:Body>
      <web:GetCurrentCentralBankBaseRate/>
   </soapenv:Body>
</soapenv:Envelope>`)
	if err != nil {
		return MNBBaseRate{}, err
	}
	var res MNBCurrentCentralBankBaseRate
	err = xml.Unmarshal(b, &res)
	return res.BaseRate, err
}
func (m MNBAlapkamatService) GetCurrentBaseRate(ctx context.Context) (MNBBaseRate, error) {
	return m.GetCurrentCentralBankBaseRate(ctx)
}

/*
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
   <s:Body>
      <GetCentralBankBaseRateResponse xmlns="http://www.mnb.hu/webservices/" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
         <GetCentralBankBaseRateResult><![CDATA[<MNBCentralBankBaseRates><BaseRate publicationDate="2020-07-21">0,60</BaseRate><BaseRate publicationDate="2020-06-23">0,75</BaseRate></MNBCentralBankBaseRates>]]></GetCentralBankBaseRateResult>
      </GetCentralBankBaseRateResponse>
   </s:Body>
</s:Envelope>
*/
// <MNBCentralBankBaseRates><BaseRate publicationDate="2015-07-22">1,35</BaseRate><BaseRate publicationDate="2015-06-24">1,50</BaseRate><BaseRate publicationDate="2015-05-27">1,65</BaseRate><BaseRate publicationDate="2015-04-22">1,80</BaseRate><BaseRate publicationDate="2015-03-25">1,95</BaseRate></MNBCentralBankBaseRates>
type MNBCentralBankBaseRates struct {
	XMLName   xml.Name      `xml:"MNBCentralBankBaseRates" json:"-"`
	BaseRates []MNBBaseRate `xml:"BaseRate"`
}

func (m MNBAlapkamatService) GetCentralBankBaseRate(ctx context.Context, start, end time.Time) ([]MNBBaseRate, error) {
	b, err := m.call(ctx, AlapkamatURL, "http://www.mnb.hu/webservices/MNBAlapkamatServiceSoap/GetCentralBankBaseRate", m.GetCentralBankBaseRateXML(start, end))
	if err != nil {
		return nil, err
	}
	var res MNBCentralBankBaseRates
	err = xml.Unmarshal(b, &res)
	return res.BaseRates, err
}

func (m MNBAlapkamatService) GetBaseRates(ctx context.Context, start, end time.Time) ([]MNBBaseRate, error) {
	return m.GetCentralBankBaseRate(ctx, start, end)
}

type MNBArfolyamService struct {
	MNB
}

/*
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
   <s:Body>
      <GetCurrenciesResponse xmlns="http://www.mnb.hu/webservices/" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
         <GetCurrenciesResult><![CDATA[<MNBCurrencies><Currencies><Curr>HUF</Curr><Curr>EUR</Curr><Curr>AUD</Curr><Curr>BGN</Curr><Curr>BRL</Curr><Curr>CAD</Curr><Curr>CHF</Curr><Curr>CNY</Curr><Curr>CZK</Curr><Curr>DKK</Curr><Curr>GBP</Curr><Curr>HKD</Curr><Curr>HRK</Curr><Curr>IDR</Curr><Curr>ILS</Curr><Curr>INR</Curr><Curr>ISK</Curr><Curr>JPY</Curr><Curr>KRW</Curr><Curr>MXN</Curr><Curr>MYR</Curr><Curr>NOK</Curr><Curr>NZD</Curr><Curr>PHP</Curr><Curr>PLN</Curr><Curr>RON</Curr><Curr>RSD</Curr><Curr>RUB</Curr><Curr>SEK</Curr><Curr>SGD</Curr><Curr>THB</Curr><Curr>TRY</Curr><Curr>UAH</Curr><Curr>USD</Curr><Curr>ZAR</Curr><Curr>ATS</Curr><Curr>AUP</Curr><Curr>BEF</Curr><Curr>BGL</Curr><Curr>CSD</Curr><Curr>CSK</Curr><Curr>DDM</Curr><Curr>DEM</Curr><Curr>EEK</Curr><Curr>EGP</Curr><Curr>ESP</Curr><Curr>FIM</Curr><Curr>FRF</Curr><Curr>GHP</Curr><Curr>GRD</Curr><Curr>IEP</Curr><Curr>ITL</Curr><Curr>KPW</Curr><Curr>KWD</Curr><Curr>LBP</Curr><Curr>LTL</Curr><Curr>LUF</Curr><Curr>LVL</Curr><Curr>MNT</Curr><Curr>NLG</Curr><Curr>OAL</Curr><Curr>OBL</Curr><Curr>OFR</Curr><Curr>ORB</Curr><Curr>PKR</Curr><Curr>PTE</Curr><Curr>ROL</Curr><Curr>SDP</Curr><Curr>SIT</Curr><Curr>SKK</Curr><Curr>SUR</Curr><Curr>VND</Curr><Curr>XEU</Curr><Curr>XTR</Curr><Curr>YUD</Curr></Currencies></MNBCurrencies>]]></GetCurrenciesResult>
      </GetCurrenciesResponse>
   </s:Body>
</s:Envelope>
*/
type MNBCurrencies struct {
	Currencies []string `xml:"Currencies>Curr"`
}

func (m MNBArfolyamService) GetCurrencies(ctx context.Context) ([]string, error) {
	b, err := m.call(ctx,
		ArfolyamokURL,
		"http://www.mnb.hu/webservices/MNBArfolyamServiceSoap/GetCurrencies",
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:web="http://www.mnb.hu/webservices/">
   <soapenv:Header/>
   <soapenv:Body>
      <web:GetCurrencies/>
   </soapenv:Body>
</soapenv:Envelope>`)
	if err != nil {
		return nil, err
	}
	var res MNBCurrencies
	err = xml.Unmarshal(b, &res)
	return res.Currencies, err
}

/*
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
   <s:Body>
      <GetCurrencyUnitsResponse xmlns="http://www.mnb.hu/webservices/" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
         <GetCurrencyUnitsResult><![CDATA[<MNBCurrencyUnits><Units><Unit curr="HUF">1</Unit></Units></MNBCurrencyUnits>]]></GetCurrencyUnitsResult>
      </GetCurrencyUnitsResponse>
   </s:Body>
</s:Envelope>
*/
type MNBCurrencyUnits struct {
	Units []Unit `xml:"Units>Unit"`
}
type Unit struct {
	Currency string `xml:"curr,attr"`
	Unit     Double `xml:",chardata"`
}

func (m MNBArfolyamService) GetCurrencyUnits(ctx context.Context, currency string) ([]Unit, error) {
	b, err := m.call(ctx, ArfolyamokURL, "http://www.mnb.hu/webservices/MNBArfolyamServiceSoap/GetCurrencyUnits", m.GetCurrencyUnitsXML(currency))
	if err != nil {
		return nil, err
	}
	var res MNBCurrencyUnits
	err = xml.Unmarshal(b, &res)
	return res.Units, err
}

/*
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
   <s:Body>
      <GetCurrentExchangeRatesResponse xmlns="http://www.mnb.hu/webservices/" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
         <GetCurrentExchangeRatesResult><![CDATA[<MNBCurrentExchangeRates><Day date="2020-08-14"><Rate unit="1" curr="AUD">209,44</Rate><Rate unit="1" curr="BGN">177,03</Rate><Rate unit="1" curr="BRL">54,58</Rate><Rate unit="1" curr="CAD">221,22</Rate><Rate unit="1" curr="CHF">321,99</Rate><Rate unit="1" curr="CNY">42,16</Rate><Rate unit="1" curr="CZK">13,26</Rate><Rate unit="1" curr="DKK">46,48</Rate><Rate unit="1" curr="EUR">346,25</Rate><Rate unit="1" curr="GBP">383,23</Rate><Rate unit="1" curr="HKD">37,81</Rate><Rate unit="1" curr="HRK">45,96</Rate><Rate unit="100" curr="IDR">1,98</Rate><Rate unit="1" curr="ILS">86,1</Rate><Rate unit="1" curr="INR">3,91</Rate><Rate unit="1" curr="ISK">2,15</Rate><Rate unit="100" curr="JPY">274,56</Rate><Rate unit="100" curr="KRW">24,7</Rate><Rate unit="1" curr="MXN">13,2</Rate><Rate unit="1" curr="MYR">69,87</Rate><Rate unit="1" curr="NOK">32,86</Rate><Rate unit="1" curr="NZD">191,57</Rate><Rate unit="1" curr="PHP">6,02</Rate><Rate unit="1" curr="PLN">78,69</Rate><Rate unit="1" curr="RON">71,59</Rate><Rate unit="1" curr="RSD">2,94</Rate><Rate unit="1" curr="RUB">3,99</Rate><Rate unit="1" curr="SEK">33,63</Rate><Rate unit="1" curr="SGD">213,53</Rate><Rate unit="1" curr="THB">9,42</Rate><Rate unit="1" curr="TRY">39,76</Rate><Rate unit="1" curr="UAH">10,71</Rate><Rate unit="1" curr="USD">293,01</Rate><Rate unit="1" curr="ZAR">16,77</Rate></Day></MNBCurrentExchangeRates>]]></GetCurrentExchangeRatesResult>
      </GetCurrentExchangeRatesResponse>
   </s:Body>
</s:Envelope>
*/
type MNBCurrentExchangeRates struct {
	Day DayRates
}

type DayRates struct {
	Day   Date   `xml:"date,attr"`
	Rates []Rate `xml:"Rate"`
}

type Rate struct {
	Currency string `xml:"curr,attr"`
	Unit     int    `xml:"unit,attr"`
	Rate     Double `xml:",chardata"`
}

func (m MNBArfolyamService) GetCurrentExchangeRates(ctx context.Context) (DayRates, error) {
	b, err := m.call(ctx,
		ArfolyamokURL,
		"http://www.mnb.hu/webservices/MNBArfolyamServiceSoap/GetCurrentExchangeRates",
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:web="http://www.mnb.hu/webservices/">
   <soapenv:Header/>
   <soapenv:Body>
      <web:GetCurrentExchangeRates/>
   </soapenv:Body>
</soapenv:Envelope>`)
	if err != nil {
		return DayRates{}, err
	}
	var res MNBCurrentExchangeRates
	err = xml.Unmarshal(b, &res)
	return res.Day, err
}

/*
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
   <s:Body>
      <GetDateIntervalResponse xmlns="http://www.mnb.hu/webservices/" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
         <GetDateIntervalResult>&lt;MNBStoredInterval>&lt;DateInterval startdate="1949-01-03" enddate="2020-08-14" />&lt;/MNBStoredInterval></GetDateIntervalResult>
      </GetDateIntervalResponse>
   </s:Body>
</s:Envelope>
*/

type MNBStoredInterval struct {
	Interval DateInterval `xml:"DateInterva"`
}
type DateInterval struct {
	Start Date `xml:"startdate,attr"`
	End   Date `xml:"enddate,attr"`
}

func (m MNBArfolyamService) GetDateIntervalResponse(ctx context.Context) (DateInterval, error) {
	b, err := m.call(ctx,
		ArfolyamokURL,
		"http://www.mnb.hu/webservices/MNBArfolyamServiceSoap/GetDateInterval",
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:web="http://www.mnb.hu/webservices/">
   <soapenv:Header/>
   <soapenv:Body>
      <web:GetDateInterval/>
   </soapenv:Body>
</soapenv:Envelope>`)
	if err != nil {
		return DateInterval{}, err
	}
	var res MNBStoredInterval
	err = xml.Unmarshal(b, &res)
	return res.Interval, err
}

/*
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
   <s:Body>
      <GetInfoResponse xmlns="http://www.mnb.hu/webservices/" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
         <GetInfoResult><![CDATA[<MNBExchangeRatesQueryValues><FirstDate>1949-01-03</FirstDate><LastDate>2020-08-14</LastDate><Currencies><Curr>HUF</Curr><Curr>EUR</Curr><Curr>AUD</Curr><Curr>BGN</Curr><Curr>BRL</Curr><Curr>CAD</Curr><Curr>CHF</Curr><Curr>CNY</Curr><Curr>CZK</Curr><Curr>DKK</Curr><Curr>GBP</Curr><Curr>HKD</Curr><Curr>HRK</Curr><Curr>IDR</Curr><Curr>ILS</Curr><Curr>INR</Curr><Curr>ISK</Curr><Curr>JPY</Curr><Curr>KRW</Curr><Curr>MXN</Curr><Curr>MYR</Curr><Curr>NOK</Curr><Curr>NZD</Curr><Curr>PHP</Curr><Curr>PLN</Curr><Curr>RON</Curr><Curr>RSD</Curr><Curr>RUB</Curr><Curr>SEK</Curr><Curr>SGD</Curr><Curr>THB</Curr><Curr>TRY</Curr><Curr>UAH</Curr><Curr>USD</Curr><Curr>ZAR</Curr><Curr>ATS</Curr><Curr>AUP</Curr><Curr>BEF</Curr><Curr>BGL</Curr><Curr>CSD</Curr><Curr>CSK</Curr><Curr>DDM</Curr><Curr>DEM</Curr><Curr>EEK</Curr><Curr>EGP</Curr><Curr>ESP</Curr><Curr>FIM</Curr><Curr>FRF</Curr><Curr>GHP</Curr><Curr>GRD</Curr><Curr>IEP</Curr><Curr>ITL</Curr><Curr>KPW</Curr><Curr>KWD</Curr><Curr>LBP</Curr><Curr>LTL</Curr><Curr>LUF</Curr><Curr>LVL</Curr><Curr>MNT</Curr><Curr>NLG</Curr><Curr>OAL</Curr><Curr>OBL</Curr><Curr>OFR</Curr><Curr>ORB</Curr><Curr>PKR</Curr><Curr>PTE</Curr><Curr>ROL</Curr><Curr>SDP</Curr><Curr>SIT</Curr><Curr>SKK</Curr><Curr>SUR</Curr><Curr>VND</Curr><Curr>XEU</Curr><Curr>XTR</Curr><Curr>YUD</Curr></Currencies></MNBExchangeRatesQueryValues>]]></GetInfoResult>
      </GetInfoResponse>
   </s:Body>
</s:Envelope>
*/

type MNBExchangeRatesQueryValues struct {
	FirstDate  Date
	LastDate   Date
	Currencies []string `xml:"Currencies>Curr"`
}

func (m MNBArfolyamService) GetInfo(ctx context.Context) (MNBExchangeRatesQueryValues, error) {
	b, err := m.call(ctx,
		ArfolyamokURL,
		"http://www.mnb.hu/webservices/MNBArfolyamServiceSoap/GetInfo",
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:web="http://www.mnb.hu/webservices/">
   <soapenv:Header/>
   <soapenv:Body>
      <web:GetInfo/>
   </soapenv:Body>
</soapenv:Envelope>`)
	if err != nil {
		return MNBExchangeRatesQueryValues{}, err
	}
	var res MNBExchangeRatesQueryValues
	err = xml.Unmarshal(b, &res)
	return res, err
}

/*
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
   <s:Body>
      <GetExchangeRatesResponse xmlns="http://www.mnb.hu/webservices/" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
         <GetExchangeRatesResult><![CDATA[<MNBExchangeRates><Day date="2020-07-31"><Rate unit="1" curr="EUR">344,74</Rate></Day><Day date="2020-07-30"><Rate unit="1" curr="EUR">345,71</Rate></Day><Day date="2020-07-29"><Rate unit="1" curr="EUR">347,25</Rate></Day><Day date="2020-07-28"><Rate unit="1" curr="EUR">346,15</Rate></Day><Day date="2020-07-27"><Rate unit="1" curr="EUR">345,79</Rate></Day><Day date="2020-07-24"><Rate unit="1" curr="EUR">347,54</Rate></Day><Day date="2020-07-23"><Rate unit="1" curr="EUR">346,73</Rate></Day><Day date="2020-07-22"><Rate unit="1" curr="EUR">350,24</Rate></Day><Day date="2020-07-21"><Rate unit="1" curr="EUR">351,67</Rate></Day><Day date="2020-07-20"><Rate unit="1" curr="EUR">352,26</Rate></Day><Day date="2020-07-17"><Rate unit="1" curr="EUR">353,78</Rate></Day><Day date="2020-07-16"><Rate unit="1" curr="EUR">353,98</Rate></Day><Day date="2020-07-15"><Rate unit="1" curr="EUR">353,87</Rate></Day><Day date="2020-07-14"><Rate unit="1" curr="EUR">355,10</Rate></Day><Day date="2020-07-13"><Rate unit="1" curr="EUR">353,84</Rate></Day><Day date="2020-07-10"><Rate unit="1" curr="EUR">353,73</Rate></Day><Day date="2020-07-09"><Rate unit="1" curr="EUR">354,34</Rate></Day><Day date="2020-07-08"><Rate unit="1" curr="EUR">355,07</Rate></Day><Day date="2020-07-07"><Rate unit="1" curr="EUR">353,78</Rate></Day><Day date="2020-07-06"><Rate unit="1" curr="EUR">352,80</Rate></Day><Day date="2020-07-03"><Rate unit="1" curr="EUR">351,17</Rate></Day><Day date="2020-07-02"><Rate unit="1" curr="EUR">351,66</Rate></Day><Day date="2020-07-01"><Rate unit="1" curr="EUR">353,63</Rate></Day><Day date="2020-06-30"><Rate unit="1" curr="EUR">356,57</Rate></Day><Day date="2020-06-29"><Rate unit="1" curr="EUR">355,65</Rate></Day><Day date="2020-06-26"><Rate unit="1" curr="EUR">354,95</Rate></Day><Day date="2020-06-25"><Rate unit="1" curr="EUR">353,84</Rate></Day><Day date="2020-06-24"><Rate unit="1" curr="EUR">350,84</Rate></Day><Day date="2020-06-23"><Rate unit="1" curr="EUR">349,10</Rate></Day><Day date="2020-06-22"><Rate unit="1" curr="EUR">346,25</Rate></Day><Day date="2020-06-19"><Rate unit="1" curr="EUR">346,18</Rate></Day><Day date="2020-06-18"><Rate unit="1" curr="EUR">344,96</Rate></Day><Day date="2020-06-17"><Rate unit="1" curr="EUR">344,53</Rate></Day><Day date="2020-06-16"><Rate unit="1" curr="EUR">345,67</Rate></Day><Day date="2020-06-15"><Rate unit="1" curr="EUR">347,21</Rate></Day><Day date="2020-06-12"><Rate unit="1" curr="EUR">345,86</Rate></Day><Day date="2020-06-11"><Rate unit="1" curr="EUR">344,70</Rate></Day><Day date="2020-06-10"><Rate unit="1" curr="EUR">343,77</Rate></Day><Day date="2020-06-09"><Rate unit="1" curr="EUR">344,83</Rate></Day><Day date="2020-06-08"><Rate unit="1" curr="EUR">343,62</Rate></Day><Day date="2020-06-05"><Rate unit="1" curr="EUR">344,67</Rate></Day><Day date="2020-06-04"><Rate unit="1" curr="EUR">345,57</Rate></Day><Day date="2020-06-03"><Rate unit="1" curr="EUR">345,94</Rate></Day><Day date="2020-06-02"><Rate unit="1" curr="EUR">344,75</Rate></Day><Day date="2020-05-29"><Rate unit="1" curr="EUR">348,35</Rate></Day><Day date="2020-05-28"><Rate unit="1" curr="EUR">350,01</Rate></Day><Day date="2020-05-27"><Rate unit="1" curr="EUR">349,25</Rate></Day><Day date="2020-05-26"><Rate unit="1" curr="EUR">349,66</Rate></Day><Day date="2020-05-25"><Rate unit="1" curr="EUR">350,53</Rate></Day><Day date="2020-05-22"><Rate unit="1" curr="EUR">349,56</Rate></Day><Day date="2020-05-21"><Rate unit="1" curr="EUR">348,99</Rate></Day><Day date="2020-05-20"><Rate unit="1" curr="EUR">349,91</Rate></Day><Day date="2020-05-19"><Rate unit="1" curr="EUR">351,97</Rate></Day><Day date="2020-05-18"><Rate unit="1" curr="EUR">353,99</Rate></Day><Day date="2020-05-15"><Rate unit="1" curr="EUR">353,88</Rate></Day><Day date="2020-05-14"><Rate unit="1" curr="EUR">354,33</Rate></Day><Day date="2020-05-13"><Rate unit="1" curr="EUR">353,30</Rate></Day><Day date="2020-05-12"><Rate unit="1" curr="EUR">350,70</Rate></Day><Day date="2020-05-11"><Rate unit="1" curr="EUR">349,68</Rate></Day><Day date="2020-05-08"><Rate unit="1" curr="EUR">349,57</Rate></Day><Day date="2020-05-07"><Rate unit="1" curr="EUR">350,56</Rate></Day><Day date="2020-05-06"><Rate unit="1" curr="EUR">349,42</Rate></Day><Day date="2020-05-05"><Rate unit="1" curr="EUR">352,06</Rate></Day><Day date="2020-05-04"><Rate unit="1" curr="EUR">353,39</Rate></Day><Day date="2020-04-30"><Rate unit="1" curr="EUR">353,01</Rate></Day><Day date="2020-04-29"><Rate unit="1" curr="EUR">355,73</Rate></Day><Day date="2020-04-28"><Rate unit="1" curr="EUR">355,31</Rate></Day><Day date="2020-04-27"><Rate unit="1" curr="EUR">353,80</Rate></Day><Day date="2020-04-24"><Rate unit="1" curr="EUR">356,15</Rate></Day><Day date="2020-04-23"><Rate unit="1" curr="EUR">357,04</Rate></Day><Day date="2020-04-22"><Rate unit="1" curr="EUR">354,30</Rate></Day><Day date="2020-04-21"><Rate unit="1" curr="EUR">355,09</Rate></Day><Day date="2020-04-20"><Rate unit="1" curr="EUR">353,24</Rate></Day><Day date="2020-04-17"><Rate unit="1" curr="EUR">350,56</Rate></Day><Day date="2020-04-16"><Rate unit="1" curr="EUR">350,15</Rate></Day><Day date="2020-04-15"><Rate unit="1" curr="EUR">351,34</Rate></Day><Day date="2020-04-14"><Rate unit="1" curr="EUR">351,75</Rate></Day><Day date="2020-04-09"><Rate unit="1" curr="EUR">355,66</Rate></Day><Day date="2020-04-08"><Rate unit="1" curr="EUR">358,76</Rate></Day><Day date="2020-04-07"><Rate unit="1" curr="EUR">359,95</Rate></Day><Day date="2020-04-06"><Rate unit="1" curr="EUR">363,35</Rate></Day><Day date="2020-04-03"><Rate unit="1" curr="EUR">364,42</Rate></Day><Day date="2020-04-02"><Rate unit="1" curr="EUR">361,76</Rate></Day><Day date="2020-04-01"><Rate unit="1" curr="EUR">364,57</Rate></Day><Day date="2020-03-31"><Rate unit="1" curr="EUR">359,09</Rate></Day><Day date="2020-03-30"><Rate unit="1" curr="EUR">357,21</Rate></Day><Day date="2020-03-27"><Rate unit="1" curr="EUR">354,30</Rate></Day><Day date="2020-03-26"><Rate unit="1" curr="EUR">357,79</Rate></Day><Day date="2020-03-25"><Rate unit="1" curr="EUR">354,49</Rate></Day><Day date="2020-03-24"><Rate unit="1" curr="EUR">350,33</Rate></Day><Day date="2020-03-23"><Rate unit="1" curr="EUR">351,55</Rate></Day><Day date="2020-03-20"><Rate unit="1" curr="EUR">349,85</Rate></Day><Day date="2020-03-19"><Rate unit="1" curr="EUR">357,62</Rate></Day><Day date="2020-03-18"><Rate unit="1" curr="EUR">350,17</Rate></Day><Day date="2020-03-17"><Rate unit="1" curr="EUR">347,34</Rate></Day><Day date="2020-03-16"><Rate unit="1" curr="EUR">340,17</Rate></Day><Day date="2020-03-13"><Rate unit="1" curr="EUR">338,00</Rate></Day><Day date="2020-03-12"><Rate unit="1" curr="EUR">337,51</Rate></Day><Day date="2020-03-11"><Rate unit="1" curr="EUR">334,86</Rate></Day><Day date="2020-03-10"><Rate unit="1" curr="EUR">335,93</Rate></Day><Day date="2020-03-09"><Rate unit="1" curr="EUR">336,22</Rate></Day><Day date="2020-03-06"><Rate unit="1" curr="EUR">337,60</Rate></Day><Day date="2020-03-05"><Rate unit="1" curr="EUR">336,04</Rate></Day><Day date="2020-03-04"><Rate unit="1" curr="EUR">335,10</Rate></Day><Day date="2020-03-03"><Rate unit="1" curr="EUR">337,03</Rate></Day><Day date="2020-03-02"><Rate unit="1" curr="EUR">337,47</Rate></Day><Day date="2020-02-28"><Rate unit="1" curr="EUR">339,88</Rate></Day><Day date="2020-02-27"><Rate unit="1" curr="EUR">339,12</Rate></Day><Day date="2020-02-26"><Rate unit="1" curr="EUR">339,56</Rate></Day><Day date="2020-02-25"><Rate unit="1" curr="EUR">337,21</Rate></Day><Day date="2020-02-24"><Rate unit="1" curr="EUR">338,32</Rate></Day><Day date="2020-02-21"><Rate unit="1" curr="EUR">337,76</Rate></Day><Day date="2020-02-20"><Rate unit="1" curr="EUR">337,86</Rate></Day><Day date="2020-02-19"><Rate unit="1" curr="EUR">335,10</Rate></Day><Day date="2020-02-18"><Rate unit="1" curr="EUR">335,44</Rate></Day><Day date="2020-02-17"><Rate unit="1" curr="EUR">334,67</Rate></Day><Day date="2020-02-14"><Rate unit="1" curr="EUR">334,94</Rate></Day><Day date="2020-02-13"><Rate unit="1" curr="EUR">338,83</Rate></Day><Day date="2020-02-12"><Rate unit="1" curr="EUR">338,76</Rate></Day><Day date="2020-02-11"><Rate unit="1" curr="EUR">337,71</Rate></Day><Day date="2020-02-10"><Rate unit="1" curr="EUR">338,09</Rate></Day><Day date="2020-02-07"><Rate unit="1" curr="EUR">338,87</Rate></Day><Day date="2020-02-06"><Rate unit="1" curr="EUR">337,09</Rate></Day><Day date="2020-02-05"><Rate unit="1" curr="EUR">335,74</Rate></Day><Day date="2020-02-04"><Rate unit="1" curr="EUR">336,36</Rate></Day><Day date="2020-02-03"><Rate unit="1" curr="EUR">338,06</Rate></Day><Day date="2020-01-31"><Rate unit="1" curr="EUR">336,65</Rate></Day><Day date="2020-01-30"><Rate unit="1" curr="EUR">337,98</Rate></Day><Day date="2020-01-29"><Rate unit="1" curr="EUR">337,61</Rate></Day><Day date="2020-01-28"><Rate unit="1" curr="EUR">337,36</Rate></Day><Day date="2020-01-27"><Rate unit="1" curr="EUR">337,16</Rate></Day><Day date="2020-01-24"><Rate unit="1" curr="EUR">336,17</Rate></Day><Day date="2020-01-23"><Rate unit="1" curr="EUR">336,90</Rate></Day><Day date="2020-01-22"><Rate unit="1" curr="EUR">335,09</Rate></Day><Day date="2020-01-21"><Rate unit="1" curr="EUR">335,53</Rate></Day><Day date="2020-01-20"><Rate unit="1" curr="EUR">336,91</Rate></Day><Day date="2020-01-17"><Rate unit="1" curr="EUR">335,49</Rate></Day><Day date="2020-01-16"><Rate unit="1" curr="EUR">333,83</Rate></Day><Day date="2020-01-15"><Rate unit="1" curr="EUR">333,21</Rate></Day><Day date="2020-01-14"><Rate unit="1" curr="EUR">332,65</Rate></Day><Day date="2020-01-13"><Rate unit="1" curr="EUR">334,98</Rate></Day><Day date="2020-01-10"><Rate unit="1" curr="EUR">333,84</Rate></Day><Day date="2020-01-09"><Rate unit="1" curr="EUR">331,58</Rate></Day><Day date="2020-01-08"><Rate unit="1" curr="EUR">331,40</Rate></Day><Day date="2020-01-07"><Rate unit="1" curr="EUR">330,71</Rate></Day><Day date="2020-01-06"><Rate unit="1" curr="EUR">329,98</Rate></Day><Day date="2020-01-03"><Rate unit="1" curr="EUR">329,45</Rate></Day><Day date="2020-01-02"><Rate unit="1" curr="EUR">329,99</Rate></Day></MNBExchangeRates>]]></GetExchangeRatesResult>
      </GetExchangeRatesResponse>
   </s:Body>
</s:Envelope>
*/

type MNBExchangeRates struct {
	Days []DayRates `xml:"Day"`
}

func (m MNBArfolyamService) GetExchangeRates(ctx context.Context, start, end time.Time, currencies ...string) ([]DayRates, error) {
	b, err := m.call(ctx, ArfolyamokURL, "http://www.mnb.hu/webservices/MNBArfolyamServiceSoap/GetExchangeRates", m.GetExchangeRatesXML(start, end, currencies...))
	if err != nil {
		return nil, err
	}
	var res MNBExchangeRates
	err = xml.Unmarshal(b, &res)
	return res.Days, err
}

// FindBody will find the first StartElement after soap:Body.
func FindBody(dec *xml.Decoder) ([]byte, error) {
	_, err := findSoapBody(dec)
	if err != nil {
		return nil, err
	}
	return nextCharAfterStart(dec)
}

// findSoapBody will find the soap:Body StartElement.
func findSoapBody(dec *xml.Decoder) (xml.StartElement, error) {
	return findSoapElt("body", dec)
}

func findSoapElt(name string, dec *xml.Decoder) (xml.StartElement, error) {
	var st xml.StartElement
	for {
		tok, err := dec.Token()
		if err != nil {
			return st, err
		}
		var ok bool
		if st, ok = tok.(xml.StartElement); ok {
			if strings.EqualFold(st.Name.Local, name) &&
				(st.Name.Space == "" ||
					st.Name.Space == "SOAP-ENV" ||
					st.Name.Space == "http://www.w3.org/2003/05/soap-envelope/" ||
					st.Name.Space == "http://schemas.xmlsoap.org/soap/envelope/") {
				return st, nil
			}
		}
	}
}

// nextStart finds the first StartElement
func nextCharAfterStart(dec *xml.Decoder) ([]byte, error) {
	var seenStart bool
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if _, ok := tok.(xml.StartElement); ok {
			seenStart = true
		} else if cd, ok := tok.(xml.CharData); ok && seenStart {
			return []byte(cd), nil
		} else if _, ok = tok.(xml.EndElement); ok {
			return nil, io.EOF
		}
	}
}

var retryStrategy = retry.Strategy{
	Delay:       100 * time.Millisecond,
	MaxDelay:    5 * time.Second,
	MaxDuration: 30 * time.Second,
	Factor:      2,
}

func (m MNB) call(ctx context.Context, defaultURL, action string, body string) ([]byte, error) {
	mLog := m.Log
	URL := m.URL
	if URL == "" {
		URL = defaultURL
	}
	reqS := xml.Header + body
	client := m.Client
	if client == nil {
		client = http.DefaultClient
	}

	var buf strings.Builder
	var firstErr error
	for iter := retryStrategy.Start(); ; {
		req, err := http.NewRequest("POST", URL, strings.NewReader(reqS))
		if err != nil {
			if mLog != nil {
				_ = mLog("msg", "request", "url", URL, "body", reqS)
			}
			return nil, err
		}
		req.GetBody = func() (io.ReadCloser, error) {
			return struct {
				io.Reader
				io.Closer
			}{strings.NewReader(reqS), io.NopCloser(nil)}, nil
		}
		req.Header.Set("SOAPAction", action)
		req.Header.Set("Content-Type", "text/xml; charset=utf-8")

		b, err := func() ([]byte, error) {
			start := time.Now()
			resp, err := client.Do(req.WithContext(ctx))
			dur := time.Since(start)
			if err != nil {
				if mLog != nil {
					_ = mLog("msg", "do", "url", URL, "body", reqS, "error", err)
				}
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				if mLog != nil {
					_ = mLog("url", req.URL, "body", reqS, "status", resp.Status)
				}
				return nil, fmt.Errorf("%s %q: %s", req.Method, req.URL, resp.Status)
			}
			buf.Reset()
			b, err := FindBody(xml.NewDecoder(io.TeeReader(resp.Body, &buf)))
			if err != nil {
				if mLog != nil {
					_ = mLog("msg", "FindBody", "url", URL, "request", reqS, "status", resp.Status, "response", buf.String(), "error", err)
				}
				return nil, fmt.Errorf("FindBody(%q): %w", buf.String(), err)
			}
			if mLog != nil {
				_ = mLog("msg", "FindBody", "url", URL, "request", reqS, "status", resp.Status, "response", buf.String(), "dur", dur, "data", string(b))
			}
			return append(make([]byte, 0, len(b)), b...), nil
		}()
		if err == nil {
			return b, nil
		}
		if firstErr == nil {
			firstErr = err
		}
		if !iter.Next(ctx.Done()) {
			return nil, firstErr
		}
	}
}

type nilLogger struct{}

func (nilLogger) Printf(string, ...interface{}) {}

type logLogger struct{ Log func(...interface{}) error }

func (lgr logLogger) Printf(pat string, args ...interface{}) { _ = lgr.Log(fmt.Sprintf(pat, args...)) }
func (lgr logLogger) msg(lvl, msg string, keysAndValues ...interface{}) {
	_ = lgr.Log(append(append(make([]interface{}, 0, 4+len(keysAndValues)), "msg", msg, "lvl", lvl), keysAndValues...)...)
}
func (lgr logLogger) Error(msg string, keysAndValues ...interface{}) {
	lgr.msg("ERROR", msg, keysAndValues...)
}

func (lgr logLogger) Info(msg string, keysAndValues ...interface{}) {
	lgr.msg("INFO", msg, keysAndValues...)
}

func (lgr logLogger) Debug(msg string, keysAndValues ...interface{}) {
	lgr.msg("DEBUG", msg, keysAndValues...)
}

func (lgr logLogger) Warn(msg string, keysAndValues ...interface{}) {
	lgr.msg("WARN", msg, keysAndValues...)
}
