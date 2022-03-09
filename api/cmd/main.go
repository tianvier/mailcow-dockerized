package main

import (
	"fmt"
	"mailcow-api/middleware"
	"mailcow-api/models"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

const DEFAULT_PWD = "123456"

func main() {

	r := gin.Default()

	var (
		db_user = "root"
		db_pass = "123456"
		db_name = "test"
		db_host = "127.0.1"
		db_port = "8889"
	)
	if os.Getenv("DBUSER") != "" {
		db_user = os.Getenv("DBUSER")
	}
	if os.Getenv("DBPASS") != "" {
		db_pass = os.Getenv("DBPASS")
	}
	if os.Getenv("DBNAME") != "" {
		db_name = os.Getenv("DBNAME")
	}
	if os.Getenv("DBHOST") != "" {
		db_host = os.Getenv("DBHOST")
	}
	if os.Getenv("DBPORT") != "" {
		db_port = os.Getenv("DBPORT")
	}
	//db_user := os.Getenv("DBUSER") ? os.Getenv("DBUSER") : "root"

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", db_user, db_pass, db_host, db_port, db_name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		fmt.Println("failed to connect database")
	}
	DB = db

	d_db, _ := DB.DB()
	defer d_db.Close()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	api := r.Group("/api")
	api.Use(middleware.AuthToken())
	{
		api.GET("/create", createUser)
		api.GET("/find", findUser)
	}

	r.Run() // listen and serve on 0.0.0.0:8080
}

func createUser(c *gin.Context) {

	username := c.Query("username")
	password := c.Query("password")
	fmt.Println(username)
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "username is required",
			"status": "false",
		})
		return
	}

	if password == "" {
		var err error
		password, err = passwordHash(DEFAULT_PWD)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  err,
				"status": "false",
			})
			return
		}
	} else {
		var err error
		password, err = passwordHash(password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  err,
				"status": "false",
			})
			return
		}
	}

	var mailbox models.Mailbox
	DB.Where("username = ?", username).First(&mailbox)

	//fmt.Println(mailbox)
	if condition := mailbox.Username; condition != "" {
		c.JSON(http.StatusOK, gin.H{
			"status": "true",
			"data":   mailbox,
		})
		return
	}

	name := strings.Split(username, "@")
	d_password := fmt.Sprintf("{BLF-CRYPT}%s", password)
	d_attributes := "{\"force_pw_update\": \"0\", \"tls_enforce_in\": \"1\", \"tls_enforce_out\": \"1\", \"sogo_access\": \"1\", \"imap_access\": \"1\", \"pop3_access\": \"1\", \"smtp_access\": \"1\", \"sieve_access\": \"1\", \"relayhost\": \"0\", \"passwd_update\": \"2022-03-02 03:37:33\", \"mailbox_format\": \"maildir:\", \"quarantine_notification\": \"hourly\", \"quarantine_category\": \"reject\"}"
	d_quota := 102400
	//d_quota := 5242880
	// sql 1
	tx1 := DB.Exec("INSERT INTO `mailbox` (`username`, `password`, `name`, `quota`, `local_part`, `domain`, `attributes`, `active`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", username, d_password, name[0], d_quota, name[0], name[1], d_attributes, 1)
	if tx1.Error != nil {
		fmt.Println(tx1.Error)
		c.JSON(200, gin.H{
			"status": "false",
			"error":  tx1.Error.Error(),
		})
		return
	}
	// sql 2
	tx2 := DB.Exec("INSERT INTO `quota2` (`username`, `bytes`, `messages`) VALUES (?, 0, 0)", username)
	if tx2.Error != nil {
		fmt.Println(tx2.Error)
		c.JSON(200, gin.H{
			"status": "false",
			"error":  tx2.Error.Error(),
		})
		return
	}
	// sql 3
	tx3 := DB.Exec("INSERT INTO `quota2replica` (`username`, `bytes`, `messages`) VALUES (?, 0, 0)", username)
	if tx3.Error != nil {
		fmt.Println(tx3.Error)
		c.JSON(200, gin.H{
			"status": "false",
			"error":  tx3.Error.Error(),
		})
		return
	}
	// sql 4
	tx4 := DB.Exec("INSERT INTO `alias` (`address`,`goto`,`domain`,`active`) VALUES (?,?,?,1)", username, username, name[1])
	if tx4.Error != nil {
		fmt.Println(tx4.Error)
		c.JSON(200, gin.H{
			"status": "false",
			"error":  tx4.Error.Error(),
		})
		return
	}
	// sql 5
	tx5 := DB.Exec("INSERT INTO `user_acl` (`username`) VALUES (?)", username)
	if tx5.Error != nil {
		fmt.Println(tx5.Error)
		c.JSON(200, gin.H{
			"status": "false",
			"error":  tx5.Error.Error(),
		})
		return
	}
	DB.Where("username = ?", username).First(&mailbox)

	c.JSON(200, gin.H{
		"status": "true",
		"data":   mailbox,
	})
}

func findUser(c *gin.Context) {
	var mailbox models.Mailbox
	var total int64
	DB.Model(&mailbox).Count(&total)

	//fmt.Println(total)
	rand.Seed(time.Now().UnixNano())
	// 随机生成用户 或者从现有用户中选择
	randNum := rand.Intn(int(total))
	//fmt.Println(randNum)
	DB.Limit(1).Offset(randNum).Find(&mailbox)

	fmt.Println(mailbox)

	c.JSON(200, mailbox)
}

func passwordHash(passwd string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(passwd), 12)
	return string(bytes), err
}
