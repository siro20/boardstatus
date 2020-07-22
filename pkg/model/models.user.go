// models.user.go

package model

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	helper "github.com/siro20/boardstatus/pkg/helper"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model
	Username string `json:"username" table_default:"" table_descr:"The username" table_list:"Username"`
	Name     string `json:"name" table_default:"" table_descr:"The real name"  table_list:"Real Name"`

	Email             string `json:"email" table_default:"" table_descr:"The e-mail"  table_list:"E-Mail"`
	Hidden            bool   `json:"hidden" table_default:"" table_descr:"Is hidden user"  table_list:"Is Hidden"` // User is invisible to public and other users
	IsAdmin           bool   `json:"is_admin" table_default:"" table_descr:"Is Admin user"  table_list:"Is Admin"` // Admins can delete, add, modify users, boards and tests
	ProfilePictureURL string `json:"profile_picture_url" table_default:"" table_descr:"Profile picture URL"`       // Admins can delete, add, modify users, boards and tests

	OAuthProvider string `json:"oauth" gorm:"oauth_provider" table_default:"" table_descr:"OAuth Provider"  table_list:"OAuth Provider"` // Admins can delete, add, modify users, boards and tests

	// A user can have an API token
	ApiToken           string `json:"api_token" gorm:"api_token" table_default:"" table_descr:"The API token"`
	BasicAuthorization string `json:"basic_auth" gorm:"basic_auth" table_default:"" table_descr:"The Basic Auth String"` // as defined in RFC 2617

	// Password isn't stored in DB
	Password     string `json:"-" gorm:"-" table_default:"" table_descr:"The secret Password"`
	PasswordHash string `json:"password_hash" gorm:"password_hash" table_default:"" table_descr:"Password hash"`
}

// Check if the supplied username is available
func isUsernameAvailable(username string) bool {
	user, err := GetUserByName(username)
	if user == nil || err != nil {
		return true
	}
	return false
}

// Return a list of all the users
func getAllUser() ([]User, error) {

	var users []User

	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	db.AutoMigrate(&User{})

	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func getUserByID(id int) (*User, error) {
	var u User

	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	db.AutoMigrate(&User{})

	if err := db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByTag(id string, value string) (*User, error) {
	var u User

	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	db.AutoMigrate(&User{})

	if err := db.Where(id+" = ?", value).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByBasicAuth(auth string) (*User, error) {
	return GetUserByTag("basic_auth", auth)
}

func GetUserByName(name string) (*User, error) {
	return GetUserByTag("username", name)
}

func GetUserByEmail(email string) (*User, error) {
	return GetUserByTag("email", email)
}

func GetUserByOAuthLogin(login string, provider string) (*User, error) {
	var u User
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	db.AutoMigrate(&User{})

	if err := db.Where(&User{Name: login, OAuthProvider: provider}).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func UserIsPasswordValid(u *User, pass string) (bool, error) {
	if u.Password == "" {
		return false, nil
	}
	byteHash := []byte(u.PasswordHash)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(pass))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (u *User) DeleteFromDB() error {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return err
	}
	defer db.Close()

	if !db.HasTable(&User{}) {
		db.CreateTable(&User{})
	}

	db.AutoMigrate(&User{})

	if err := db.Delete(u).Error; err != nil {
		return err
	}
	return nil
}

func (u *User) InsertIntoDB() error {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return err
	}
	defer db.Close()

	if !db.HasTable(&User{}) {
		db.CreateTable(&User{})
	}

	db.AutoMigrate(&User{})

	// Update fields
	if u.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)
		if err != nil {
			log.Println(err)
		} // GenerateFromPassword returns a byte slice so we need to
		// convert the bytes to a string and return it
		u.PasswordHash = string(hash)
	}
	if err := db.Create(u).Error; err != nil {
		return err
	}
	return nil
}

func (u User) render(c *gin.Context, showOnly bool) {
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
		if user, err := getUserByID(ID); err == nil {
			RenderItems, err := getRenderItem(user)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
			} else {
				helper.Render(c, gin.H{
					"Name":        user.Name,
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

func (u User) RenderShow(c *gin.Context) {
	u.render(c, true)
}

func (u User) RenderEdit(c *gin.Context) {
	u.render(c, false)
}

func (u User) RenderAll(c *gin.Context) {
	var Item string

	if strings.Contains(c.Request.RequestURI, "/") {
		Item = strings.Split(c.Request.RequestURI, "/")[1]
	} else {
		Item = c.Request.RequestURI
	}

	// Check if the board exists
	if users, err := getAllUser(); err == nil {
		Header, List, err := getRenderList(users)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		} else {
			helper.Render(c, gin.H{
				"Name":        "All users",
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
