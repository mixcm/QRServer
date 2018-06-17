package main

import (
	"github.com/xtlsoft/router"
	"net/http"
	"io/ioutil"
	"flag"
)

type req struct {
	writer http.ResponseWriter
	request *http.Request
}

func readFile(path string) string {
	r, _ := ioutil.ReadFile(path)
	return string(r)
}

func main() {

	port := flag.String("port", ":21384", "Port to listen.")
	node := flag.String("node", "Default", "Server node name.")

	http.Handle("/assets", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	http.Handle("/doc", http.StripPrefix("/doc", http.FileServer(http.Dir("./doc"))))

	r := router.New()

	r.SetHandler(
		func(isMatched bool, request interface{}, controller interface{}, variables map[string]string) interface{} {

			c, r := controller.(func(*req, map[string]string)), request.(*req)

			r.writer.Header().Add("X-Powered-By", "Mixcm-QRServer")
			r.writer.Header().Add("Server", "Mixcm-QRServer")
			r.writer.Header().Add("X-Mixcm-Node", *node)

			c(r, variables)

			return nil

		},
	)

	r.Any(func (r *req, vars map[string]string){
		r.writer.WriteHeader(200)
		r.writer.Write([]byte(readFile("./template/404.html")))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, ree *http.Request){
		
		re := &req{
			writer: w,
			request: ree,
		}
		r.Handle(ree.Method, ree.RequestURI, re)

	})

	http.ListenAndServe(*port, nil)

}