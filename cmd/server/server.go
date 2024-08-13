package main

import (
	"Gallery/controllers"
	"Gallery/migrations"
	"Gallery/models"
	"Gallery/templates"
	"Gallery/views"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}
	cfg.PSQL = models.PostgresConfig{
		Host:     os.Getenv("PSQL_HOST"),
		Port:     os.Getenv("PSQL_PORT"),
		User:     os.Getenv("PSQL_USER"),
		Password: os.Getenv("PSQL_PASSWORD"),
		Database: os.Getenv("PSQL_DATABASE"),
		SSLMode:  os.Getenv("PSQL_SSLMODE"),
	}

	if cfg.SMTP.Host == "" && cfg.PSQL.Port == "" {
		return cfg, fmt.Errorf("np psql config provided")
	}

	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")

	// cfg.CSRF.Key = "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	cfg.CSRF.Secure = os.Getenv("CSRF_SECURE") == "true"

	// cfg.Server.Address = ":3000"
	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")

	return cfg, nil
}

func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}
	err = run(cfg)
	if err != nil {
		panic(err)
	}
}

func run(cfg config) error {
	// Set up a database connection
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		return err
	}
	defer db.Close()
	// err = models.Migrate(db, "migrations") // 数据库迁移
	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		return err
	}

	// Set up service
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}
	pwResetService := &models.PasswordResetService{
		DB: db,
	}
	galleryService := &models.GalleryService{
		DB: db,
	}
	emailService := models.NewEmailService(cfg.SMTP)

	// set up middleware
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}
	// csrfKey应当只包含A-Z, a-z, 0-9 范围内的字符，且长度必须是32字节
	csrfMw := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		// TODO: Fix this before deploying
		csrf.Secure(cfg.CSRF.Secure),
		csrf.Path("/"), //设置为所有路径使用csrf，csrf的默认情况是为/<name>使用，子路径多的情况无法使用
	)

	// Set up controllers
	usersC := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: pwResetService,
		EmailService:         emailService,
	}
	galleriesC := controllers.Galleries{
		GalleryService: galleryService,
	}
	galleriesC.Templates.New = views.Must(views.ParseFS(templates.FS, "galleries/new.gohtml", "tailwind.gohtml"))
	galleriesC.Templates.Edit = views.Must(views.ParseFS(templates.FS, "galleries/edit.gohtml", "tailwind.gohtml"))
	galleriesC.Templates.Index = views.Must(views.ParseFS(templates.FS, "galleries/index.gohtml", "tailwind.gohtml"))
	galleriesC.Templates.Show = views.Must(views.ParseFS(templates.FS, "galleries/show.gohtml", "tailwind.gohtml"))
	usersC.Templates.New = views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))
	usersC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(templates.FS, "forgot-pw.gohtml", "tailwind.gohtml"))
	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(templates.FS, "check-your-email.gohtml", "tailwind.gohtml"))
	usersC.Templates.ResetPassword = views.Must(views.ParseFS(templates.FS, "reset-pw.gohtml", "tailwind.gohtml"))
	// Set up router and routes
	// "/"表示所有路由的默认访问处理句柄
	// r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "home.gohtml", "layout-parts.gohtml"))))
	r := chi.NewRouter()
	r.Use(csrfMw) // 添加中间件
	r.Use(umw.SetUser)

	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))

	// r.Get("/signup", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))))
	r.Get("/signup", usersC.New)
	r.Post("/signup", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.With(umw.RequireUser).Post("/signout", usersC.ProcessSignOut)
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)

	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})
	// r.Get("/users/me", controllers.MakeMiddleware(usersC.CurrentUser))

	r.Route("/galleries", func(r chi.Router) { // 定义一个路由前缀为/galleries的路由组
		r.Get("/{id}", galleriesC.Show)
		r.Get("/{id}/images/{filename}", galleriesC.Image)
		r.Group(func(r chi.Router) { // 定义另外一个路由组，便于使用中间件，不会改变路由的路径
			r.Use(umw.RequireUser)
			r.Get("/", galleriesC.Index)
			r.Get("/new", galleriesC.New)
			r.Post("/", galleriesC.Create)
			r.Get("/{id}/edit", galleriesC.Edit)
			r.Post("/{id}", galleriesC.Update)
			r.Post("/{id}/delete", galleriesC.Delete)
			// Make sure this is requires a user
			r.Post("/{id}/images/{filename}/delete", galleriesC.DeleteImage)
			r.Post("/{id}/images", galleriesC.UploadImage)
		})
	})
	// 将文件目录转化为路由，从而当作句柄处理，之后相关的路径由文件系统处理
	// 举例：当有人请求/assets/somefile.txt时，服务器会查找assets/somefile.txt文件并返回它
	// http.StripPrefix保证assestsHandler接收到的路径不包含/assets前缀，所以能够正确地查找文件
	assetsHandler := http.FileServer(http.Dir("assets"))
	r.Get("/assets/*", http.StripPrefix("/assets", assetsHandler).ServeHTTP) // 删除路由后的前缀，然后由句柄处理

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start the server
	fmt.Printf("Starting the server on%s...", cfg.Server.Address)
	// 这里设置的中间件，先进行csrf的验证，然后设置user，最后进入http处理
	return http.ListenAndServe(cfg.Server.Address, r)
	// 确实，使用 r.Delete 和 DELETE HTTP 方法来删除资源在直观上可能看起来更合适。
	// 但在实践中，不使用 JavaScript 来创建执行 DELETE 操作的链接和表单是相当麻烦的。
	// 因此，我们通常会选择使用 POST 方法来执行删除操作。

	// 我们不希望使用 GET 方法，因为浏览器和某些 JavaScript 库可能会跟随这些链接，并尝试缓存页面以加快浏览体验。
	// GET 请求通常被设计用来检索（读取）数据，而不是修改数据。GET 请求的结果通常可以被缓存，并且在多次请求时可能返回相同的结果。
	// 由于这些特性，任何会修改数据的页面都不应该通过 GET 请求来访问。
}
