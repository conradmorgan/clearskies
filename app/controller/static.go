package controller

import "net/http"

func Static(w http.ResponseWriter, r *http.Request) {
	root := http.Dir("static")
	file, err := root.Open(r.URL.Path)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))
		return
	} else if stat, _ := file.Stat(); stat.IsDir() {
		_, err := root.Open(r.URL.Path + "/index.html")
		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte("404 page not found"))
			return
		} else {
			r.URL.Path += "/index.html"
		}
	}
	http.FileServer(root).ServeHTTP(w, r)
}
