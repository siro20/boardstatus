// models.board.go

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

type Board struct {
	gorm.Model
	Name         string `json:"name" yaml:"name" gorm:"size:255" table_title:"Name" table_default:"" table_descr:"Unique board name" table_list:"Name"`
	Manufacturer string `json:"manufacturer" yaml:"manufacturer" gorm:"size:255" table_default:"Emulation" table_descr:"The mainboard manufacturer, as in SMBIOS Type 1 'Manufacturer'" table_list:"Manufacturer"` // SMBIOS Type 1
	ProductName  string `json:"product_name" yaml:"product_name" gorm:"size:255" table_default:"Standard PC" table_descr:"The mainboard name, as in SMBIOS Type 1 'Product Name'"`                                 // SMBIOS Type 1
	Version      string `json:"version" yaml:"version" gorm:"size:255" table_default:"pc-i440fx" table_descr:"The mainboard name, as in SMBIOS Type 1 'Version'"`                                                  // SMBIOS Type 1
	Sku          string `json:"sku" yaml:"sku" gorm:"size:255"  table_default:"" table_descr:"The mainboard sku, as in SMBIOS Type 1 'Sku Number'"`                                                                // SMBIOS Type 1
	Family       string `json:"family" yaml:"family" gorm:"size:255"  table_default:"" table_descr:"The mainboard family, as in SMBIOS Type 1 'Family'"`                                                           // SMBIOS Type 1

	// Enclosure
	BoardType string `json:"board_type" yaml:"board_type" gorm:"size:255" table_title:"Enclosure" table_default:"ATX" table_descr:"Board type"` // SMBIOS Type 2
	Enclosure string `json:"enclosure" yaml:"enclosure" gorm:"size:255" table_default:"Pizzabox" table_descr:"Enclosure"`                       // SMBIOS Type 2

	// Integrated components
	NorthbridgeName       string `json:"northbridge_name" yaml:"northbridge_name" gorm:"size:255"  table_title:"Integrated components" table_default:"" table_descr:"Name of the nortbridge"`
	SouthbridgeName       string `json:"southbridge_name" yaml:"southbridge_name"  gorm:"size:255" table_default:"" table_descr:"Name of the southbridge"`
	SuperIOName           string `json:"superio_name" yaml:"superio_name" gorm:"size:255" table_default:"" table_descr:"Name of the SuperI/O"`  // Leave empty if not present
	ECName                string `json:"ec_name" yaml:"ec_name" gorm:"size:255" table_default:"" table_descr:"Name of the Embedded Controller"` // Leave empty if not present
	FlashICName           string `json:"flash_ic_name" yaml:"flash_ic_name"  gorm:"size:255" table_default:"" table_descr:"Name of the FlashIC"`
	FlashICCapacityInByte int    `json:"flash_ic_capacity_byte" yaml:"flash_ic_capacity_byte" table_default:"" table_descr:"Size of flash IC"`
	// Processor
	ProcessorManufacturer string `json:"processor_manufacturer" yaml:"processor_manufacturer" gorm:"size:255" table_title:"Processor" table_default:"" table_descr:""` // SMBIOS Type 4
	ProcessorFamily       string `json:"processor_family" yaml:"processor_family" gorm:"size:255" table_default:"" table_descr:""`                                     // SMBIOS Type 4
	ProcessorType         string `json:"processor_type" yaml:"processor_type" gorm:"size:255" table_default:"" table_descr:""`                                         // SMBIOS Type 4
	ProcessorSocket       string `json:"processor_socket" yaml:"processor_socket" gorm:"size:255" table_default:"" table_descr:""`                                     // SMBIOS Type 4
	ProcessorSocketCount  int    `json:"processor_socket_count" yaml:"processor_socket_count" table_default:"" table_descr:""`
	// Memory
	MaxMemorySlots         int `json:"memory_slots" yaml:"memory_slots" table_title:"Memory" table_default:"" table_descr:""`    // SMBIOS Type 16
	MaxSupportedMemoryInGB int `json:"max_supported_memory_gib" yaml:"max_supported_memory_gib" table_default:"" table_descr:""` // SMBIOS Type 16
	SolderedDownMemoryInGB int `json:"soldered_down_memory_gib" yaml:"soldered_down_memory_gib" table_default:"" table_descr:""` // SMBIOS Type 17
	// Software
	FirstCommit        string    `json:"first_commit" yaml:"first_commit" gorm:"size:255" table_title:"Software" table_default:"" table_descr:""`                   // When added tp master
	LastCommit         string    `json:"last_commit" yaml:"last_commit" gorm:"size:255" table_default:"" table_descr:""`                                            // When removed from master
	LastFailedCommit   string    `json:"last_failed_commit" yaml:"last_failed_commit" gorm:"size:255" table_default:"" table_descr:"" table_list:"Last bad commit"` // The last bad commit
	LastGoodCommit     string    `json:"last_good_commit" yaml:"last_good_commit" gorm:"size:255" table_default:"" table_descr:"" table_list:"Last good commit"`    // The last good commit
	TestedCommit       string    `json:"tested_commit" yaml:"tested_commit" gorm:"size:255" table_default:"" table_descr:""`                                        // The last tested commit
	NameOfTestedCommit string    `json:"name_of_commit" yaml:"name_of_commit" gorm:"size:255" table_default:"" table_descr:""`                                      // e.g. coreboot-4.12-123-dirty
	TestedCommitTime   time.Time `json:"tested_commit_time" yaml:"tested_commit_time" table_default:"" table_descr:""`                                              // When the last tested commit was uploaded

	// Status
	Status        string `json:"status" yaml:"status" gorm:"size:255" table_title:"Status" table_default:"" table_descr:"" table_list:"Status"`   // one of PASS, FAIL, UNKN
	StatusComment string `json:"status_comment" yaml:"status_comment" gorm:"size:255" table_default:"" table_descr:"" table_list:"Status reason"` // e.g. doesn't boot into OS
	// fixme latested test
	Comment string `json:"comment" yaml:"comment" gorm:"size:65536" table_default:"" table_descr:""`
}

