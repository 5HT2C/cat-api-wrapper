package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
)

var (
	addr  = flag.String("addr", "localhost:6060", "TCP address to listen to")
	token = flag.String("token", "", "TheCatAPI token")
)

func main() {
	log.Print("- Loading cat-api-wrapper")

	maxBodySize := 100 * 1024 * 1024 // 100 MiB

	s := &fasthttp.Server{
		Handler:            requestHandler,
		Name:               "srp-go",
		MaxRequestBodySize: maxBodySize,
	}

	if err := s.ListenAndServe(*addr); err != nil {
		log.Fatalf("- Error in ListenAndServe: %s", err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Method()) {
	case fasthttp.MethodGet:
		handleGet(ctx)
	case fasthttp.MethodPost:
		handleApi(ctx)
	default:
		handleError(ctx, 400, "Bad Method")
	}
}

type CatApiObj struct {
	Breeds []string `json:"breeds"`
	ID     string   `json:"id"`
	URL    string   `json:"url"`
	Width  int64    `json:"width"`
	Height int64    `json:"height"`
}

func requestCat(ctx *fasthttp.RequestCtx) CatApiObj {
	body, err := RequestUrl("https://api.thecatapi.com/v1/images/search", "x-api-key", *token)
	if err != nil {
		panic(err)
	}

	log.Printf("body: %s\n", body)
	var objs []CatApiObj
	if err := json.Unmarshal(body, &objs); err != nil {
		panic(err)
	}

	return objs[0]
}

func handleGet(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/api/random":
		cat := requestCat(ctx)
		fmt.Fprintf(ctx, "%s\n", cat.URL)
	default:
		handleError(ctx, 400, "Bad GET path")
	}
}

// handleApi will handle POST requests to the API
func handleApi(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/api/created":
		ctx.SetStatusCode(201)
		fmt.Fprintf(ctx, "201 Created\n")
	default:
		handleError(ctx, 400, "Bad API path")
	}
}

func handleError(ctx *fasthttp.RequestCtx, status int, msg string) {
	ctx.SetStatusCode(status)
	fmt.Fprintf(ctx, "%v %s\n", status, msg)
}

// RequestUrl will return the bytes of the body of url
func RequestUrl(url string, header string, value string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)
	req.Header.Set(header, value)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the request
	err := fasthttp.Do(req, resp)
	if err != nil {
		fmt.Printf("Client get failed: %s\n", err)
		return nil, err
	}
	return resp.Body(), nil
}
