package main

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
)

const (
	port = ":63053"
)

type ReqXml struct {
	Xmlns   string `xml:"xmlns,attr"`
	Imsi    string `xml:"imsi"`
	GroupId string `xml:"groupId"`
}

type BodyXml struct {
	AddReq    ReqXml `xml:"AddRequest"`
	DeleteReq ReqXml `xml:"DeleteRequest"`
}
type SoapXml struct {
	XMLName xml.Name //`xml:"SOAP-ENV:Envelope"`
	Env     string   `xml:"xmlns:SOAP-ENV,attr"`
	Xsi     string   `xml:"xmlns:xsi,attr"`
	Body    BodyXml  //`xml:"SOAP-ENV:Body"`
}

type ErrorCode struct {
	ErrorCode int `xml:"errorCode"`
}

type Response struct {
	Xmlns    string    `xml:"xmlns,attr"`
	Response ErrorCode `xml:"response"`
}

type BodyXmlAdd struct {
	Req Response `xml:"AddRequestResponse"`
}
type AnswerXmlAdd struct {
	XMLName xml.Name   `xml:"SOAP-ENV:Envelope"`
	Env     string     `xml:"xmlns:SOAP-ENV,attr"`
	Xsi     string     `xml:"xmlns:xsi,attr"`
	Body    BodyXmlAdd `xml:"SOAP-ENV:Body"`
}

type BodyXmlDelete struct {
	Req Response `xml:"DeleteRequestResponse"`
}
type AnswerXmlDelete struct {
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
	} `json:"database"`
	logPath string `json:"path"`
}

type server struct {
	conn    *sql.DB
	logPath string
}

