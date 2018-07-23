package main

import (
	"errors"

	"github.com/cheikhshift/gos/core"
	gosweb "github.com/cheikhshift/gos/web"
	"gopkg.in/mgo.v2/bson"
	//iogos-replace
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/cheikhshift/db"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/fatih/color"
	"github.com/gorilla/sessions"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
)

var store = sessions.NewCookieStore([]byte("a very very very very secret key"))

var Prod = true

var TemplateFuncStore template.FuncMap
var templateCache = gosweb.NewTemplateCache()

func StoreNetfn() int {
	TemplateFuncStore = template.FuncMap{"a": gosweb.Netadd, "s": gosweb.Netsubs, "m": gosweb.Netmultiply, "d": gosweb.Netdivided, "js": gosweb.Netimportjs, "css": gosweb.Netimportcss, "sd": gosweb.NetsessionDelete, "sr": gosweb.NetsessionRemove, "sc": gosweb.NetsessionKey, "ss": gosweb.NetsessionSet, "sso": gosweb.NetsessionSetInt, "sgo": gosweb.NetsessionGetInt, "sg": gosweb.NetsessionGet, "form": gosweb.Formval, "eq": gosweb.Equalz, "neq": gosweb.Nequalz, "lte": gosweb.Netlt, "hash": Nethash}
	return 0
}

var FuncStored = StoreNetfn()

type dbflf db.O

func renderTemplate(w http.ResponseWriter, p *gosweb.Page) {
	defer func() {
		if n := recover(); n != nil {
			color.Red(fmt.Sprintf("Error loading template in path : web%s.tmpl reason : %s", p.R.URL.Path, n))

			DebugTemplate(w, p.R, fmt.Sprintf("web%s", p.R.URL.Path))
			w.WriteHeader(http.StatusInternalServerError)

			pag, err := loadPage("/your-500-page")

			if err != nil {
				log.Println(err.Error())
				return
			}

			if pag.IsResource {
				w.Write(pag.Body)
			} else {
				pag.R = p.R
				pag.Session = p.Session
				renderTemplate(w, pag) ///your-500-page"

			}
		}
	}()

	// TemplateFuncStore

	if _, ok := templateCache.Get(p.R.URL.Path); !ok || !Prod {
		var tmpstr = string(p.Body)
		var localtemplate = template.New(p.R.URL.Path)

		localtemplate.Funcs(TemplateFuncStore)
		localtemplate.Parse(tmpstr)
		templateCache.Put(p.R.URL.Path, localtemplate)
	}

	outp := new(bytes.Buffer)
	err := templateCache.JGet(p.R.URL.Path).Execute(outp, p)
	if err != nil {
		log.Println(err.Error())
		DebugTemplate(w, p.R, fmt.Sprintf("web%s", p.R.URL.Path))
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "text/html")
		pag, err := loadPage("/your-500-page")

		if err != nil {
			log.Println(err.Error())
			return
		}
		pag.R = p.R
		pag.Session = p.Session

		if pag.IsResource {
			w.Write(pag.Body)
		} else {
			renderTemplate(w, pag) // "/your-500-page"

		}
		return
	}

	// p.Session.Save(p.R, w)

	var outps = outp.String()
	var outpescaped = html.UnescapeString(outps)
	outp = nil
	fmt.Fprintf(w, outpescaped)

}

// Access you .gxml's end tags with
// this http.HandlerFunc.
// Use MakeHandler(http.HandlerFunc) to serve your web
// directory from memory.
func MakeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if attmpt := apiAttempt(w, r); !attmpt {
			fn(w, r)
		}

	}
}

