package mnb

// Generated by https://github.com/hooklift/gowsdl
// Do not modify
// Copyright (c) 2015, Hooklift. All rights reserved.
import (
	"encoding/xml"
	"time"

	gowsdl "github.com/hooklift/gowsdl/generator"
)

// against "unused imports"
var _ time.Time
var _ xml.Name

type GetCurrentCentralBankBaseRate struct {
	XMLName xml.Name `xml:"http://www.mnb.hu/webservices/ GetCurrentCentralBankBaseRate"`
}

type GetCurrentCentralBankBaseRateResponse struct {
	XMLName xml.Name `xml:"http://www.mnb.hu/webservices/ GetCurrentCentralBankBaseRateResponse"`

	GetCurrentCentralBankBaseRateResult string `xml:"GetCurrentCentralBankBaseRateResult,omitempty"`
}

type GetCentralBankBaseRate struct {
	XMLName xml.Name `xml:"http://www.mnb.hu/webservices/ GetCentralBankBaseRate"`

	StartDate string `xml:"startDate,omitempty"`

	EndDate string `xml:"endDate,omitempty"`
}

type GetCentralBankBaseRateResponse struct {
	XMLName xml.Name `xml:"http://www.mnb.hu/webservices/ GetCentralBankBaseRateResponse"`

	GetCentralBankBaseRateResult string `xml:"GetCentralBankBaseRateResult,omitempty"`
}

type MNBAlapkamatServiceSoap struct {
	client *gowsdl.SoapClient
}

func NewMNBAlapkamatServiceSoap(url string, tls bool) *MNBAlapkamatServiceSoap {
	if url == "" {
		url = "http://www.mnb.hu/alapkamat.asmx"
	}
	client := gowsdl.NewSoapClient(url, tls)

	return &MNBAlapkamatServiceSoap{
		client: client,
	}
}

func (service *MNBAlapkamatServiceSoap) GetCurrentCentralBankBaseRate(request *GetCurrentCentralBankBaseRate) (*GetCurrentCentralBankBaseRateResponse, error) {
	response := &GetCurrentCentralBankBaseRateResponse{}
	err := service.client.Call("http://www.mnb.hu/webservices/GetCurrentCentralBankBaseRate", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *MNBAlapkamatServiceSoap) GetCentralBankBaseRate(request *GetCentralBankBaseRate) (*GetCentralBankBaseRateResponse, error) {
	response := &GetCentralBankBaseRateResponse{}
	err := service.client.Call("http://www.mnb.hu/webservices/GetCentralBankBaseRate", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
