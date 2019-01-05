package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	gin "github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/gocolly/colly"
)

/**
 * StartServer is the main entrance into the server.
 */
func StartServer(serverPort string) {
	go startHealthServerNative(serverPort)

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()
	router.GET("/extract", extEndpoint)
	router.Run(":" + serverPort)

}

type pageInfo struct {
	StatusCode int
	Body      string
	Redirect []string
}

func extEndpoint(c *gin.Context) {

	url := c.Query("url");
	if(len(url) < 1){
		log.Warn("Url not found")
		return;
	}

	collector := colly.NewCollector(colly.AllowURLRevisit())

	p := &pageInfo{}
	

	collector.RedirectHandler = func(req *http.Request, via []*http.Request) error{
		p.Redirect = make([]string,0);
		p.Redirect = append(p.Redirect, req.URL.String())
		for _, element := range via {
			p.Redirect = append(p.Redirect, element.URL.String());
		}
		return nil
	};

	collector.OnError(func(r *colly.Response, err error) {
		log.Println("error:", r.StatusCode, err)
		p.StatusCode = r.StatusCode
		c.JSON(404,p)
	})

	collector.OnResponse(func(response *colly.Response) {
		p.Body = string(response.Body);
		p.StatusCode = response.StatusCode
		c.JSON(http.StatusOK, p)
	})

	collector.Visit(url)

}

//---- End of the Gin-based implementation of the home endpoint
//----
// Note: the native (non Gin-based) implementation below is hidden
// from docker-compose and the code is only shown here as a demonstration of
// how to implement things without Gin, if you wanted to do so
func startHealthServerNative(serverPort string) {
	http.HandleFunc("/health", viewHandlerHealth)
	healthPort, errParse := strconv.Atoi(serverPort)
	healthPort = healthPort + 1
	healthPortStr := strconv.Itoa(healthPort)
	abortIfErr(errParse)
	err := http.ListenAndServe(":"+healthPortStr, nil)
	if err != nil {
		log.Fatal("ERROR: couldn't start server: ", err)
	} else {
		log.Info("Healthcheck started successfully at: ", healthPortStr)
	}
}

func viewHandlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/health+json")
	response := make(map[string]string)
	response["status"] = "pass"

	res, err := json.MarshalIndent(response, "", "  ")
	abortIfErr(err)
	fmt.Fprintf(w, string(res))
}

// Simple exit if error, to avoid putting same 4 lines of code in too many places
func abortIfErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
