package main

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "github.com/lib/pq"
	"regexp"
	"strings"
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

func (nn *server) checkImsi(imsi string){
	rows, err := nn.conn.Query("select param_value from com_param WHERE code = 'HPMN_MCC_MNC'")
	var param string
	checkError("Cannot get com_param", nn.logPath, err)
	if err != nil {
		loging("error query - "+err.Error(), nn.logPath)
		//return 2000, errors.New("Cannot get grp_imsi")
	}
	for rows.Next() {
		rows.Scan(&param)
		fmt.Println(strings.Split(param,","))
	}
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
	check_row_exist := checkInsert
	if q.Body.AddReq.Imsi != "" && q.Body.DeleteReq.Imsi == "" {
		imsi_str = q.Body.AddReq.Imsi
		group_id_str = q.Body.AddReq.GroupId
		check_row_exist = checkInsert
		query = fmt.Sprintf("INSERT INTO grp_imsi (list_id, imsi) VALUES (%s, %s)", group_id_str, imsi_str)
		requestResponse = "AddRequestResponse"
	} else if q.Body.UpdateReq.Imsi != "" {
		imsi_str = q.Body.UpdateReq.Imsi
		group_id_str = q.Body.UpdateReq.Imsi_replace
		query = fmt.Sprintf("UPDATE grp_imsi SET imsi=%s WHERE imsi='%s'", group_id_str, imsi_str)
		check_row_exist = checkUpdate//update
		requestResponse = "UpdateRequestResponse"
	} else if q.Body.AddReq.Imsi == "" && q.Body.DeleteReq.Imsi != "" {
		imsi_str = q.Body.DeleteReq.Imsi
		group_id_str = q.Body.DeleteReq.GroupId
		check_row_exist = checkDelete//delete
		query = fmt.Sprintf("DELETE from grp_imsi WHERE list_id = %s AND imsi = '%s'", group_id_str, imsi_str)
		requestResponse = "DeleteRequestResponse"
	} else {
		loging(fmt.Sprintf("ip - s% incorrect query addImsi - %s deleteImsi - %s", r.RemoteAddr, q.Body.AddReq.Imsi, q.Body.DeleteReq.Imsi), nn.logPath)
		requestResponse = "DeleteRequestResponse"
	}
	status, err := nn.checkData(imsi_str, group_id_str)
	nn.checkImsi(imsi_str)
	if err == nil {
		status, err = nn.doImsi(imsi_str, group_id_str, query, check_row_exist)
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

func checkInsert(coun int, conn *sql.DB, param string ) (int, string) {
	if coun != 0{
		return 1002, ""
	}
	_, err := conn.Query("select count(id) from grp_list WHERE id = $1", param)
	if err != nil {
		return 2000, "Cannot get grp_imsi"
	}
	return 0,""
}

func checkUpdate(coun int, conn *sql.DB, param string ) (int, string) {
	if coun == 0{
		return 1004, ""
	}
	//_, err := conn.Query("select count(id) from grp_imsi WHERE imsi = $1", param)
	//if err != nil {
	//	return 2000, "Cannot get grp_imsi"
	//}
	return 0,""
}

func checkDelete(coun int, conn *sql.DB, param string ) (int, string) {
	if coun == 0{
		return 1002, ""
	}
	_, err := conn.Query("select count(id) from grp_list WHERE id = $1", param)
	if err != nil {
		return 2000, "Cannot get grp_imsi"
	}
	return 0,""
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
	fmt.Println(query)
	checkError("Cannot set param", nn.logPath, err)
	_, err = tx.Exec(query)
	checkError("Сould not complete request"+query, nn.logPath, err)
	fmt.Println(err)
	if err != nil {
		fmt.Println("error query - " + err.Error())
		loging("error query - "+err.Error(), nn.logPath)
		return err
	}
	err = tx.Commit()
	return err
}

func (nn *server) doImsi(imsi, group, cmd_query string, check_exist_row func(coun int, conn *sql.DB,param string) (int, string)) (int, error) {
	var coun int
	rows, err := nn.conn.Query("select count(id) from grp_imsi WHERE imsi = $1", imsi)
	checkError("Cannot get grp_imsi", nn.logPath, err)
	if err != nil {
		loging("error query - "+err.Error(), nn.logPath)
		return 2000, errors.New("Cannot get grp_imsi")
	}
	for rows.Next() {
		rows.Scan(&coun)
	}
	if code, errString := check_exist_row(coun, nn.conn, group); code!=0 {
		return code, errors.New(errString)
	}
	err = nn.query_to_Db(cmd_query)
	if err != nil {
		if strings.Contains(err.Error(),"duplicate key value violates unique constraint"){
			return 1004, errors.New("Preset imsi already exists")
		}
		return 2000, errors.New("Cannot insert grp_imsi")
	}
	return 0, nil
}

func (nn *server) checkData(imsi, group string) (int, error) {
	if imsi == "" || group == "" {
		return 800, errors.New("imsi AND group cannot be empty")
	}

	if _, ok := nn.prefImsi[imsi[0:5]]; ! ok {
		return 800, errors.New("imsi check error")
	}
	r, _ := regexp.Compile("^[0-9]+$")

	if !r.Match([]byte(imsi)) || !r.Match([]byte(group)) {
		return 800, errors.New("imsi AND group should consist only from digits")
	}
	/*i, err := strconv.Atoi(group)
	if err != nil {
		return 800, errors.New(err.Error())
	}
	if i > 2147483647 {
		return 800, errors.New("groupID excceed the max value")
	}*/
	if len(imsi) != 15 {
		return 800, errors.New("IMSI should consist of 15 digits")
	}
	return 0, nil
}

func (nn *server) Init() {
	//pathConf := "/opt/svyazcom/etc/serverSOAP/"
	pathConf := "/home/maxim/go/src/SOAP/serverSOAP/"
	config := LoadConfiguration(pathConf + "soap.conf")
	connStr := "user=" + config.Database.User + " dbname=" + config.Database.Dbname + " host=" + config.Database.Host + " password=" + config.Database.Password + " port=" + config.Database.Port + " sslmode=disable"
	fmt.Print(connStr)
	db, err := sql.Open("postgres", connStr)
	checkError("error db connect", nn.logPath, err)
	nn.conn = db
	nn.logPath = config.logPath
	nn.conn.Query("SET search_path TO steer_web, steer, public")
	rows, err := nn.conn.Query("select param_value from com_param WHERE code = 'HPMN_MCC_MNC'")
	var param string
	checkError("Cannot get com_param", nn.logPath, err)
	if err != nil {
		loging("error query - "+err.Error(), nn.logPath)
		//return 2000, errors.New("Cannot get grp_imsi")
	}
	nn.prefImsi = make(map[string]bool)
	for rows.Next() {
		rows.Scan(&param)
		s:=strings.Split(param,",")
		for i:=0;i<len(s);i++{
			nn.prefImsi[s[i]]=true;
		}
		//fmt.Println(param)
	}
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
