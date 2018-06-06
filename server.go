package main

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "github.com/lib/pq"
	"strings"
	"regexp"
	_ "net/http/pprof"
)

const (
	port = ":63053"
)

type ReqXml struct {
	Xmlns   string `xml:"xmlns,attr"`
	Imsi    string `xml:"imsi"`
	GroupId string `xml:"groupId"`
	Imsi_replace string `xml:"newImsi"`
}

type BodyXml struct {
	AddReq    ReqXml `xml:"AddRequest"`
	DeleteReq ReqXml `xml:"DeleteRequest"`
	UpdateReq ReqXml `xml:"UpdateRequest"`
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
	prefImsi map[string]bool
}

func (nn *server) processing(w http.ResponseWriter, r *http.Request) {
	xmlFile, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var q SoapXml
	_ = xmlFile
	xml.Unmarshal((xmlFile), &q)
	var requestResponse string
	var imsi_str string
	var group_id_str string
	var query string
	var typeQuary string
	if q.Body.AddReq.Imsi != "" && q.Body.DeleteReq.Imsi == "" {
		imsi_str = q.Body.AddReq.Imsi
		group_id_str = q.Body.AddReq.GroupId
		typeQuary="insert"
		query = fmt.Sprintf("INSERT INTO grp_imsi (list_id, imsi) VALUES (%s, %s)", group_id_str, imsi_str)
		requestResponse = "AddRequestResponse"
	} else if q.Body.UpdateReq.Imsi != "" {
		imsi_str = q.Body.UpdateReq.Imsi
		group_id_str = q.Body.UpdateReq.Imsi_replace
		query = fmt.Sprintf("UPDATE grp_imsi SET imsi=%s WHERE imsi='%s'", group_id_str, imsi_str)
		typeQuary="update"
		requestResponse = "UpdateRequestResponse"
	} else if q.Body.AddReq.Imsi == "" && q.Body.DeleteReq.Imsi != "" {
		imsi_str = q.Body.DeleteReq.Imsi
		group_id_str = q.Body.DeleteReq.GroupId
		typeQuary="delete"
		query = fmt.Sprintf("DELETE from grp_imsi WHERE list_id = %s AND imsi = '%s'", group_id_str, imsi_str)
		requestResponse = "DeleteRequestResponse"
	} else {
		loging(fmt.Sprintf("ip - s% incorrect query addImsi - %s deleteImsi - %s", r.RemoteAddr, q.Body.AddReq.Imsi, q.Body.DeleteReq.Imsi), nn.logPath)
		requestResponse = "DeleteRequestResponse"
	}
	status, err := nn.checkData(imsi_str, group_id_str)
	if err == nil {
		status, err = nn.doImsi(imsi_str, group_id_str, query, typeQuary)
		nn.logQuery("add query", imsi_str, group_id_str, r.RemoteAddr, status, err)
	} else {
		nn.logQuery("add query", imsi_str, group_id_str, r.RemoteAddr, status, err)
	}
	x := []byte(fmt.Sprintf(`<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
		<SOAP-ENV:Body>
		  <%s xmlns="urn:bicwsdl">
			<response>
			  <errorCode>%d</errorCode>
			</response>
		  </%s>
		</SOAP-ENV:Body>
	  </SOAP-ENV:Envelope>`, requestResponse, status, requestResponse))
	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}


func (nn *server) logQuery(header, imsi, group, ip string, status int, err error) {
	errText := ""
	if err != nil {
		errText = err.Error()
	}
	loging(fmt.Sprintf("ip - %s %s imsi - %s, groupID - %s; status - %d error - %s", ip, header, imsi, group, status, errText), nn.logPath)

}

func (nn *server) query_to_Db(query string) error {
	tx, _ := nn.conn.Begin()
	tx.Exec("select set_config('user.id', '17', false)")
	_, err := tx.Exec("select current_setting('user.id')")
	checkError("Cannot set param", nn.logPath, err)
	_, err = tx.Exec(query)
	checkError("Cannot update grp_imsi"+query, nn.logPath, err)
	if err != nil {
		loging("error query - "+err.Error(), nn.logPath)
		return err
	}
	defer tx.Commit()
	return nil
}

func (nn *server) doImsi(imsi, group, cmd_query, typeQuery string) (int, error) {

	err := nn.query_to_Db(cmd_query)
	if err != nil {
		fmt.Println(err.Error())
		if strings.Contains(err.Error(), "duplicate key value violates") {
			switch typeQuery {
			case "insert" :
				return 1002, errors.New("Can not add: data already exist")
			case "update" :
				return 1004, errors.New("Can not update: data already exist")
			case "delete" :
				return 1002, errors.New("Can not delete data")
			}
		}
		if strings.Contains(err.Error(), `insert or update on table "grp_imsi" violates`) {
			return 1004, errors.New("Can not add: group not exist")
		}
		return 2000, errors.New("Unexpected error")
	}
	//defer rows.Close()
	return 0, nil
}

func (nn *server)  checkData(imsi, group string) (int, error) {
	if imsi == "" || group == "" {
		return 800, errors.New("imsi AND group cannot be empty")
	}

	if len(imsi) != 15 {
		return 800, errors.New("IMSI should consist of 15 digits")
	}

	if _, ok := nn.prefImsi[imsi[0:5]]; ! ok {
		return 800, errors.New("imsi check error")
	}
	r, _ := regexp.Compile("^[0-9]+$")

	if !r.Match([]byte(imsi)) || !r.Match([]byte(group)) {
		return 800, errors.New("imsi AND group should consist only from digits")
	}

	return 0, nil
}

func (nn *server) Init() {
	pathConf := "/opt/svyazcom/etc/serverSOAP/"
	config := LoadConfiguration(pathConf + "soap.conf")
	connStr := "user=" + config.Database.User + " dbname=" + config.Database.Dbname + " host=" + config.Database.Host + " password=" + config.Database.Password + " port=" + config.Database.Port + " sslmode=disable"
	fmt.Print(connStr)
	db, err := sql.Open("postgres", connStr)
	checkError("error db connect", nn.logPath, err)
	nn.conn = db
	nn.logPath = config.logPath
	nn.conn.Exec("SET search_path TO steer_web, steer, public")
	rows := nn.conn.QueryRow("select param_value from com_param WHERE code = 'HPMN_MCC_MNC'")
	var param string
	checkError("Cannot get com_param", nn.logPath, err)
	nn.prefImsi = make(map[string]bool)
	rows.Scan(&param)
	paramWithoutSpaces := strings.Replace(param, " ", "", -1)
	s := strings.Split(paramWithoutSpaces, ",")
	for i := 0; i < len(s); i++ {
		if len(s[i]) >= 5 && len(s[i]) <= 6 {
			nn.prefImsi[s[i]] = true
		}
	}
	for v, k := range nn.prefImsi {
		fmt.Println(v, k)
	}
	loging("ServerSo start ", nn.logPath)
	// fasthttp.ListenAndServe(port, nn.testProcessing)
	// http.HandleFunc("/", nn.testProcessing)

	http.HandleFunc("/", nn.processing)
	http.ListenAndServe(port, nil)
}

/*func (nn *server) testProcessing(ctx *fasthttp.RequestCtx) {
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
}*/
