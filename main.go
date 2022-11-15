package ustcProjects

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type StudentFormData struct {
	studentId     int
	studentNumber int
	studentName   string
	studentAge    int
	studentGender string
	studentGrade  int
	studentMajor  string
	studentPhone  int
}

type TeacherFormData struct {
	teacherId     int
	teacherName   string
	teacherAge    int
	teacherGender string
	teacherPhone  int
}

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

// 返回当前解析到的参数
func getRouteVariable(parameterName string, r *http.Request) string {
	vars := mux.Vars(r)
	return vars[parameterName]
}

func forceHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 设置标头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 2. 继续处理请求
		next.ServeHTTP(w, r)
	})
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 除首页以外，移除所有请求路径后面的斜杆
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}

		// 2. 将请求传递下去
		next.ServeHTTP(w, r)
	})
}

// 返回以学生姓名的内容
func getStudentInfoByName(id string) (StudentFormData,
	error) {
	studentInfo := StudentFormData{}
	query := "SELECT * FROM students WHERE stu_name = ?"
	err := db.QueryRow(query, id).Scan(&studentInfo.
		studentId, &studentInfo.studentNumber,
		&studentInfo.studentName, &studentInfo.studentAge,
		&studentInfo.studentGender,
		&studentInfo.studentGrade,
		&studentInfo.studentMajor, &studentInfo.studentPhone)
	return studentInfo, err
}

func saveStudentToDB(studentInfo StudentFormData) (int64,
	error) {
	// 变量初始化
	var (
		id   int64
		err  error
		rs   sql.Result
		stmt *sql.Stmt
	)

	// 1. 获取一个 prepare 声明语句
	stmt, err = db.Prepare("insert into students(stu_id, " +
		"stu_number, stu_name, stu_age, stu_gender, " +
		"stu_grade, stu_major, stu_phone) VALUES(? ,?,?," +
		"?,?,?,?,?)")
	// 例行错误检测
	if err != nil {
		return 0, err
	}

	// 2.在此函数运行结束后关闭此语句，防止占用 SQL 连接
	defer stmt.Close()

	// 3.执行请求，传参进入绑定设备
	rs, err = stmt.Exec(studentInfo)
	if err != nil {
		return 0, err
	}

	// 4.插入成功的话，返回自增 ID
	if id, err = rs.LastInsertId(); id > 0 {
		return id, nil
	}
	return 0, err

}

func studentShowHandler(w http.ResponseWriter,
	r *http.Request) {

	// 1. 获取 URL 参数
	name := getRouteVariable("name", r)

	// 2. 读取对应的学生数据
	studentInfo := StudentFormData{}
	query := "SELECT * FROM students WHERE stu_name" +
		" = ?" //根据输入的姓名查询
	stmt, err := db.Prepare(query)
	checkError(err)
	defer stmt.Close()
	err = stmt.QueryRow(name).Scan(&studentInfo.
		studentId, &studentInfo.studentNumber,
		&studentInfo.studentName, &studentInfo.studentAge,
		&studentInfo.studentGender,
		&studentInfo.studentGrade,
		&studentInfo.studentMajor, &studentInfo.studentPhone)
	// 3. 如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			// 3.1 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 学生信息未找到")
		} else {
			// 3.2 数据库错误
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		// 4. 读取成功，显示学生信息
		fmt.Fprint(w, "读取成功，学生信息如下所示")
	}
}

func studentEditHandler(w http.ResponseWriter,
	r *http.Request) {

	// 1. 获取 URL 参数
	vars := mux.Vars(r)
	id := vars["id"] //根据学生的编号

	// 2. 读取对应的文章数据
	studentInfo := StudentFormData{}
	query := "SELECT * FROM students WHERE stu_id = ?"
	err := db.QueryRow(query, id).Scan(&studentInfo.
		studentId, &studentInfo.studentNumber,
		&studentInfo.studentName, &studentInfo.studentAge,
		&studentInfo.studentGender,
		&studentInfo.studentGrade,
		&studentInfo.studentMajor, &studentInfo.studentPhone)

	// 3. 如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			// 3.1 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 学生信息未找到")
		} else {
			// 3.2 数据库错误
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		// 4. 读取成功，显示表单
		updateURL, _ := router.Get("student.update").URL(
			"id", id)
		data := StudentFormData{}
		tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
		checkError(err)

		err = tmpl.Execute(w, data)
		checkError(err)
	}
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
