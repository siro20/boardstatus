// models.test.go

package model

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	helper "github.com/siro20/boardstatus/pkg/helper"
	"gopkg.in/yaml.v2"
)

type Test struct {
	gorm.Model
	Name string    `json:"name" yaml:"name" gorm:"size:255" table_list:"Name"`
	Time time.Time `json:"time" yaml:"time" table_list:"Tested"`

	Checksum string `json:"checksum" yaml:"checksum" gorm:"size:255"`

	ReferenceExternalValidation string `json:"exeternal_ref" yaml:"exeternal_ref"` // e.g http://lava.test.invalid/test5

	// Test results in key value format
	FailedTest        []TestCase `json:"failed_tests" yaml:"failed_tests"`
	PassedTest        []TestCase `json:"passed_tests" yaml:"passed_tests"`
	SkippedTest       []TestCase `json:"skipped_tests" yaml:"skipped_tests"`
	FailedTestsCount  int        `json:"failed_tests_count" yaml:"failed_tests_count" table_list:"Failed tests #"`
	PassedTestsCount  int        `json:"passed_tests_count" yaml:"passed_tests_count" table_list:"Passed tests #"`
	SkippedTestsCount int        `json:"skipped_tests_count" yaml:"skipped_tests_count" table_list:"Skipped tests #"`

	// Collected data, raw ASCII, compressed
	FileKernelLog     []byte `json:"file_kernel_log" yaml:"file_kernel_log"`
	FileCMOS          []byte `json:"file_cmos" yaml:"file_cmos"`
	FileConfig        []byte `json:"file_config" yaml:"file_config"`
	FileBootlog       []byte `json:"file_bootlog" yaml:"file_bootlog"`
	FileTimestamps    []byte `json:"file_timestamps" yaml:"file_timestamps"`
	FilePayloadconfig []byte `json:"file_payload_config" yaml:"file_payload_config"`

	// Status
	Status        string `json:"status" yaml:"status" gorm:"size:255" table_list:"Status"`                         // one of PASS, FAIL, UNKN
	StatusComment string `json:"status_comment" yaml:"status_comment" gorm:"size:255" table_list:"Status comment"` // e.g. doesn't boot into OS
	Comment       string `json:"comment" yaml:"comment" gorm:"size:65536"`
	BoardID       uint   `json:"board_id" yaml:"board_id"`
}

type TestCase struct {
	gorm.Model
	Name   string `json:"name" yaml:"name" gorm:"size:65536"`
	Result string `json:"result" yaml:"result" gorm:"size:65536"`
	TestID uint   `json:"test_id" yaml:"test_id"`
}

// Return a list of all the boards
func getAllTests() ([]Test, error) {

	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Test{})

	// Read
	var testList []Test
	db.Find(&testList)

	return testList, nil
}

// Fetch an test based on the ID supplied
func getTestByID(id int) (*Test, error) {
	var t Test

	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	db.First(&t, id)
	return &t, nil
}

// Create a new board with the title and content provided
func createNewTest(data string) (*Test, error) {

	var t Test

	err := yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		fmt.Printf("Failed to unmarshal %v\n", err)
		return nil, err
	}

	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if !db.HasTable(&Test{}) {
		db.CreateTable(&Test{})
	}
	db.Create(&t)

	return &t, nil
}

func (t Test) render(c *gin.Context, showOnly bool) {
	// Check if the item ID is valid
	if ID, err := strconv.Atoi(c.Param("id")); err == nil {
		var Item string

		if !showOnly {
			if strings.Contains(c.Request.RequestURI, "/") {
				Item = strings.Split(c.Request.RequestURI, "/")[1]
			} else {
				Item = c.Request.RequestURI
			}
		}

		// Check if the board exists
		if test, err := getTestByID(ID); err == nil {
			RenderItems, err := getRenderItem(test)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
			} else {
				helper.Render(c, gin.H{
					"Name":        test.Name,
					"PostURL":     Item,
					"DisplayOnly": showOnly,
					"payload":     RenderItems}, "listitem.html")
			}
		} else {
			// If the item is not found, abort with an error
			c.AbortWithError(http.StatusNotFound, err)
		}
	}
}

func (t Test) RenderShow(c *gin.Context) {
	t.render(c, true)
}

func (t Test) RenderEdit(c *gin.Context) {
	t.render(c, false)
}

func (t Test) RenderAll(c *gin.Context) {
	var Item string

	if strings.Contains(c.Request.RequestURI, "/") {
		Item = strings.Split(c.Request.RequestURI, "/")[1]
	} else {
		Item = c.Request.RequestURI
	}

	// Check if the board exists
	if tests, err := getAllTests(); err == nil {
		Header, List, err := getRenderList(tests)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		} else {
			helper.Render(c, gin.H{
				"Name":        "All tests",
				"PostURL":     Item,
				"DisplayOnly": true,
				"payload":     List,
				"header":      Header}, "list.html")
		}
	} else {
		// If the item is not found, abort with an error
		c.AbortWithError(http.StatusNotFound, err)
	}
}
