// Copyright 2026 Tamás Gulácsi.
//
// SPDX-License-Identifier: Apache-2.0

package mnb

//go : generate go tool gowsdl -p mnb_arf -o generated_arfolyamok.go "http://www.mnb.hu/arfolyamok.asmx?WSDL"
//go : generate go tool gowsdl -p mnb_kam -o generated_alapkamat.go "http://www.mnb.hu/alapkamat.asmx?WSDL"
//go : generate sed -i -e "/context/d; /github.com\\/hooklift\\/gowsdl\\/soap/d" mnb_arf/generated_arfolyamok.go mnb_kam/generated_alapkamat.go
//go:generate go tool qtc
