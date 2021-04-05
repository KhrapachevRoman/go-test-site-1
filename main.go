package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Article struct {
	Id                     uint16
	Title, Anons, FullText string
}

type Handler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {

	// Выборка данных
	rows, err := h.DB.Query("SELECT * FROM `articles`")
	__err_panic(err)
	defer rows.Close()

	articles := []Article{}
	for rows.Next() {
		article := Article{}
		err := rows.Scan(&article.Id, &article.Title, &article.Anons, &article.FullText)
		__err_panic(err)
		articles = append(articles, article)
	}

	// Выводим шаблон
	err = h.Tmpl.ExecuteTemplate(w, "index", articles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) AddForm(w http.ResponseWriter, r *http.Request) {
	// Выводим шаблон
	err := h.Tmpl.ExecuteTemplate(w, "create", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) AddArticle(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	anons := r.FormValue("anons")
	fullText := r.FormValue("full_text")

	// Мини валидация
	if title == "" || anons == "" || fullText == "" {
		fmt.Fprintf(w, "Не все данные заполнены")
		return
	}

	// Установка данных
	_, err := h.DB.Exec(
		"INSERT INTO `articles` (`title`,`anons`,`full_text`) VALUES(?, ?, ?)",
		title,
		anons,
		fullText,
	)
	__err_panic(err)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	__err_panic(err)

	article := Article{}
	// QueryRow сам закрывает коннект
	row := h.DB.QueryRow("SELECT id, title, anons, full_text FROM articles WHERE id = ?", id)

	err = row.Scan(&article.Id, &article.Title, &article.Anons, &article.FullText)
	__err_panic(err)

	// Выводим шаблон
	err = h.Tmpl.ExecuteTemplate(w, "article", article)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleFunc() {

	//++ БЛОК РАБОТЫ С БД
	// основные настройки к базе
	dsn := "root:root@tcp(127.0.0.1:8889)/golang"
	db, err := sql.Open("mysql", dsn)
	db.SetMaxOpenConns(10)
	err = db.Ping() // вот тут будет первое подключение к базе
	__err_panic(err)
	//-- БЛОК РАБОТЫ С БД

	handlers := &Handler{
		DB:   db,
		Tmpl: template.Must(template.ParseGlob("templates/*")),
	}

	router := mux.NewRouter()
	router.HandleFunc("/", handlers.Index).Methods("GET")
	router.HandleFunc("/articles/new", handlers.AddForm).Methods("GET")
	router.HandleFunc("/articles/new", handlers.AddArticle).Methods("POST")
	router.HandleFunc("/articles/{id:[0-9]+}", handlers.Edit).Methods("GET")

	http.Handle("/", router)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))) //обработка статичных файлов

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}

func main() {

	handleFunc()
}

//TODO обработать ошибки, ошибка должна всегда явно обрабатываться
func __err_panic(err error) {
	if err != nil {
		panic(err)
	}
}
