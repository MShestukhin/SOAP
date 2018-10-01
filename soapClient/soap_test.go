package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	_ "github.com/lib/pq"
)


func BenchmarkSendAddPackag(b *testing.B) {
	clear()
	imsi := 257022345678989
	concurrency := 30
	sem := make(chan bool, concurrency)
	for i1 := 0; i1 < b.N; i1++ {
		//for i1 := 0; i1 < 1000; i1++ {
			sem <- true
			go func(imsi int) {
				defer func() { <-sem }()
				bytexml := fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
								<SOAP-ENV:Body>
								<DeleteRequest xmlns="urn:bicwsdl">
									<imsi>%d</imsi>
									<groupId>5</groupId>
								</DeleteRequest>
								</SOAP-ENV:Body>
							</SOAP-ENV:Envelope>`, imsi)

				// url := "http://steering.local/operator/inner247"
				// url := "http://soap.local/"
				//url := "https://172.18.121.145/scripts"
				_ = bytexml
				_ = url
				var xmlStr = []byte((bytexml))
				req, err := http.NewRequest("POST", url, bytes.NewBuffer(xmlStr))
				req.Header.Set("Authorization", "Basic c3Z5YXpjb20yOjEyMzQ1Ng==")
				req.Header.Set("Content-Type", "text/xml; charset=UTF-8")

				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}

				client := &http.Client{Transport: tr}
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
			}(imsi)
			imsi++
		//}
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}

func BenchmarkSendDeletePackag(b *testing.B) {
	imsi := 150012345678989
	concurrency := 30
	sem := make(chan bool, concurrency)
	for i1 := 0; i1 < b.N; i1++ {
		// for i1 := 0; i1 < 1000; i1++ {
		sem <- true
		go func(imsi int) {
			defer func() { <-sem }()
			bytexml := fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
								<SOAP-ENV:Body>
								<DeleteRequest xmlns="urn:bicwsdl">
									<imsi>%d</imsi>
									<groupId>4</groupId>
								</DeleteRequest>
								</SOAP-ENV:Body>
							</SOAP-ENV:Envelope>`, imsi)

			// url := "http://steering.local/operator/inner247"
			// url := "http://soap.local/"
			fmt.Print(imsi)
			url := "https://172.18.121.145/scripts"
			_ = bytexml
			_ = url
			var xmlStr = []byte((bytexml))
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(xmlStr))
			req.Header.Set("Authorization", "Basic dXNlcjE6MTIzNDU2")
			req.Header.Set("Content-Type", "text/xml; charset=UTF-8")

			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}

			client := &http.Client{Transport: tr}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
		}(imsi)
		imsi++
		// }
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}