func mResponse(v interface{}) string {
	data, _ := json.Marshal(&v)
	return string(data)
}
func apiAttempt(w http.ResponseWriter, r *http.Request) (callmet bool) {
	var response string
	response = ""
	var session *sessions.Session
	var er error
	if session, er = store.Get(r, "session-"); er != nil {
		session, _ = store.New(r, "session-")
	}

	if strings.Contains(r.URL.Path, "/get_configuration") {

		if r.Method == "OPTIONS" {
			return true
		}

		if r.Method == "GET" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,TOKEN")
		}

	}
	if r.Method == "RESET" {
		return true
	} else if isURL := (r.URL.Path == "/login" && r.Method == strings.ToUpper("POST")); !callmet && isURL {

		var u User

		err := Login(r, &u)
		if err != nil {
			HandleError(w, err, 404)
			return true
		}

		session.Values["userid"] = u.Id.Hex()

		response = mResponse(bson.M{})

		callmet = true
	} else if isURL := (r.URL.Path == "/join" && r.Method == strings.ToUpper("POST")); !callmet && isURL {

		var u User

		err := Join(r, &u)
		if err != nil {
			HandleError(w, err, 400)
			return true
		}

		session.Values["userid"] = u.Id.Hex()

		response = mResponse(bson.M{})

		callmet = true
	} else if isURL := (r.URL.Path == "/" && r.Method == strings.ToUpper("GET")); !callmet && isURL {

		http.Redirect(w, r, "/login", 301)

		callmet = true
	} else if isURL := (r.URL.Path == "/logout" && r.Method == strings.ToUpper("GET")); !callmet && isURL {

		if _, ok := session.Values["userid"]; ok {
			delete(session.Values, "userid")
		}

		response = mResponse(bson.M{})

		callmet = true
	} else if isURL := (r.URL.Path == "/reset" && r.Method == strings.ToUpper("POST")); !callmet && isURL {

		var u User
		err := ParseBody(r, &u)
		if err != nil {
			HandleError(w, err, 400)
			return true
		}
		passw := core.NewLen(6)
		newPassword := Nethash(passw)

		err = ResetPassword(u.Email, newPassword)
		if err != nil {
			HandleError(w, err, 500)
			return true
		}

		err = SendEmail(u.Email, "Password reset email", fmt.Sprintf("Hi,\n\n Please use the following password to log into your account :\n🔒Password : %s\n\nThank you for your patience.", passw))
		if err != nil {
			HandleError(w, err, 404)
			return true
		}

		response = mResponse(bson.M{})

		callmet = true
	} else if isURL := (r.URL.Path == "/delete_account" && r.Method == strings.ToUpper("GET")); !callmet && isURL {

		var u User
		var c Configuration

		id := session.Values["userid"].(string)

		err := dbs.Query(u, bson.M{"_id": bson.ObjectIdHex(id)}).One(&u)
		if err != nil {
			HandleError(w, err, 500)
			return true
		}

		_, err = dbs.RemoveAll(u, bson.M{"_id": bson.ObjectIdHex(id)})
		if err != nil {
			HandleError(w, err, 500)
			return true
		}

		_, err = dbs.RemoveAll(c, bson.M{"owner": id})

		if err != nil {
			HandleError(w, err, 500)
			return true
		}

		_, _ = sub.Cancel(u.StripeID, nil)

		response = mResponse(bson.M{})

		callmet = true
	} else if isURL := (r.URL.Path == "/get_configuration" && r.Method == strings.ToUpper("GET")); !callmet && isURL {

		if token := r.Header.Get("TOKEN"); token != "" {

			var c Configuration

			err := dbs.Query(c, bson.M{"apikey": token}).One(&c)

			if err != nil {
				HandleError(w, err, 401)
				return
			}

			c.ApiKey = ""

			response = mResponse(c)

		}

		callmet = true
	} else if isURL := (r.URL.Path == "/account_status" && r.Method == strings.ToUpper("GET")); !callmet && isURL {

		s := HasSubscribed(session.Values["userid"].(string))
		if s {
			response = mResponse(bson.M{})
		}

		if !s {
			err := errors.New("Information not specified!")
			HandleError(w, err, 500)
			return true
		}

		callmet = true
	} else if isURL := (r.URL.Path == "/process_stripe" && r.Method == strings.ToUpper("POST")); !callmet && isURL {

		if token := r.FormValue("stripeToken"); token != "" {
			id := session.Values["userid"].(string)
			params := &stripe.CustomerParams{Description: stripe.String(fmt.Sprintf("Customer for  %s", id))}
			params.SetSource(token)
			cus, err := customer.New(params)
			if err != nil {
				log.Println(err)
				http.Redirect(w, r, "/account?error=Error! Please try again", 307)
				return true
			}

			paramk := &stripe.SubscriptionParams{Customer: stripe.String(cus.ID), Items: []*stripe.SubscriptionItemsParams{{Plan: stripe.String("configd")}}}
			s, err := sub.New(paramk)
			if err != nil {
				log.Println(err)
				http.Redirect(w, r, "/account?error=Error! Please try again", 307)
				return true
			}

			var u User
			subId := s.ID

			err = dbs.Query(u, bson.M{"_id": bson.ObjectIdHex(id)}).One(&u)
			if err != nil {
				log.Println(err)
				http.Redirect(w, r, "/account?error=Error! Please try again", 307)
				return true
			}

			u.StripeID = subId

			dbs.Update(&u)

			http.Redirect(w, r, "/account?success=Card information saved!", 307)
			return true
		}

		callmet = true
	} else if isURL := (r.URL.Path == "/update_password" && r.Method == strings.ToUpper("POST")); !callmet && isURL {

		var u User
		err := ParseBody(r, &u)
		if err != nil {
			HandleError(w, err, 400)
			return true
		}

		err = UpdatePassword(session.Values["userid"].(string), u.Email, u.Password)
		if err != nil {
			HandleError(w, err, 500)
			return true
		}

		response = mResponse(bson.M{})

		callmet = true
	} else if !callmet && gosweb.UrlAtZ(r.URL.Path, "/configurations") {

		if userid, ok := session.Values["userid"].(string); ok {

			if r.Method == "GET" {
				var configs []Configuration

				err := ListConfigurations(r, userid, &configs)
				if err != nil {
					HandleError(w, err, 500)
					return true
				}
				response = mResponse(configs)
			}

			if r.Method == "POST" {

				err := AddConfiguration(r, userid)
				if err != nil {
					HandleError(w, err, 500)
					return true
				}
				response = mResponse(bson.M{})
			}

			if r.Method == "PUT" {
				err := UpdateConfiguration(r, userid)
				if err != nil {
					HandleError(w, err, 500)
					return true
				}
				response = mResponse(bson.M{})
			}

			if r.Method == "DELETE" {

				err := DeleteConfiguration(r, userid)
				if err != nil {
					HandleError(w, err, 500)
					return true
				}
				response = mResponse(bson.M{})

			}
		} else if !ok {
			HandleError(w, nil, 401)
		}

		callmet = true
	}

	if callmet {
		session.Save(r, w)
		session = nil
		if response != "" {
			//Unmarshal json
			//w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(response))
		}
		return
	}
	session = nil
	return
}

