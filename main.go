package ustcProjects

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

var router = mux.NewRouter()
var db *sql.DB

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
func initDB() {
	var err error

	config := mysql.Config{
		User:                 "root",
		Passwd:               "12345678",
		Addr:                 "127.0.0.1:3306",
		Net:                  "tcp",
		DBName:               "ustcTest",
		AllowNativePasswords: true,
	}

	// 准备数据库连接池
	db, err := sql.Open("mysql", config.FormatDSN())
	checkError(err)
	// 设置最大连接数
	db.SetMaxOpenConns(25)

	// 设置最大空闲连接数
	db.SetMaxIdleConns(25)

	// 设置每个链接的过期时间
	db.SetConnMaxLifetime(5 * time.Minute)

	// 尝试连接，失败会报错
	err = db.Ping()
	checkError(err)

}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello, 欢迎来到 情绪管理管理系统！</h1>")
}
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w,
		"此系统用于学生和老师进行互动，以便于老师实时了解学生对课程的掌握情况。，如您有反馈或建议，请联系 "+
			"<a href=\"mailto:herrluk@example."+
			"com\">herrluk@example.com</a>")
}
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>请求页面未找到 :(</h1><p>如有疑惑，请联系我们。</p>")
}

func main() {
	initDB()
	//createTables()

	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHandler).Methods("GET").Name("about")

	// 自定义 404 界面
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println(err)
	}
}
