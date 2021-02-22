package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	asciiUppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	GSUrlRegexp    = "http(s?):\\/\\/docs\\.google\\.com\\/spreadsheets\\/d\\/(.+)\\/(.*)"
)

type ParsedData struct {
	GSService    *sheets.Service
	ParsedFields []ParsedFieldData
	RangeStr     string
	Range
	SheetId string
}

type Data struct {
	URL        string
	ConfigPath string
	SheetName  string
	Fields     []FieldData
}

type Range struct {
	x1 int
	y1 int
	x2 int
	y2 int
}

type FieldData struct {
	Field   string
	Content string
}

type ParsedFieldData struct {
	Row     int
	Col     int
	Content string
}

func (d *ParsedData) DoReq(ch chan int) {
	var values [][]interface{}
	var row []interface{}

	for i := 0; i < d.x2 - d.x1 + 1; i++ {
		row = append(row, nil)
	}

	for i := 0; i < d.y2 - d.y1 + 1; i++ {
		tmp := make([]interface{}, d.x2 - d.x1 + 1)
		copy(tmp, row)
		values = append(values, tmp)
	}

	for _,v := range d.ParsedFields{
		values[v.Col-d.y1][v.Row-d.x1] = v.Content
	}

	log.Println(values, len(values))

	doReq(values, d.RangeStr, d.GSService, d.SheetId, ch)
}

func (d *Data) ValidateAndParse() (*ParsedData, error) {
	client := ServiceAccount(d.ConfigPath) // Please set the json file of Service account.
	srv, err := sheets.New(client)
	if err != nil {
		return &ParsedData{}, err
	}

	var pfd []ParsedFieldData
	for _, v := range d.Fields {
		p, err := v.Parse()
		if err != nil {
			return &ParsedData{}, err
		}
		pfd = append(pfd, p)
	}

	rngs, rng, err := parseRange(pfd)
	if err != nil {
		return &ParsedData{}, err
	}

	shtId, err := ParseSheetId(d.URL)
	if err != nil {
		return &ParsedData{}, err
	}

	return &ParsedData{
		GSService:    srv,
		ParsedFields: pfd,
		RangeStr:     d.SheetName + "!" + rngs,
		Range:        rng,
		SheetId:      shtId,
	}, nil
}

func parseRange(pfd []ParsedFieldData) (string, Range, error) {
	a1 := pfd[0].Row
	a2 := pfd[0].Col
	b1 := pfd[0].Row
	b2 := pfd[0].Col
	for i := 1; i < len(pfd); i++ {
		if a1 > pfd[i].Row {
			a1 = pfd[i].Row
		}
		if a2 > pfd[i].Col {
			a2 = pfd[i].Col
		}
		if b1 < pfd[i].Row {
			b1 = pfd[i].Row
		}
		if b2 < pfd[i].Col {
			b2 = pfd[i].Col
		}
	}
	if a1 > len(asciiUppercase) || b1 > len(asciiUppercase) {
		return "", Range{}, fmt.Errorf("too big value")
	}
	return fmt.Sprintf("%v%v:%v%v", string(asciiUppercase[a1]),
		a2, string(asciiUppercase[b1]), b2), Range{a1, a2, b1, b2}, nil
}

func ParseSheetId(url string) (string, error) {
	re := regexp.MustCompile(GSUrlRegexp)
	s := re.FindStringSubmatch(url)
	if len(s) < 3 {
		return "", fmt.Errorf("incorrect url")
	}
	return s[2], nil
}

func (fd FieldData) Parse() (ParsedFieldData, error) {
	if fd.Field == "" || fd.Content == "" {
		return ParsedFieldData{}, fmt.Errorf("there is empty field")
	}

	p1 := string(fd.Field[0])
	p2 := fd.Field[1:]

	if !strings.Contains(asciiUppercase, p1) {
		return ParsedFieldData{}, fmt.Errorf("incorrect field")
	}
	num1 := strings.Index(asciiUppercase, p1)

	num2, err := strconv.Atoi(p2)
	if err != nil {
		return ParsedFieldData{}, err
	}

	return ParsedFieldData{num1, num2, fd.Content}, err
}

func ServiceAccount(credentialFile string) *http.Client {
	b, err := ioutil.ReadFile(credentialFile)
	if err != nil {
		log.Fatal(err)
	}
	var c = struct {
		Email      string `json:"client_email"`
		PrivateKey string `json:"private_key"`
	}{}
	json.Unmarshal(b, &c)
	config := &jwt.Config{
		Email:      c.Email,
		PrivateKey: []byte(c.PrivateKey),
		Scopes: []string{
			docs.DocumentsScope,
			sheets.SpreadsheetsScope,
		},
		TokenURL: google.JWTTokenURL,
	}
	client := config.Client(oauth2.NoContext)
	return client
}

func doReq(values [][]interface{}, rng string, srv *sheets.Service, spreadsheetID string, ch chan int) {
	for {
		select {
		case <-ch:
			log.Println("TERMINATED")
			return
		default:
		}
		sheetValues, err := srv.Spreadsheets.Values.Get(spreadsheetID, rng).Do()
		if err != nil {
			log.Println(err.Error())
			time.Sleep(time.Second)
			continue
		}
		if sheetValues == nil {
			log.Println(rng + " NIL VALUES")
			time.Sleep(time.Second)
			continue
		}
		vls := sheetValues.Values
		if reflect.DeepEqual(values, vls) {
			log.Println(rng + " EQUALS")
			time.Sleep(time.Second)
			continue
		}
		ctx := context.Background()
		rb := &sheets.BatchUpdateValuesRequest{
			ValueInputOption: "USER_ENTERED",
		}
		rb.Data = append(rb.Data, &sheets.ValueRange{
			Range:  rng,
			Values: values,
		})
		_, err = srv.Spreadsheets.Values.BatchUpdate(spreadsheetID, rb).Context(ctx).Do()
		if err == nil {
			log.Println(rng + " INSERT")
			time.Sleep(time.Second)
			//break
		} else {
			log.Println(err.Error())
			time.Sleep(time.Second)
		}
	}
}