func DebugTemplate(w http.ResponseWriter, r *http.Request, tmpl string) {
	lastline := 0
	linestring := ""
	defer func() {
		if n := recover(); n != nil {
			log.Println()
			// log.Println(n)
			log.Println("Error on line :", lastline+1, ":"+strings.TrimSpace(linestring))
			//http.Redirect(w,r,"/your-500-page",307)
		}
	}()

	p, err := loadPage(r.URL.Path)
	filename := tmpl + ".tmpl"
	body, err := Asset(filename)
	session, er := store.Get(r, "session-")

	if er != nil {
		session, er = store.New(r, "session-")
	}
	p.Session = session
	p.R = r
	if err != nil {
		log.Print(err)

	} else {

		lines := strings.Split(string(body), "\n")
		// log.Println( lines )
		linebuffer := ""
		waitend := false
		open := 0
		for i, line := range lines {

			processd := false

			if strings.Contains(line, "{{with") || strings.Contains(line, "{{ with") || strings.Contains(line, "with}}") || strings.Contains(line, "with }}") || strings.Contains(line, "{{range") || strings.Contains(line, "{{ range") || strings.Contains(line, "range }}") || strings.Contains(line, "range}}") || strings.Contains(line, "{{if") || strings.Contains(line, "{{ if") || strings.Contains(line, "if }}") || strings.Contains(line, "if}}") || strings.Contains(line, "{{block") || strings.Contains(line, "{{ block") || strings.Contains(line, "block }}") || strings.Contains(line, "block}}") {
				linebuffer += line
				waitend = true

				endstr := ""
				processd = true
				if !(strings.Contains(line, "{{end") || strings.Contains(line, "{{ end") || strings.Contains(line, "end}}") || strings.Contains(line, "end }}")) {

					open++

				}
				for i := 0; i < open; i++ {
					endstr += "\n{{end}}"
				}
				//exec
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string(body))
				lastline = i
				linestring = line
				erro := t.Execute(outp, p)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}
			}

			if waitend && !processd && !(strings.Contains(line, "{{end") || strings.Contains(line, "{{ end")) {
				linebuffer += line

				endstr := ""
				for i := 0; i < open; i++ {
					endstr += "\n{{end}}"
				}
				//exec
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string(body))
				lastline = i
				linestring = line
				erro := t.Execute(outp, p)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}

			}

			if !waitend && !processd {
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string(body))
				lastline = i
				linestring = line
				erro := t.Execute(outp, p)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}
			}

			if !processd && (strings.Contains(line, "{{end") || strings.Contains(line, "{{ end")) {
				open--

				if open == 0 {
					waitend = false

				}
			}
		}

	}

}

