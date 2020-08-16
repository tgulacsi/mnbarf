package mnb

//go:generate sh -c "curl http://www.mnb.hu/arfolyamok.asmx?WSDL | xmllint --format >arfolyamok.wsdl"
//go:generate
//go:generate sh -c "curl http://www.mnb.hu/alapkamat.asmx?WSDL | xmllint --format /alapkamat.wsdl"
