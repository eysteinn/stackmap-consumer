package api

import (
	"log"
	"net/http"
	"strings"
	"test/api/files"

	_ "embed"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed post.html
var post_html []byte

/*func ParseURL(obj *prog.ConsumerObject, val url.Values) error {

	timestr := val.Get("timestamp")
	if timestr != "" {
		tmp, err := time.Parse("20060102T150405", timestr)
		if err != nil {
			return err
		}
		obj.Timestamp = &tmp
	}
	if val.Has("product") {
		obj.Product = val.Get("product")
	}
	if val.Has("project") {
		obj.Project = val.Get("project")
	}

	file_web_url := val.Get("file.web.url")
	file_web_type := val.Get("file.web.type")
	if file_web_url != "" && file_web_type != "" {
		web := prog.Web{}
		web.Url = file_web_url
		web.Type = file_web_type
		obj.File.Web = &web
	}

	file_local_path := val.Get("file.local.path")
	if file_local_path != "" {
		local := prog.Local{}
		local.Path = val.Get("file.local.path")
		obj.File.Local = &local
	}
	return nil
}*/

func extractContentType(request *http.Request) string {
	contenttype := request.Header.Get("Content-Type")
	contenttype = strings.Split(contenttype, ";")[0]
	return contenttype
}

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		next.ServeHTTP(w, r)
	})
}
func filesRoutes() *chi.Mux {
	router := chi.NewRouter()
	//router.Post("/upload", uploadPost)
	router.Post("/projects/{project}/files", files.PostHandler)

	router.Get("/projects/{project}/files", files.GetHandler)
	router.Get("/projects/{project}/products/{product}/files", files.GetHandler)

	router.Delete("/projects/{project}/files/{uuid}", files.DeleteHandle)

	//router.Options("/projects/{project}/files", uploadPost)

	return router
}

func Run() {

	/*err := prog.Psql.TryFill()
	if err != nil {
		log.Println(err)
	}*/
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(Cors)
	router.Mount("/api/v1/", filesRoutes())
	//mux.Get("/html/simplepost", html)

	/*router.Get("/upload", html)
	router.Post("/upload", uploadPost)*/
	/*router.Get("/projects/{project}/files", files.GetHandler)
	router.Delete("/projects/{project}/files/{uuid}", files.DeleteHandle)*/
	/*mux.Route("/projects/{project}/files", func(r chi.Router) {
		r.Delete("/{uuid}", files.DeleteHandle)
	})*/
	//mux.Delete("/files", deleteFile) make this dynamic /files/{uuid}
	/*mux := http.NewServeMux()
	mux.HandleFunc("/upload", upload)
	mux.HandleFunc("/delete", delete)*/
	hostaddress := "0.0.0.0:3333"
	log.Println("Starting to listen: " + hostaddress)
	err := http.ListenAndServe(hostaddress, router)
	if err != nil {
		log.Panic(err)
	}
}