func DebugTemplatePath(tmpl string, intrf interface{}) {
	lastline := 0
	linestring := ""
	defer func() {
		if n := recover(); n != nil {

			log.Println("Error on line :", lastline+1, ":"+strings.TrimSpace(linestring))
			log.Println(n)
			//http.Redirect(w,r,"/your-500-page",307)
		}
	}()

	filename := tmpl
	body, err := Asset(filename)

	if err != nil {
		log.Print(err)

	} else {

		lines := strings.Split(string(body), "\n")
		// log.Println( lines )
		linebuffer := ""
		waitend := false
		open := 0
		for i, line := range lines {

			processd := false

			if strings.Contains(line, "{{with") || strings.Contains(line, "{{ with") || strings.Contains(line, "with}}") || strings.Contains(line, "with }}") || strings.Contains(line, "{{range") || strings.Contains(line, "{{ range") || strings.Contains(line, "range }}") || strings.Contains(line, "range}}") || strings.Contains(line, "{{if") || strings.Contains(line, "{{ if") || strings.Contains(line, "if }}") || strings.Contains(line, "if}}") || strings.Contains(line, "{{block") || strings.Contains(line, "{{ block") || strings.Contains(line, "block }}") || strings.Contains(line, "block}}") {
				linebuffer += line
				waitend = true

				endstr := ""
				if !(strings.Contains(line, "{{end") || strings.Contains(line, "{{ end") || strings.Contains(line, "end}}") || strings.Contains(line, "end }}")) {

					open++

				}

				for i := 0; i < open; i++ {
					endstr += "\n{{end}}"
				}
				//exec

				processd = true
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string([]byte(fmt.Sprintf("%s%s", linebuffer, endstr))))
				lastline = i
				linestring = line
				erro := t.Execute(outp, intrf)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}
			}

			if waitend && !processd && !(strings.Contains(line, "{{end") || strings.Contains(line, "{{ end") || strings.Contains(line, "end}}") || strings.Contains(line, "end }}")) {
				linebuffer += line

				endstr := ""
				for i := 0; i < open; i++ {
					endstr += "\n{{end}}"
				}
				//exec
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string([]byte(fmt.Sprintf("%s%s", linebuffer, endstr))))
				lastline = i
				linestring = line
				erro := t.Execute(outp, intrf)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}

			}

			if !waitend && !processd {
				outp := new(bytes.Buffer)
				t := template.New("PageWrapper")
				t = t.Funcs(TemplateFuncStore)
				t, _ = t.Parse(string([]byte(fmt.Sprintf("%s%s", linebuffer))))
				lastline = i
				linestring = line
				erro := t.Execute(outp, intrf)
				if erro != nil {
					log.Println("Error on line :", i+1, line, erro.Error())
				}
			}

			if !processd && (strings.Contains(line, "{{end") || strings.Contains(line, "{{ end") || strings.Contains(line, "end}}") || strings.Contains(line, "end }}")) {
				open--

				if open == 0 {
					waitend = false

				}
			}
		}

	}

}
func Handler(w http.ResponseWriter, r *http.Request) {
	var p *gosweb.Page
	p, err := loadPage(r.URL.Path)
	var session *sessions.Session
	var er error
	if session, er = store.Get(r, "session-"); er != nil {
		session, _ = store.New(r, "session-")
	}

	if err != nil {
		log.Println(err.Error())

		w.WriteHeader(http.StatusNotFound)

		pag, err := loadPage("/your-404-page")

		if err != nil {
			log.Println(err.Error())
			//
			return
		}
		pag.R = r
		pag.Session = session
		if p != nil {
			p.Session = nil
			p.Body = nil
			p.R = nil
			p = nil
		}

		if pag.IsResource {
			w.Write(pag.Body)
		} else {
			renderTemplate(w, pag) //"/your-500-page"
		}
		session = nil

		return
	}

	if !p.IsResource {
		w.Header().Set("Content-Type", "text/html")
		p.Session = session
		p.R = r
		renderTemplate(w, p) //fmt.Sprintf("web%s", r.URL.Path)
		session.Save(r, w)
		// log.Println(w)
	} else {
		w.Header().Set("Cache-Control", "public")
		if strings.Contains(r.URL.Path, ".css") {
			w.Header().Add("Content-Type", "text/css")
		} else if strings.Contains(r.URL.Path, ".js") {
			w.Header().Add("Content-Type", "application/javascript")
		} else {
			w.Header().Add("Content-Type", http.DetectContentType(p.Body))
		}

		w.Write(p.Body)
	}

	p.Session = nil
	p.Body = nil
	p.R = nil
	p = nil
	session = nil

	return
}

