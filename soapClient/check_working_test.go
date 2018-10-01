package main

import (
	"testing"
	"fmt"
	"net/http"
	"bytes"
	"crypto/tls"
	"io/ioutil"
)

//var url="http://172.18.121.145/scripts"
//var url="https://172.18.121.146/scripts"
//url := "http://localhost:63053"

func TestInsert(t *testing.T) {
	imsi := 257027519747529
	bytexml := fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		<AddRequest xmlns="urn:bicwsdl">
			<imsi>%d</imsi>
			<groupId>4</groupId>
		</AddRequest>
		</SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`, imsi)
	// url := "http://soap.local/"
	//url := "http://localhost:63053"
	//url := "https://172.18.121.145/scripts"
	var xmlStr = []byte(string(bytexml))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(xmlStr))
	// req.Header.Set("Authorization", "Basic dXNlcjE6MTIzNDU2")
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
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body)==`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		  <AddRequestResponse xmlns="urn:bicwsdl">
			<response>
			  <errorCode>1002</errorCode>
			</response>
		  </AddRequestResponse>
		</SOAP-ENV:Body>
	  </SOAP-ENV:Envelope>`{
	  	t.Error("Row exist error 1002", string(body))
	}
	if string(body)==`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		  <AddRequestResponse xmlns="urn:bicwsdl">
			<response>
			  <errorCode>2000</errorCode>
			</response>
		  </AddRequestResponse>
		</SOAP-ENV:Body>
	  </SOAP-ENV:Envelope>`{
		t.Error("Server unrichble error 2000", string(body))
	}
	if string(body)==`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		  <AddRequestResponse xmlns="urn:bicwsdl">
			<response>
			  <errorCode>800</errorCode>
			</response>
		  </AddRequestResponse>
		</SOAP-ENV:Body>
	  </SOAP-ENV:Envelope>`{
		t.Error("Error 800 ", string(body))
	}
}

func TestUpdate(t *testing.T) {
	imsi := 257027519747529
	bytexml := fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		<UpdateRequest xmlns="urn:bicwsdl">
			<imsi>%d</imsi>
			<newImsi>257027519747530</newImsi>
		</UpdateRequest>
		</SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`, imsi)
	// url := "http://soap.local/"
	//url := "http://localhost:63053"
	//url := "https://172.18.121.145/scripts"
	var xmlStr = []byte(string(bytexml))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(xmlStr))
	// req.Header.Set("Authorization", "Basic dXNlcjE6MTIzNDU2")
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
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body)==`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		  <UpdateRequestResponse xmlns="urn:bicwsdl">
			<response>
			  <errorCode>1002</errorCode>
			</response>
		  </UpdateRequestResponse>
		</SOAP-ENV:Body>
	  </SOAP-ENV:Envelope>`{
		t.Error("Row not exist error 1002 ", string(body))
	}
	if string(body)==`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		  <UpdateRequestResponse xmlns="urn:bicwsdl">
			<response>
			  <errorCode>2000</errorCode>
			</response>
		  </UpdateRequestResponse>
		</SOAP-ENV:Body>
	  </SOAP-ENV:Envelope>`{
		t.Error("Server unrichble error 2000 ", string(body))
	}
	if string(body)==`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		  <UpdateRequestResponse xmlns="urn:bicwsdl">
			<response>
			  <errorCode>800</errorCode>
			</response>
		  </UpdateRequestResponse>
		</SOAP-ENV:Body>
	  </SOAP-ENV:Envelope>`{
		t.Error("Error 800 ", string(body))
	}

}

func TestDelete(t *testing.T) {
	imsi := 257027519747530
	bytexml := fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		<DeleteRequest xmlns="urn:bicwsdl">
			<imsi>%d</imsi>
			<groupId>4</groupId>
		</DeleteRequest>
		</SOAP-ENV:Body>
	</SOAP-ENV:Envelope>`, imsi)
	// url := "http://soap.local/"
	//url := "http://localhost:63053"
	//url := "https://172.18.121.145/scripts"
	var xmlStr = []byte(string(bytexml))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(xmlStr))
	// req.Header.Set("Authorization", "Basic dXNlcjE6MTIzNDU2")
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
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body)==`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		  <DeleteRequestResponse xmlns="urn:bicwsdl">
			<response>
			  <errorCode>1002</errorCode>
			</response>
		  </DeleteRequestResponse>
		</SOAP-ENV:Body>
	  </SOAP-ENV:Envelope>`{
		t.Error("Row not exist error 1002 ", string(body))
	}
	if string(body)==`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		  <DeleteRequestResponse xmlns="urn:bicwsdl">
			<response>
			  <errorCode>1002</errorCode>
			</response>
		  </DeleteRequestResponse>
		</SOAP-ENV:Body>
	  </SOAP-ENV:Envelope>`{
		t.Error("Error 2000 ", string(body))
	}
	if string(body)==`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		  <DeleteRequestResponse xmlns="urn:bicwsdl">
			<response>
			  <errorCode>1002</errorCode>
			</response>
		  </DeleteRequestResponse>
		</SOAP-ENV:Body>
	  </SOAP-ENV:Envelope>`{
		t.Error("Error 800 ", string(body))
	}
}