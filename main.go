package main

import (
	"code.google.com/p/go-charset/charset"
  	_ "code.google.com/p/go-charset/data"
  	"crypto/md5"
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

import _ "github.com/denisenkom/go-mssqldb"
import "database/sql"

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

	SqlQueres []*SqlProfile `json:"sql_queres"`
}

type SqlProfile struct {
	Hash string `json:"hash_sql_format"`
	DBName string `json:"db_name"`
	SeqNumber int `json:"sequence_number"`
	SqlFormat string `json:"sql_format"`
	SqlParams string `json:"sql_params"`
	SqlQuery string `json:"sql_query"`

	DataColumnt []string `json:"columns"`
	DataDB []map[string]interface{} `json:"data_db"`
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

	log.Printf("INFO: Parse xml...")

	for _sqlSeqNumber, event := range events.Events {
		if event.ObjectName == "sp_prepexec" {
			sql := new(SqlProfile)
			
			sql.DBName = event.DatabaseName
			sql.SeqNumber = _sqlSeqNumber
			
			// Находим SQL и параметры
			// Кол-во параметров в SQL должно равняться кол-ву параметров
			// reSQL := regexp.MustCompile("'(select.+?)'")
			reSQL := regexp.MustCompile(`N'(select.+?)'`)
			findSQL := reSQL.FindStringSubmatch(event.Value)

			if len(findSQL) > 0 {
				sql.SqlFormat = strings.Trim(findSQL[1], " ");
				sql.Hash = fmt.Sprintf("%x", md5.Sum([]byte(sql.SqlFormat)))

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

					paramReplase := regexp.MustCompile(fmt.Sprintf("@P%d([\\D]|$)", index))
					// sql.SqlQuery = paramReplase.ReplaceAllString(sql.SqlQuery, value + " ")
					sql.SqlQuery = paramReplase.ReplaceAllStringFunc(sql.SqlQuery, func (str string) string {
						return strings.Replace(str, fmt.Sprintf("@P%d", index), value, -1)
						})

					// sql.SqlQuery = strings.Replace(sql.SqlQuery, fmt.Sprintf("@P%d", index), value, -1)
				}
			}

			page.SqlQueres = append(page.SqlQueres, sql)
		}
	}

	log.Printf("INFO: Extracted data from DB... ")

	var port = 10019
	var server = "192.168.1.37"
	var user = "sa_test"
	var password = "123qwe"

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d", server, user, password, port)

	log.Println("DEBUG: Connstring", connString)

	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()

	for _, _sql := range page.SqlQueres {
		conn.Exec(fmt.Sprintf("USE %s", _sql.DBName))

		rows, err := conn.Query(_sql.SqlQuery)

		if err != nil {
			log.Printf("ERROR: Sql query '%s'. %s\n", _sql.SqlQuery, err)
			continue
		}

		if rows.Err() != nil {
			log.Printf("ERROR: Rows %s", rows.Err())
			continue
		}

		cols, _ := rows.Columns()
		_sql.DataColumnt = cols

		rawResult := make([][]byte, len(cols))
    	result := []map[string]interface{}{}

    	// http://stackoverflow.com/questions/14477941/read-select-columns-into-string-in-go
    	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	    for i, _ := range rawResult {
	        dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	    }

		for rows.Next() {
			row := make(map[string]interface{})
			err = rows.Scan(dest...)

			if err != nil {
				fmt.Println("Failed to scan row", err)
				continue
			}

			for i, colName := range cols {
				raw := rawResult[i]

	            if raw == nil {
	                row[colName] = "\\N"
	            } else {
	                row[colName] = string(raw)
	            }
	        }

	        result = append(result, row);
			// return
		}

		rows.Close()

		// result := make(map[string]interface{})

		// log.Printf("%v", values)

		// for _index, _columnName := range columns {
			// log.Printf("%v - %v", _index, _columnName)
			// result[_columnName] = values[_index]
		// }

		_sql.DataDB = result
	}

	log.Printf("INFO: Saved...")

	pageBytes, _ := json.Marshal(page)
	ioutil.WriteFile(fmt.Sprintf("%s.profiles.json", filePath), pageBytes, 0644)

	for _, _sql := range page.SqlQueres {
		sqlTexts = append(sqlTexts, "--")
		sqlTexts = append(sqlTexts, fmt.Sprintf("-- Hash: %s", _sql.Hash))
		sqlTexts = append(sqlTexts, fmt.Sprintf("-- Seq: #%d", _sql.SeqNumber))
		sqlTexts = append(sqlTexts, fmt.Sprintf("-- DB: %s", _sql.DBName))
		sqlTexts = append(sqlTexts, fmt.Sprintf("-- SQLFormat: %s", _sql.SqlFormat))
		sqlTexts = append(sqlTexts, fmt.Sprintf("-- SQLParams: %s", _sql.SqlParams))
		sqlTexts = append(sqlTexts, "-- Description: ")
		sqlTexts = append(sqlTexts, "-- ")
		sqlTexts = append(sqlTexts, "-- ")
		sqlTexts = append(sqlTexts, fmt.Sprintf("USE %s;", _sql.DBName))
		sqlTexts = append(sqlTexts, _sql.SqlQuery + ";")
		sqlTexts = append(sqlTexts, "")
		sqlTexts = append(sqlTexts, "")

		sqlTexts = append(sqlTexts, "-- ")
		sqlTexts = append(sqlTexts, fmt.Sprintf("-- Total rows: %d", len(_sql.DataDB)))
		sqlTexts = append(sqlTexts, "-- ")
		sqlTexts = append(sqlTexts, "-- Columns: " + strings.Join(_sql.DataColumnt, "\t"))
		sqlTexts = append(sqlTexts, "-- ")

		for _i, _row := range _sql.DataDB {

			rowValues := []string{}

			for _, _cName := range _sql.DataColumnt {
				rowValues = append(rowValues, _row[_cName].(string))
			}

			sqlTexts = append(sqlTexts, "-- " + strings.Join(rowValues, "\t"))

			if _i > 100 {
				sqlTexts = append(sqlTexts, fmt.Sprintf("-- ... total count rows %d", len(_sql.DataDB)))
				sqlTexts = append(sqlTexts, "--")
				break
			}
		}
		sqlTexts = append(sqlTexts, "-- ")
	}

	ioutil.WriteFile(fmt.Sprintf("%s.profiles.sql", filePath), []byte(strings.Join(sqlTexts, "\n")), 0644)

	log.Println("INFO: Bye-bye...")
}
