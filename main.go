package main

import (
	"github.com/xtlsoft/router"
	"net/http"
	"strconv"
	"io/ioutil"
	"os"
	"bytes"
	"image/png"
	"image/jpeg"
	"fmt"
	"image/gif"
	"flag"
	"github.com/skip2/go-qrcode"
	"encoding/base64"
)

type req struct {
	writer http.ResponseWriter
	request *http.Request
}

func checkCache(r *req, data string, size int, levelString string, typ string) bool {

	fileName := fmt.Sprintf("%s-%dpx-level%s.%s", base64.StdEncoding.EncodeToString([]byte(data)), size, levelString, typ)

	if d, err := ioutil.ReadFile("./data/cache/" + fileName); err == nil {
		r.writer.Write(d)
		var mime string
		switch typ {
			case "jpg":
				mime = "image/jpeg"
			case "png":
				mime = "image/png"
			case "gif":
				mime = "image/gif"
			default:
				mime = "text/plain"
		}
		r.writer.Header().Add("content-type", mime)
		return true
	}else{
		return false
	}

}

func generateQr(r *req, data string, size int, levelString string, typ string) {
	var level qrcode.RecoveryLevel
	switch levelString {
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

	buf := bytes.NewBufferString("")

	switch typ {
		case "jpg":
			jpeg.Encode(buf, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
			r.writer.Header().Add("content-type", "image/jpeg")
		case "png":
			png.Encode(buf, img)
			r.writer.Header().Add("content-type", "image/png")
		case "gif":
			gif.Encode(buf, img, &gif.Options{})
			r.writer.Header().Add("content-type", "image/gif")
	}

	fileName := fmt.Sprintf("%s-%dpx-level%s.%s", base64.StdEncoding.EncodeToString([]byte(data)), size, levelString, typ)

	ioutil.WriteFile("./data/cache/" + fileName, buf.Bytes(), os.ModeAppend)

	r.writer.Write(buf.Bytes())

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
			
			if data == "" || typ == "" {
				r.writer.Write([]byte("Invalid Request."))
				r.writer.WriteHeader(500)
				return
			}
			
			if !checkCache(r, data, size, r.request.FormValue("level"), typ){
				generateQr(r, data, size, r.request.FormValue("level"), typ)
			}

			r.writer.WriteHeader(200)

		})
		g.Get("/qr/{data}/{@int:size}px-level{[1-4]:level}.{@ident:type}", func(r *req, vars map[string]string){
			
			dBytes, err := base64.StdEncoding.DecodeString(vars["data"])
			if err != nil {
				r.writer.Write([]byte("Invalid Request"))
				r.writer.WriteHeader(500)
				return
			}
			size, _ := strconv.Atoi(vars["size"])
			data, level, typ := string(dBytes),
									  vars["level"],
									  vars["type"]
			
			if !checkCache(r, data, size, level, typ){
				generateQr(r, data, size, level, typ)
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