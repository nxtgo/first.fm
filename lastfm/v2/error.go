package lfm

import "encoding/xml"

type Envelope struct {
	XMLName xml.Name `xml:"lfm"`
	Status  string   `xml:"status,attr"`
	Error   *Error   `xml:"error"`
}

type Error struct {
	Code    int    `xml:"code,attr"`
	Message string `xml:",chardata"`
}
