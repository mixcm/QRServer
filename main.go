package main

import (
	"github.com/xtlsoft/router"
	"net/http"
	"strconv"
	"io/ioutil"
	"image/png"
	"image/jpeg"
	"image/gif"
	"flag"
	"github.com/skip2/go-qrcode"
)

type req struct {
	writer http.ResponseWriter
	request *http.Request
}

func readFile(path string) string {
	r, _ := ioutil.ReadFile(path)
	return string(r)
}

func toInt(str string) int {
	r, _ := strconv.Atoi(str)
	return r
}

func main() {

	port := flag.String("port", ":21384", "Port to listen.")
	node := flag.String("node", "Default", "Server node name.")
	flag.Parse()

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
		r.writer.WriteHeader(404)
		r.writer.Write([]byte(readFile("./template/404.html")))
	})

	// Main Code
	r.Group("/", func (g *router.Group){
		g.Get("/", func (r *req, vars map[string]string){
			r.writer.Header().Add("location", "/doc")
			r.writer.WriteHeader(302)
		})
		g.Get("/generate", func (r *req, vars map[string]string){
			data, size, typ := r.request.FormValue("data"),
							   toInt(r.request.FormValue("size")),
							   r.request.FormValue("type")
			var level qrcode.RecoveryLevel
			switch r.request.FormValue("level") {
				case "1":
					level = qrcode.Low
				case "2":
					level = qrcode.Medium
				case "3":
					level = qrcode.High
				case "4":
					level = qrcode.Highest
			}
			qr, _ := qrcode.New(data, level)
			img := qr.Image(size)

			switch typ {
				case "jpg":
					jpeg.Encode(r.writer, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
					r.writer.Header().Add("content-type", "image/jpeg")
				case "png":
					png.Encode(r.writer, img)
					r.writer.Header().Add("content-type", "image/png")
				case "gif":
					gif.Encode(r.writer, img, &gif.Options{})
					r.writer.Header().Add("content-type", "image/gif")
			}

			r.writer.WriteHeader(200)

		})
	})
	// End Main Code

	http.HandleFunc("/", func(w http.ResponseWriter, ree *http.Request){
		
		re := &req{
			writer: w,
			request: ree,
		}
		r.Handle(ree.Method, ree.RequestURI, re)

	})

	http.Handle("/assets", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	http.Handle("/doc", http.StripPrefix("/doc", http.FileServer(http.Dir("./doc"))))

	println("Server started at " + *port)

	http.ListenAndServe(*port, nil)

}