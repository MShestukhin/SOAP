package main

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	_ "strings"
	"sync"
	"time"
)

type ReqXml struct {
	Xmlns   string `xml:"xmlns,attr"`
	Imsi    string `xml:"imsi"`
	GroupId string `xml:"groupId"`
}

type BodyXmlAdd struct {
	AddReq ReqXml `xml:"AddRequest"`
}

type BodyXmlDelete struct {
	AddReq ReqXml `xml:"DeleteRequest"`
}

type SoapXmlAdd struct {
	XMLName xml.Name   `xml:"SOAP-ENV:Envelope"`
	Env     string     `xml:"xmlns:SOAP-ENV,attr"`
	Xsi     string     `xml:"xmlns:xsi,attr"`
	Body    BodyXmlAdd `xml:"SOAP-ENV:Body"`
}

type SoapXmlDelete struct {
	XMLName xml.Name      `xml:"SOAP-ENV:Envelope"`
	Env     string        `xml:"xmlns:SOAP-ENV,attr"`
	Xsi     string        `xml:"xmlns:xsi,attr"`
	Body    BodyXmlDelete `xml:"SOAP-ENV:Body"`
}

type Config struct {
	Database struct {
		Host     string `json:"host"`
		Password string `json:"password"`
		User     string `json:"user"`
		Dbname   string `json:"dbname"`
		Port     string `json:"port"`
		SoapUser string `json:"soapUser"`
		MaxConnection int `json:"maxConnection"`
	} `json:"database"`
	logPath string `json:"path"`
}

type client struct {
	conn    *sql.DB
	logPath string
	prefImsi map[string]bool
	soapId    string
	sem chan bool
}
var nn client
//var url ="http://172.18.121.145/scripts"
//var url ="https://172.18.121.146/scripts"
var url = "http://localhost:63053"
func LoadConfiguration(configPath string) Config {
	var config Config
	configFile, _ := os.Open(configPath)
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	err := jsonParser.Decode(&config)
	fmt.Println(err)
	return config
}

func main() {
	//argsWithoutProg := os.Args[1:]
	//i, _ := strconv.ParseInt(argsWithoutProg[0], 10, 64)
	//i,_:=strconv.Atoi(argsWithoutProg[0])
	//fmt.Println(reflect.TypeOf(i))
	//clear()
	//globalTest(i)
	///*for i:=0;i<100000;i++{
		sendXmlWithout()
	//}*/
	//updateXmlWithout()
	//deleteXmlWithout()
	//deleteXmlSubscriber()
}

func  clear(){
	pathConf := "/opt/svyazcom/etc/serverSOAP/"
	config := LoadConfiguration(pathConf + "soap.conf")
	//fmt.Println(config.Database.MaxConnection)
	connStr := "user=" + config.Database.User + " dbname=" + config.Database.Dbname + " host=" + config.Database.Host + " password=" + config.Database.Password + " port=" + config.Database.Port + " sslmode=disable"
	//fmt.Println(connStr)
	db, err := sql.Open("postgres", connStr)
	//fmt.Println(err)
	//nn.conn = db

	/*tx, err := nn.conn.Begin()
	if err != nil {
		fmt.Println("Can not open transaction - "+err.Error(), nn.logPath)
	}
	defer tx.Commit()*/
	userQuery := fmt.Sprintf("select set_config('user.id', '17', false);")
	//db.Exec(userQuery)
	query:=userQuery+`delete
	FROM steer.grp_imsi
	where list_id=13;`
	//fmt.Println(query)
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println("error query - "+err.Error(), nn.logPath)
	}
	//defer tx.Commit()
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
func globalTest(argsWithoutProg int) {
		clear();
		imsi := 257022345678989
		concurrency := 50
		sem := make(chan bool, concurrency)
		countSend:=0
		countNoSend:=0
		mut := sync.Mutex{}
		for i1 := 0; i1 < argsWithoutProg; i1++ {
			//time.Sleep(time.Second)
			// for i1 := 0; i1 < 1000; i1++ {
			sem <- true
			go func(imsi int) {
				defer func() { <-sem }()
				//myrand := random(13, 15)
				bytexml := fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
				<SOAP-ENV:Body>
				<AddRequest xmlns="urn:bicwsdl">
					<imsi>%d</imsi>
					<groupId>13</groupId>
				</AddRequest>
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

				client := &http.Client{
					Transport: tr,
					Timeout: 15 * time.Second,
				}
				resp, err := client.Do(req)

				if err != nil {
					//println(err)
					//panic(err)
				}
				if resp != nil {
					defer resp.Body.Close()
					mut.Lock()
					countSend++
					mut.Unlock()
				}else{
					mut.Lock()
					countNoSend++
					mut.Unlock()
					//println(resp)
				}

			}(imsi)
			imsi++
			if i1%1000 == 0 {
				fmt.Println(i1)
			}
		}
		//fmt.Println(countSend)
		//fmt.Println(countNoSend)
		//fmt.Println(countNoSend + countSend)
}
func sendXmlWithout() {
	//257027519747522
	imsi := 257022345678792
	bytexml := fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		<AddRequest xmlns="urn:bicwsdl">
			<imsi>%d</imsi>
			<groupId>13</groupId>
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
	fmt.Println(resp)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func updateXmlWithout() {
	imsi := 257022345678792
	bytexml := fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		<UpdateRequest xmlns="urn:bicwsdl">
			<imsi>%d</imsi>
			<newImsi>257022345678793</newImsi>
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

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
func deleteXmlWithout() {
	imsi := 257022345678792
	bytexml := fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		<DeleteRequest xmlns="urn:bicwsdl">
			<imsi>%d</imsi>
			<groupId>13</groupId>
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

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func deleteXmlSubscriber() {
	imsi := 257022345678793
	bytexml := fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		<DeleteSubscriber xmlns="urn:bicwsdl">
			<imsi>%d</imsi>
		</DeleteSubscriber>
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

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func sendXml() {
	var requestxml SoapXmlAdd
	requestxml.Env = "http://schemas.xmlsoap.org/soap/envelope/"
	requestxml.Xsi = "http://www.w3.org/2001/XMLSchema-instance"
	requestxml.Body.AddReq.Xmlns = "urn:bicwsdl"
	requestxml.Body.AddReq.Imsi = "250012345678989"
	requestxml.Body.AddReq.GroupId = "1"
	bytexml, err := xml.MarshalIndent(&requestxml, "", "  ")

	url := "http://soap.local/"
	fmt.Println("URL:>", url)

	var xmlStr = []byte(string(bytexml))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(xmlStr))
	req.Header.Set("Authorization", "Basic dXNlcjE6MTIzNDU2")
	req.Header.Set("Content-Type", "text/xml; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func readXml() {
	xmlFile, err := os.Open("soap.xml")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened soap.xml")
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var users SoapXmlAdd
	xml.Unmarshal(byteValue, &users)
	fmt.Println(users)
}
