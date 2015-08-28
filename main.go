package main

import (
	"code.google.com/p/go-charset/charset"
  	_ "code.google.com/p/go-charset/data"
  	"regexp"
	"encoding/xml"
	"encoding/json"
	"bytes"
	"log"
	"io/ioutil"
	"flag"
	"strings"
	"fmt"
)

type Event struct {
	ObjectName string `xml:"ObjectName,attr"`
	DatabaseName string `xml:"DatabaseName,attr"`
	Value  string `xml:",chardata"`
}

type Events struct {
	XMLName xml.Name  `xml:events`
  	Events   []Event `xml:"event"`

}

type PageProfile struct {
	PageName string `json:"page_name"`

	SqlQueres []SqlProfile `json:"sql_queres"`
}

type SqlProfile struct {
	SeqNumber int `json:"sequence_number"`
	SqlFormat string `json:"sql_format"`
	SqlParams string `json:"sql_params"`
	SqlQuery string `json:"sql_query"`
}

var filePath string
var pageName string

func init() {
	flag.StringVar(&filePath, "f", "", "path file")
	flag.StringVar(&pageName, "title", "", "page name")
}

func main() {
	flag.Parse()

	log.Printf("INFO: Open file '%s'", filePath)

	// Открыли файл
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Выбрали все SQL запросы с параметрами
	var events = Events{}
	var page = new(PageProfile)
	page.PageName = pageName
	var sqlTexts = []string{
		fmt.Sprintf("-- Page name: '%s'", pageName),
	}
	
	reader := bytes.NewReader(data)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReader
	err = decoder.Decode(&events)
	if err != nil {
		log.Printf("ERROR: Decode xml. %s", err)
		return
	}
	// Записали все SQL запросы с вставленными параметрами

	for _sqlSeqNumber, event := range events.Events {
		if event.ObjectName == "sp_prepexec" {
			sql := SqlProfile{}
			sql.SeqNumber = _sqlSeqNumber
			
			// Находим SQL и параметры
			// Кол-во параметров в SQL должно равняться кол-ву параметров
			// reSQL := regexp.MustCompile("'(select.+?)'")
			reSQL := regexp.MustCompile(`N'(select.+?)'`)
			findSQL := reSQL.FindStringSubmatch(event.Value)

			if len(findSQL) > 0 {
				sql.SqlFormat = strings.Trim(findSQL[1], " ");

				// reRemoveParame := regexp.MustCompile("@P\\d+");
				// reRemoveParame.ReplaceAllString(src, )

				reParams := regexp.MustCompile(`N'select.+?',(.*)`)
				findParams := reParams.FindStringSubmatch(event.Value)

				if len(findParams) == 0 {
					continue
				}

				sql.SqlParams = strings.Trim(findParams[1], " ")

				paramsArray := strings.Split(sql.SqlParams, ",")

				sql.SqlQuery = sql.SqlFormat

				for index, param := range paramsArray {

					var value string

					switch {
					case string(param[0]) == "N":
						value = param[1:len(param)]
					default: 
						value = param
					}

					paramReplase := regexp.MustCompile(fmt.Sprintf("(@P%d)(?:[^0-9])??", index))
					// sql.SqlQuery = paramReplase.ReplaceAllString(sql.SqlQuery, value + " ")
					sql.SqlQuery = paramReplase.ReplaceAllStringFunc(sql.SqlQuery, func (str string) string {
						return strings.Replace(str, fmt.Sprintf("@P%d", index), value, -1)
						})

					// sql.SqlQuery = strings.Replace(sql.SqlQuery, fmt.Sprintf("@P%d", index), value, -1)
				}
			}

			page.SqlQueres = append(page.SqlQueres, sql)
			sqlTexts = append(sqlTexts, fmt.Sprintf("-- Sq: #%d", _sqlSeqNumber))
			sqlTexts = append(sqlTexts, fmt.Sprintf("-- SQLFormat: %s", sql.SqlFormat))
			sqlTexts = append(sqlTexts, fmt.Sprintf("-- SQLParams: %s", sql.SqlParams))
			sqlTexts = append(sqlTexts, sql.SqlQuery)
			sqlTexts = append(sqlTexts, "")
			sqlTexts = append(sqlTexts, "")
		}
	}

	pageBytes, _ := json.Marshal(page)

	ioutil.WriteFile(fmt.Sprintf("%s.sql_profiles.json", filePath), pageBytes, 0644)
	ioutil.WriteFile(fmt.Sprintf("%s.sql_profiles.sql", filePath), []byte(strings.Join(sqlTexts, "\n")), 0644)
	log.Println("Done")
	
}