var WebCache = gosweb.NewCache()

func loadPage(title string) (*gosweb.Page, error) {

	if lPage, ok := WebCache.Get(title); ok {
		return &lPage, nil
	}

	var nPage = gosweb.Page{}
	if roottitle := (title == "/"); roottitle {
		webbase := "web/"
		fname := fmt.Sprintf("%s%s", webbase, "index.html")
		body, err := Asset(fname)
		if err != nil {
			fname = fmt.Sprintf("%s%s", webbase, "index.tmpl")
			body, err = Asset(fname)
			if err != nil {
				return nil, err
			}
			nPage.Body = body
			WebCache.Put(title, nPage)
			body = nil
			return &nPage, nil
		}
		nPage.Body = body
		nPage.IsResource = true
		WebCache.Put(title, nPage)
		body = nil
		return &nPage, nil

	}

	filename := fmt.Sprintf("web%s.tmpl", title)

	if body, err := Asset(filename); err != nil {
		filename = fmt.Sprintf("web%s.html", title)

		if body, err = Asset(filename); err != nil {
			filename = fmt.Sprintf("web%s", title)

			if body, err = Asset(filename); err != nil {
				return nil, err
			} else {
				if strings.Contains(title, ".tmpl") {
					return nil, nil
				}
				nPage.Body = body
				nPage.IsResource = true
				WebCache.Put(title, nPage)
				body = nil
				return &nPage, nil
			}
		} else {
			nPage.Body = body
			nPage.IsResource = true
			WebCache.Put(title, nPage)
			body = nil
			return &nPage, nil
		}
	} else {
		nPage.Body = body
		WebCache.Put(title, nPage)
		body = nil
		return &nPage, nil
	}

}

var dbs db.DB

//
func Nethash(args ...interface{}) string {
	input := args[0]

	sum := sha256.Sum256([]byte(input.(string)))
	return fmt.Sprintf("%x", sum)

}
func dummy_timer() {
	dg := time.Second * 5
	log.Println(dg)
}
func main() {
	fmt.Fprintf(os.Stdout, "%v\n", os.Getpid())

	//psss go code here : func main()
	var err error

	if MongoURI == "" {
		MongoURI = "localhost"
	}

	if MongoDBName == "" {
		MongoDBName = "configd"
	}

	dbs, err = db.Connect(MongoURI, MongoDBName)
	if err != nil {
		panic(err)
	}
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
		Domain:   "",
	}

	port := ":8000"
	if envport := os.ExpandEnv("$PORT"); envport != "" {
		port = fmt.Sprintf(":%s", envport)
	}
	log.Printf("Listenning on Port %v\n", port)

	//+++extendgxmlmain+++

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt)
	http.Handle("/dist/", http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "web"}))
	http.HandleFunc("/", MakeHandler(Handler))

	h := &http.Server{Addr: port}

	go func() {
		errgos := h.ListenAndServe()
		if errgos != nil {
			log.Fatal(errgos)
		}
	}()

	<-stop

	log.Println("\nShutting down the server...")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	h.Shutdown(ctx)

	log.Println("Server gracefully stopped")

}

//+++extendgxmlroot+++
