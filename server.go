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
	_ "net/http/pprof"
	"time"
	//_ "golang.org/x/net/netutil"
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
	DeleteSubscriber ReqXml `xml:"DeleteSubscriber"`
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
		SoapUser string `json:"soapUser"`
		MaxConnection int `json:"maxConnection"`
	} `json:"database"`
	LogPath string `json:"logPath"`
}

type server struct {
	conn    *sql.DB
	logPath string
	prefImsi map[string]bool
	soapId    string
	sem chan bool
}



func (nn *server) processing(w http.ResponseWriter, r *http.Request) {
	nn.sem <- true
	defer func() { <-nn.sem }()
	xmlFile, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	loging(fmt.Sprintf("%s",xmlFile), nn.logPath)
	var q SoapXml
	_ = xmlFile
	xml.Unmarshal((xmlFile), &q)
	var requestResponse string
	var imsi_str string
	var group_id_str string
	var query string
	var typeQuary string
	time.Sleep(time.Millisecond * 50)
	if q.Body.AddReq.Imsi != "" && q.Body.AddReq.GroupId !="" {
		imsi_str = q.Body.AddReq.Imsi
		group_id_str = q.Body.AddReq.GroupId
		typeQuary="insert"
		query = fmt.Sprintf("INSERT INTO grp_imsi (list_id, imsi) VALUES (%s, %s)", group_id_str, imsi_str)
		requestResponse = "AddRequestResponse"
	} else if q.Body.UpdateReq.Imsi != "" && q.Body.UpdateReq.Imsi_replace!="" {
		imsi_str = q.Body.UpdateReq.Imsi
		group_id_str = q.Body.UpdateReq.Imsi_replace
		query = fmt.Sprintf("UPDATE grp_imsi SET imsi=%s WHERE imsi='%s'", group_id_str, imsi_str)
		typeQuary="update"
		requestResponse = "UpdateRequestResponse"
	} else if q.Body.DeleteReq.Imsi != "" && q.Body.DeleteReq.GroupId !="" {
		imsi_str = q.Body.DeleteReq.Imsi
		group_id_str = q.Body.DeleteReq.GroupId
		typeQuary="delete"
		query = fmt.Sprintf("DELETE from grp_imsi WHERE list_id = %s AND imsi = '%s'", group_id_str, imsi_str)
		requestResponse = "DeleteRequestResponse"
	} else if q.Body.DeleteSubscriber.Imsi != "" {
		imsi_str = q.Body.DeleteSubscriber.Imsi
		query = fmt.Sprintf("DELETE from grp_imsi WHERE imsi = '%s'", imsi_str)
		typeQuary = "deleteSubscriber"
		requestResponse = "DeleteSubscriber"
	}	else {
		typeQuary="errParam"
		loging(fmt.Sprintf("ip - s% incorrect query addImsi - %s deleteImsi - %s", r.RemoteAddr, q.Body.AddReq.Imsi, q.Body.DeleteReq.Imsi), nn.logPath)
		requestResponse = "DeleteRequestResponse"
	}
	status, err := nn.checkData(imsi_str, group_id_str,typeQuary)
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
	if status != 0 && status != 1002{
		w.WriteHeader(500)
	}
	loging(fmt.Sprintf("%s",x), nn.logPath)
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
	tx, err := nn.conn.Begin()
	if err != nil {
		loging("Can not open transaction - "+err.Error(), nn.logPath)
	}
	defer tx.Commit()
	userQuery := fmt.Sprintf("select set_config('user.id', '%s', false)", nn.soapId)
	tx.Exec(userQuery)
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

	if typeQuery=="update"{
		_, err := nn.checkData(group, group,typeQuery)
		if err!=nil{
			return 1004, errors.New("Wrong format IMSI")
		}
	}
	if typeQuery=="delete" {
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
	}
	if typeQuery=="delete" || typeQuery=="deleteSubscriber" || typeQuery=="update"{
		rows, err := nn.conn.Query("select count(id) from grp_imsi WHERE imsi = $1", imsi)
		checkError("Cannot get grp_imsi", nn.logPath, err)
		if err != nil {
			loging("error query - "+err.Error(), nn.logPath)
			return 2000, errors.New("Cannot get grp_imsi")
		}
		var coun int
		for rows.Next() {
			rows.Scan(&coun)
		}
		if coun == 0 {
			return 1002, nil
		}
	}
	err := nn.query_to_Db(cmd_query)
	if err != nil {
		fmt.Println(err.Error())
		if strings.Contains(err.Error(), "duplicate key value violates") {
			switch typeQuery {
			case "insert" :
				return 1002, errors.New("Can not add: data already exist")
			case "update" :
				return 1002, errors.New("Can not update: data already exist")
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

func (nn *server)  checkData(imsi, group,typeQuery string) (int, error) {
	if typeQuery =="errParam" {
		return 800, errors.New("imsi check error")
	}

	if len(imsi) != 15 {
		return 1004, errors.New("IMSI should consist of 15 digits")
	}

	if _, ok := nn.prefImsi[imsi[0:5]]; ! ok {
		return 1004, errors.New("imsi check error")
	}

	//r, _ := regexp.Compile("^[0-9]+$")

	/*if !r.Match([]byte(imsi)) || !r.Match([]byte(group)) {
		return 800, errors.New("imsi AND group should consist only from digits")
	}*/

	return 0, nil
}

func (nn *server) Init() {
	pathConf := "/opt/svyazcom/etc/"
	config := LoadConfiguration(pathConf + "soap.conf")
	fmt.Println(config.Database.MaxConnection)
	nn.sem = make(chan bool, config.Database.MaxConnection)
	connStr := "user=" + config.Database.User + " dbname=" + config.Database.Dbname + " host=" + config.Database.Host + " password=" + config.Database.Password + " port=" + config.Database.Port + " sslmode=disable"
	nn.soapId = config.Database.SoapUser
	db, err := sql.Open("postgres", connStr)
	checkError("error db connect", nn.logPath, err)
	nn.conn = db
	nn.logPath = config.LogPath
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
	http.HandleFunc("/", nn.processing)
	http.ListenAndServe(port, nil)
	//srv := &http.Server{
	//	Addr: "localhost:" + port,
	//}
	//ln, err := net.Listen("tcp", srv.Addr)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//defer ln.Close()
	//ln= netutil.LimitListener(ln, 20);
}

