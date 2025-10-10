package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ApiOptions struct {
	Port int
}

type Api struct {
	router *gin.Engine
	server *http.Server
}

func NewApi(opts ApiOptions) *Api {
	router := gin.Default()

	s := &http.Server{
		Addr: fmt.Sprintf(":%d", opts.Port),
		Handler: router,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &Api{
		router: router,
		server: s,
	}
}

func (api *Api) Serve(ctx context.Context) error {
	if err := api.server.ListenAndServe(); err != nil {
		fmt.Println("unable to serve http")
		panic(err)
	}

	defer api.server.Close()

	return nil
}

type Context struct {
	*gin.Context
}

type HandlerFunc func (*gin.Context) (status int, results any, err error)


func (api *Api) AddAPIHandler(path string, method string, handler HandlerFunc) {
	apiRouter := api.router.Group("/api")
	wrapped := func (ctx *gin.Context) {
		status, result, err := handler(ctx)
		
		if err != nil {
			ctx.Error(errors.Unwrap(err))
			return
		}

		ctx.JSON(status, result)
	}
	

	switch method {
    case "GET":
        apiRouter.GET(path, wrapped)
    case "POST":
        apiRouter.POST(path, wrapped)
	default:
		fmt.Errorf("Unknown method %s", method)
		return
    }
}

