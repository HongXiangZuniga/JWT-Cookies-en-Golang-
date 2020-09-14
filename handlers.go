package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"

	_ "github.com/go-sql-driver/mysql"
)

type TokenDetails struct {
	AccessToken string
	AccessUuid  string
	AtExpires   int64
}

type User struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var RequestUser User
	var BDUser User

	if err := c.ShouldBindJSON(&RequestUser); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "Invalid json provided")
		return
	}

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/pc2")
	if err != nil {
		c.JSON(http.StatusFailedDependency, err.Error())
		return
	}

	//Busqueda en la base de datos
	check := false
	result, err := db.Query("SELECT * FROM User WHERE User.Username  =  ?", RequestUser.Username)
	if err != nil {
		c.JSON(http.StatusFailedDependency, err.Error())
		return
	} else {
		for result.Next() { //Se busca por cada dato obtenido de la consulta
			err = result.Scan(&BDUser.ID, &BDUser.Username, &BDUser.Password)
			if err != nil {
				c.JSON(http.StatusFailedDependency, err.Error())
				return
			}
			if BDUser.Username == RequestUser.Username && BDUser.Password == RequestUser.Password {
				check = true //Si existe el usuario
			}
		}
	}

	db.Close()

	//Caso de Username o Password Erroneos
	if check != true {
		c.JSON(http.StatusUnauthorized, "Please provide valid login details")
		return
	}
	ts, err := CreateToken(BDUser.ID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	saveErr := CreateAuth(BDUser.ID, ts)
	if saveErr != nil {
		c.JSON(http.StatusUnprocessableEntity, saveErr.Error())
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "session_token",
		Value:   ts.AccessToken,
		Expires: time.Now().Add(15 * time.Minute),
	})

	tokens := map[string]string{
		"access_token": ts.AccessToken,
	}
	c.JSON(http.StatusOK, tokens)
}

func CreateAuth(userid uint64, td *TokenDetails) error {
	at := 900
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		return err
	}

	_, errAccess := conn.Do("SET", td.AccessUuid, strconv.Itoa(int(userid)))
	if errAccess != nil {
		fmt.Println("1" + errAccess.Error())
		return errAccess
	}
	_, errAccess = conn.Do("EXPIRE", td.AccessUuid, at)
	if errAccess != nil {
		fmt.Println("1" + errAccess.Error())
		return errAccess
	}
	return nil
}
