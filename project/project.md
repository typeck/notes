**Swagger 使用方法**

* Add comments to your API source code, [See Declarative Comments Format](https://github.com/swaggo/swag#declarative-comments-format).
* Download [Swag](https://github.com/swaggo/swag) for Go by using:
```sh
$ go get github.com/swaggo/swag/cmd/swag
```
* Run the swag in your Go project root folder which contains `main.go` file.
  `swag int`

* import it.
```go
import (
    echoSwagger "github.com/swaggo/echo-swagger"
    _ "PROJECT/docs"
    e.GET("/swagger/*", echoSwagger.WrapHandler)
```
* Run it, and browser to http://localhost:1323/swagger/index.html, you can see Swagger 2.0 Api documents.
* example comments.
```go
// @title ab test system API
// @version 1.0
// @description This is ab test system server.

// @contact.name type
// @contact.email tangyp@ushareit.com

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name token
func main()
```
```go
// @Summary Edit test
// @Tags test
// @Produce  json
// @Accept  json
// @Param base_test_id query int false "base test id"
// @Param layer_id query string  false "layer id"
// @Param  version body []models.Abtest true "Update test"
// @Success 200 {object} handler.Response
// @Failure 500 {object} handler.Response
// @Security ApiKeyAuth
// @Router /test/{id} [post]
func Update()
```

