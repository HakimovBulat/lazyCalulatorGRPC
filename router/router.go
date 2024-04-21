package router

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/HakimovBulat/lazyCalulatorGRPC/utils"
	"github.com/apaxa-go/eval"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Expression struct {
	Id            int
	StringVersion string
	Answer        string
	Status        string
	StartDate     time.Time
	EndDate       time.Time
}
type User struct {
	Name     string
	Password string
}

var mapOperatorsTime = map[string]int{
	"-": 10,
	"+": 10,
	"*": 10,
	"/": 10,
}

var connectionString = "host=0.0.0.0 port=5432 user=postgres password=Love_and_elephant42 dbname=Expression sslmode=disable"
var connection *sql.DB

func SetupRouter() *gin.Engine {
	utils.SetupLogger()
	var err error
	connection, err = sql.Open("postgres", connectionString)
	//connection.Query(`DROP TABLE "Expression"`)
	if err != nil {
		utils.Logger.Error(err.Error())
	}
	_, err = connection.Query(`CREATE TABLE IF NOT EXISTS "Expression" (
		"id" serial NOT NULL,
		"StringVersion" text NOT NULL,
		"Status" text NOT NULL,
		"Answer" text NOT NULL,
		"StartDate" TIMESTAMP WITH TIME ZONE NOT NULL,
		"EndDate" TIMESTAMP WITH TIME ZONE NOT NULL,
		CONSTRAINT "Expression_pk" PRIMARY KEY ("id")
	) WITH (
	  OIDS=FALSE
	);`)
	if err != nil {
		utils.Logger.Error(err.Error())
	}
	_, err = connection.Query(`
	CREATE TABLE IF NOT EXISTS "Users"(
		"id" serial NOT NULL UNIQUE,
		"Name" text NOT NULL,
		"Password" text NOT NULL,
		CONSTRAINT "Users_pk" PRIMARY KEY ("id")
	) WITH (
	  OIDS=FALSE
	);`)
	if err != nil {
		utils.Logger.Error(err.Error())
	}
	router := gin.Default()
	router.GET("/", inputExpression)
	router.POST("/", createExpression)
	router.GET("/operators", showOperatorsTime)
	router.PUT("/operators", replaceOperatorsTime)
	router.GET("/static_operators", operatorsStatic)
	router.POST("/static_operators", operatorsStatic)
	router.GET("/get_expression/:id", getExpression)
	router.GET("/login", getLoginHandler)
	router.GET("/logout", logoutHandler)
	router.POST("/login", postLoginHandler)
	router.GET("/register", getRegisterHandler)
	router.POST("/register", postRegisterHandler)
	router.LoadHTMLGlob("templates/*.html")
	return router
}
func logoutHandler(c *gin.Context) {
	cookie := &http.Cookie{
		Name:    "user",
		MaxAge:  -1,
		Value:   "",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(c.Writer, cookie)
	c.HTML(200, "index.html", gin.H{
		"Expressions": getExpressionsList(),
		"User":        "",
	})
}
func getRegisterHandler(c *gin.Context) {
	c.HTML(200, "register.html", nil)
}
func getLoginHandler(c *gin.Context) {
	c.HTML(200, "login.html", nil)
}
func postLoginHandler(c *gin.Context) {
	login, password := c.PostForm("login"), c.PostForm("password")
	rows, err := connection.Query(`SELECT "Name", "Password" FROM "Users" WHERE "Name"=$1 AND "Password"=$2`, login, password)
	if err != nil {
		utils.Logger.Error(err.Error())
		c.HTML(http.StatusOK, "error.html", gin.H{"Error": "Пользователя не существует"})
	}
	users := []User{}
	for rows.Next() {
		user := new(User)
		rows.Scan(
			&user.Name,
			&user.Password,
		)
		users = append(users, *user)
		rows.Close()
	}
	if len(users) != 1 {
		c.HTML(http.StatusOK, "error.html", gin.H{"Error": "Пользователя не существует"})
	}
	cookie := &http.Cookie{Name: "user", Value: login, MaxAge: 300}
	http.SetCookie(c.Writer, cookie)
	http.Redirect(c.Writer, c.Request, "/", http.StatusFound)

}
func postRegisterHandler(c *gin.Context) {
	login, password := c.PostForm("login"), c.PostForm("password")
	_, err := connection.Query(`
	CREATE TABLE IF NOT EXISTS "Users"(
		"id" serial NOT NULL,
		"Name" text NOT NULL,
		"Password" text NOT NULL,
		CONSTRAINT "Users_pk" PRIMARY KEY ("id")
	) WITH (
	  OIDS=FALSE
	);`)
	if err != nil {
		utils.Logger.Error(err.Error())
	}
	rows, err := connection.Query(`SELECT "Name", "Password" FROM "Users" WHERE "Name"=$1`, login)
	if err != nil {
		utils.Logger.Error(err.Error())
	}
	users := []User{}
	for rows.Next() {
		user := new(User)
		rows.Scan(&user.Name, &user.Password)
		users = append(users, *user)
		rows.Close()
	}
	if len(users) > 0 {
		c.HTML(http.StatusOK, "error.html", gin.H{"Error": "Пользователь с таким логином уже существует"})
	} else {
		_, err = connection.Query(`INSERT INTO "Users" ("Name", "Password") values ($1, $2)`, login, password)
		if err != nil {
			utils.Logger.Error(err.Error())
		}
		cookie := &http.Cookie{Name: "user", Value: login, MaxAge: 300}
		http.SetCookie(c.Writer, cookie)
		http.Redirect(c.Writer, c.Request, "/login", http.StatusFound)
	}
}

func inputExpression(c *gin.Context) {

	var err error
	cookie, err := c.Request.Cookie("user")
	if err != nil {
		cookie = &http.Cookie{Name: "user", Value: "", MaxAge: 300}
	}
	http.SetCookie(c.Writer, cookie)
	rows, err := connection.Query(`SELECT * FROM "Expression"`)
	if err != nil {
		utils.Logger.Error(err.Error())
	}
	defer rows.Close()
	expressions := []Expression{}
	for rows.Next() {
		validExpression := new(Expression)
		rows.Scan(
			&validExpression.Id,
			&validExpression.StringVersion,
			&validExpression.Status,
			&validExpression.Answer,
			&validExpression.StartDate,
			&validExpression.EndDate,
		)
		now := time.Now()
		if validExpression.EndDate.Before(now) {
			if validExpression.Answer != "not found" {
				validExpression.Status = "ok"
			} else {
				validExpression.Status = "cancel"
			}
		}
		_, err = connection.Query(`UPDATE "Expression" SET "StringVersion"=$2, "Status"=$3, "Answer"=$4, "StartDate"=$5, "EndDate"=$6
			WHERE "id"=$1`,
			validExpression.Id,
			validExpression.StringVersion,
			validExpression.Status,
			validExpression.Answer,
			validExpression.StartDate,
			validExpression.EndDate,
		)
		if err != nil {
			utils.Logger.Error(err.Error())
		}
		expressions = append(expressions, *validExpression)
	}
	c.HTML(200, "index.html", gin.H{
		"Expressions": expressions,
		"User":        cookie.Value,
	})
}
func getExpressionsList() []Expression {
	rows, err := connection.Query(`SELECT * FROM "Expression"`)
	if err != nil {
		utils.Logger.Error(err.Error())
	}
	defer rows.Close()
	expressions := []Expression{}
	for rows.Next() {
		validExpression := new(Expression)
		rows.Scan(
			&validExpression.Id,
			&validExpression.StringVersion,
			&validExpression.Status,
			&validExpression.Answer,
			&validExpression.StartDate,
			&validExpression.EndDate,
		)
		now := time.Now()
		if validExpression.EndDate.Before(now) {
			if validExpression.Answer != "not found" {
				validExpression.Status = "ok"
			} else {
				validExpression.Status = "cancel"
			}
		}
		_, err = connection.Query(`UPDATE "Expression" SET "StringVersion"=$2, "Status"=$3, "Answer"=$4, "StartDate"=$5, "EndDate"=$6
			WHERE "id"=$1`,
			validExpression.Id,
			validExpression.StringVersion,
			validExpression.Status,
			validExpression.Answer,
			validExpression.StartDate,
			validExpression.EndDate,
		)
		if err != nil {
			utils.Logger.Error(err.Error())
		}
		expressions = append(expressions, *validExpression)
	}
	return expressions
}
func createExpression(c *gin.Context) {
	now := time.Now()
	var newExpression Expression
	mathExpression := c.PostForm("math")
	utils.Logger.Info(mathExpression)
	endDate := getTime(mathExpression, now)
	newExpression = Expression{
		Id:            0,
		StringVersion: mathExpression,
		Status:        "ok",
		StartDate:     now,
		EndDate:       endDate,
	}
	expr, err := eval.ParseString(newExpression.StringVersion, "")
	if err != nil {
		newExpression.Status = "cancel"
	} else {
		answer, err := expr.EvalToInterface(nil)
		if err != nil || answer == nil {
			newExpression.Answer = "not found"
		} else {
			newExpression.Answer = fmt.Sprint(answer)
		}
	}
	newExpression.Status = "process"
	expressions := []Expression{}
	if mathExpression != "" {
		_, err := connection.Query(`INSERT INTO "Expression" ("StringVersion", "Status", "Answer", "StartDate", "EndDate")
		VALUES($1, $2, $3, $4, $5)`,
			newExpression.StringVersion,
			newExpression.Status,
			newExpression.Answer,
			newExpression.StartDate,
			newExpression.EndDate,
		)
		if err != nil {
			utils.Logger.Error(err.Error())
		}
		rows, err := connection.Query(`SELECT * FROM "Expression"`)
		if err != nil {
			utils.Logger.Error(err.Error())
		}
		for rows.Next() {
			validExpression := new(Expression)
			rows.Scan(
				&validExpression.Id,
				&validExpression.StringVersion,
				&validExpression.Status,
				&validExpression.Answer,
				&validExpression.StartDate,
				&validExpression.EndDate,
			)
			expressions = append(expressions, *validExpression)
		}
		rows.Close()
	}
	cookie, err := c.Request.Cookie("user")
	if err != nil {
		cookie = &http.Cookie{Name: "user", Value: "", MaxAge: 300}
	}
	http.SetCookie(c.Writer, cookie)
	c.HTML(200, "index.html", gin.H{
		"Expressions": expressions,
		"User":        cookie.Value,
	})
}

func showOperatorsTime(c *gin.Context) {
	cookie, err := c.Request.Cookie("user")
	if err != nil {
		cookie = &http.Cookie{Name: "user", Value: "", MaxAge: 300}
	}
	http.SetCookie(c.Writer, cookie)
	c.HTML(200, "operators.html", gin.H{
		"addition":       mapOperatorsTime["+"],
		"substraction":   mapOperatorsTime["-"],
		"multiplication": mapOperatorsTime["*"],
		"division":       mapOperatorsTime["/"],
		"User":           cookie.Value,
	})
}

func replaceOperatorsTime(c *gin.Context) {
	cookie, err := c.Request.Cookie("user")
	if err != nil {
		cookie = &http.Cookie{Name: "user", Value: "", MaxAge: 300}
	}
	http.SetCookie(c.Writer, cookie)
	mapOperatorsTime["+"], _ = strconv.Atoi(c.PostForm("addition"))
	mapOperatorsTime["-"], _ = strconv.Atoi(c.PostForm("substraction"))
	mapOperatorsTime["*"], _ = strconv.Atoi(c.PostForm("multiplication"))
	mapOperatorsTime["/"], _ = strconv.Atoi(c.PostForm("division"))
	c.HTML(200, "operators.html", gin.H{
		"addition":       mapOperatorsTime["+"],
		"substraction":   mapOperatorsTime["-"],
		"multiplication": mapOperatorsTime["*"],
		"division":       mapOperatorsTime["/"],
		"User":           cookie.Value,
	})
}
func operatorsStatic(c *gin.Context) {
	cookie, err := c.Request.Cookie("user")
	if err != nil {
		cookie = &http.Cookie{Name: "user", Value: "", MaxAge: 300}
	}
	http.SetCookie(c.Writer, cookie)
	if c.Request.Method == "POST" {
		mapOperatorsTime["+"], _ = strconv.Atoi(c.PostForm("addition"))
		mapOperatorsTime["-"], _ = strconv.Atoi(c.PostForm("substraction"))
		mapOperatorsTime["*"], _ = strconv.Atoi(c.PostForm("multiplication"))
		mapOperatorsTime["/"], _ = strconv.Atoi(c.PostForm("division"))
	}
	c.HTML(200, "static_operators.html", gin.H{
		"addition":       mapOperatorsTime["+"],
		"substraction":   mapOperatorsTime["-"],
		"multiplication": mapOperatorsTime["*"],
		"division":       mapOperatorsTime["/"],
		"User":           cookie.Value,
	})
}
func getExpression(c *gin.Context) {
	var expression Expression
	connection.QueryRow(`SELECT * FROM "Expression" WHERE "id"=$1`, c.Param("id")).Scan(
		&expression.Id,
		&expression.StringVersion,
		&expression.Status,
		&expression.Answer,
		&expression.StartDate,
		&expression.EndDate,
	)
	var answer any
	var err error
	expr, err := eval.ParseString(expression.StringVersion, "")
	if err != nil {
		expression.Status = "cancel"
		c.JSON(400, expression)
		return
	}
	now := time.Now()
	if expression.EndDate.Before(now) {
		answer, err = expr.EvalToInterface(nil)
		if err != nil || answer == nil {
			expression.Status = "cancel"
			c.JSON(400, expression)
			return
		}
		expression.Answer = fmt.Sprint(answer)
		expression.Status = "ok"
	}
	c.JSON(200, expression)
}
func getTime(expression string, now time.Time) time.Time {
	var seconds int
	for _, symbol := range expression {
		if string(symbol) == "-" || string(symbol) == "+" || string(symbol) == "/" || string(symbol) == "*" {
			seconds += mapOperatorsTime[string(symbol)]
		}
	}
	return now.Add(time.Duration(seconds) * time.Second)
}