func (nn *server) processing(w http.ResponseWriter, r *http.Request) {
	xmlFile, _ := ioutil.ReadAll(r.Body)
	// fmt.Println(string(xmlFile))
	r.Body.Close()
	var q SoapXml
	_ = xmlFile
	xml.Unmarshal((xmlFile), &q)
	// fmt.Println(q)
	if q.Body.AddReq.Imsi != "" && q.Body.DeleteReq.Imsi == "" {
		status, err := checkData(q.Body.AddReq.Imsi, q.Body.AddReq.GroupId)
		if err == nil {
			status, err = nn.addImsi(q.Body.AddReq.Imsi, q.Body.AddReq.GroupId)
			nn.logQuery("add query", q.Body.AddReq.Imsi, q.Body.AddReq.GroupId, r.RemoteAddr, status, err)
		} else {
			nn.logQuery("add query", q.Body.AddReq.Imsi, q.Body.AddReq.GroupId, r.RemoteAddr, status, err)
		}
		x := []byte(fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
			<SOAP-ENV:Body>
			  <AddRequestResponse xmlns="urn:bicwsdl">
				<response>
				  <errorCode>%d</errorCode>
				</response>
			  </AddRequestResponse>
			</SOAP-ENV:Body>
		  </SOAP-ENV:Envelope>`, status))
		w.Header().Set("Content-Type", "application/xml")
		w.Write(x)
	} else if q.Body.AddReq.Imsi == "" && q.Body.DeleteReq.Imsi != "" {
		status, err := checkData(q.Body.DeleteReq.Imsi, q.Body.DeleteReq.GroupId)
		if err == nil {
			status, err = nn.deleteImsi(q.Body.DeleteReq.Imsi, q.Body.DeleteReq.GroupId)
			nn.logQuery("delete query", q.Body.DeleteReq.Imsi, q.Body.DeleteReq.GroupId, r.RemoteAddr, status, err)
		} else {
			nn.logQuery("delete query", q.Body.DeleteReq.Imsi, q.Body.DeleteReq.GroupId, r.RemoteAddr, status, err)
		}
		x := []byte(fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
			<SOAP-ENV:Body>
			  <DeleteRequestResponse xmlns="urn:bicwsdl">
				<response>
				  <errorCode>%d</errorCode>
				</response>
			  </DeleteRequestResponse>
			</SOAP-ENV:Body>
		  </SOAP-ENV:Envelope>`, status))
		w.Header().Set("Content-Type", "application/xml")
		w.Write(x)
	} else {
		loging(fmt.Sprintf("ip - s% incorrect query addImsi - %s deleteImsi - %s", r.RemoteAddr, q.Body.AddReq.Imsi, q.Body.DeleteReq.Imsi), nn.logPath)
		x := []byte(fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
			<SOAP-ENV:Body>
			  <DeleteRequestResponse xmlns="urn:bicwsdl">
				<response>
				  <errorCode>%d</errorCode>
				</response>
			  </DeleteRequestResponse>
			</SOAP-ENV:Body>
		  </SOAP-ENV:Envelope>`, 800))
		w.Header().Set("Content-Type", "application/xml")
		w.Write(x)
	}
}

func (nn *server) logQuery(header, imsi, group, ip string, status int, err error) {
	errText := ""
	if err != nil {
		errText = err.Error()
	}
	loging(fmt.Sprintf("ip - %s %s imsi - %s, groupID - %s; status - %d error - %s", ip, header, imsi, group, status, errText), nn.logPath)

}

func (nn *server) addImsi(imsi, group string) (int, error) {

	rows, err := nn.conn.Query("select count(id) from grp_list WHERE id = $1", group)
	checkError("Cannot get grp_list", nn.logPath, err)
	if err != nil {
		loging("error query - "+err.Error(), nn.logPath)
		return 2000, errors.New("Cannot get grp_list")
	}
	var coun int
	for rows.Next() {
		rows.Scan(&coun)
	}
	if coun == 0 {
		return 1004, nil
	}
	rows, err = nn.conn.Query("select count(id) from grp_imsi WHERE list_id = $1  AND imsi = $2", group, imsi)
	checkError("Cannot get grp_imsi", nn.logPath, err)
	if err != nil {
		loging("error query - "+err.Error(), nn.logPath)
		return 2000, errors.New("Cannot get grp_imsi")
	}
	for rows.Next() {
		rows.Scan(&coun)
	}
	if coun != 0 {
		return 1002, nil
	}

	err = nn.insertImsi(imsi, group)
	if err != nil {
		return 2000, errors.New("Cannot insert into grp_imsi")
	}
	return 0, nil
}

func checkData(imsi, group string) (int, error) {
	if imsi == "" || group == "" {
		return 800, errors.New("imsi AND group cannot be empty")
	}
	r, _ := regexp.Compile("^[0-9]+$")
	if !r.Match([]byte(imsi)) || !r.Match([]byte(group)) {
		return 800, errors.New("imsi AND group should consist only from digits")
	}
	i, err := strconv.Atoi(group)
	if err != nil {
		return 800, errors.New(err.Error())
	}
	if i > 2147483647 {
		return 800, errors.New("groupID excceed the max value")
	}
	if len(imsi) != 15 {
		return 800, errors.New("IMSI should consist of 15 digits")
	}
	return 0, nil
}

func (nn *server) deleteImsi(imsi, group string) (int, error) {
	err := nn.deleteImsiTx(imsi, group)
	if err != nil {
		return 2000, errors.New("Cannot delete from grp_imsi")
	}
	return 0, nil
}

func (nn *server) Init() {
	pathConf := "/opt/svyazcom/etc/serverSOAP/"
	config := LoadConfiguration(pathConf + "soap.conf")
	connStr := "user=" + config.Database.User + " dbname=" + config.Database.Dbname + " host=" + config.Database.Host + " password=" + config.Database.Password + " port=" + config.Database.Port + " sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	checkError("error db connect", nn.logPath, err)
	nn.conn = db
	nn.logPath = config.logPath
	nn.conn.Query("SET search_path TO steer_web, steer, public")

	// fasthttp.ListenAndServe(port, nn.testProcessing)
	// http.HandleFunc("/", nn.testProcessing)
	http.HandleFunc("/", nn.processing)
	http.ListenAndServe(port, nil)
}

func (nn *server) insertImsi(imsi, group string) error {
	tx, _ := nn.conn.Begin()
	tx.Exec("select set_config('user.id', '17', false)")
	_, err := tx.Exec("select current_setting('user.id')")
	checkError("Cannot set param", nn.logPath, err)
	query := fmt.Sprintf("INSERT INTO grp_imsi (list_id, imsi) VALUES (%s, %s)", group, imsi)
	_, err = tx.Exec(query)
	checkError("Cannot insert into grp_imsi"+query, nn.logPath, err)
	if err != nil {
		loging("error query - "+err.Error(), nn.logPath)
		return errors.New("Cannot insert into grp_imsi")
	}
	err = tx.Commit()
	return err
}

func (nn *server) deleteImsiTx(imsi, group string) error {
	tx, _ := nn.conn.Begin()
	tx.Exec("select set_config('user.id', '17', false)")
	_, err := tx.Exec("select current_setting('user.id')")
	checkError("Cannot set param", nn.logPath, err)
	query := fmt.Sprintf("DELETE from grp_imsi WHERE list_id = %s AND imsi = '%s'", group, imsi)
	_, err = tx.Exec(query)
	checkError("Cannot delete from grp_imsi"+query, nn.logPath, err)
	if err != nil {
		loging("error query - "+err.Error(), nn.logPath)
		return errors.New("Cannot delete from grp_imsi")
	}
	err = tx.Commit()
	return err
}

func (nn *server) testProcessing(ctx *fasthttp.RequestCtx) {
	// set some headers and status code first
	ctx.SetContentType("foo/bar")
	ctx.SetStatusCode(fasthttp.StatusOK)

	// then write the first part of body
	fmt.Fprintf(ctx, "this is the first part of body\n")

	// then set more headers
	ctx.Response.Header.Set("Foo-Bar", "baz")

	// then write more body
	fmt.Fprintf(ctx, "this is the second part of body\n")

	// then override already written body
	ctx.SetBody([]byte("this is completely new body contents"))

	// then update status code
	ctx.SetStatusCode(fasthttp.StatusNotFound)

	// basically, anything may be updated many times before
	// returning from RequestHandler.
	//
	// Unlike net/http fasthttp doesn't put response to the wire until
	// returning from RequestHandler.
}