// Return a list of all the boards
func GetAllBoards() ([]Board, error) {

	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Board{})

	// Read
	var boardList []Board
	db.Find(&boardList)

	return boardList, nil
}

// Fetch an board based on the ID supplied
func getBoardByID(id int) (*Board, error) {
	var b Board

	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.First(&b, id)
	return &b, nil
}

// Create a new board with the title and content provided
func CreateNewBoard(data string) (*Board, error) {

	var b Board

	err := yaml.Unmarshal([]byte(data), &b)
	if err != nil {
		fmt.Printf("Failed to unmarshal %v\n", err)
		return nil, err
	}
	b.Status = "UNKN"
	b.StatusComment = "Not tested yet"

	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	if !db.HasTable(&Board{}) {
		db.CreateTable(&Board{})
	}
	db.Create(&b)

	return &b, nil
}

func (b Board) render(c *gin.Context, showOnly bool) {
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
		if board, err := getBoardByID(ID); err == nil {
			RenderItems, err := getRenderItem(board)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
			} else {
				helper.Render(c, gin.H{
					"Name":        board.Name,
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

func (b Board) RenderShow(c *gin.Context) {
	b.render(c, true)
}

func (b Board) RenderEdit(c *gin.Context) {
	b.render(c, false)
}

func (b Board) RenderAll(c *gin.Context) {
	var Item string

	if strings.Contains(c.Request.RequestURI, "/") {
		Item = strings.Split(c.Request.RequestURI, "/")[1]
	} else {
		Item = c.Request.RequestURI
	}

	// Check if the board exists
	if boards, err := GetAllBoards(); err == nil {
		Header, List, err := getRenderList(boards)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		} else {
			helper.Render(c, gin.H{
				"Name":        "All boards",
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
