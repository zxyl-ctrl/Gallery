package main

// func main() {
// 	Demo()
// 	Demo(1)
// 	Demo(1, 2, 3)
// }

// // 可变形参必须是函数最后定义的变量，否则会报错
// // 为什么不使用[]int而使用...int呢，编写代码更容易

// func Demo(numbers ...int) {
// 	for _, number := range numbers {
// 		fmt.Print(number, " ")
// 	}
// 	fmt.Println()
// }

import (
	"io"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
)

// side effect副作用：在这里，在包pgx/v4/stdlib包运行init函数将会将会替换database/sql包的全局状态通过sql.Register
// 一般来说，不使用副作用，有以下原因
// 1. 无法防止副作用发生：即使您尝试避免副作用，但只要您因为其他原因导入了相同的包，该包中的初始化代码（如注册驱动）仍会执行。
// 这可能导致问题，尤其是当两个驱动使用相同的名称时。为了避免这种冲突，开发者需要确保他们使用的库和包遵循良好的命名约定，并且尽可能避免全局状态的共享。
// 2. 感觉像是“魔法”，难以调试和追踪：当代码中有隐藏的副作用时，它可能会让开发者感到困惑，尤其是在调试时。
// 您可能不知道为何某个函数在没有显式调用的情况下就被执行了，或者为什么某个变量在没有显式赋值的情况下就被设置了值。这种不可预测的行为会增加代码的复杂性和维护成本。
// 3. 具有副作用的代码测试具有挑战性和敏感性：测试包含副作用的代码通常更加困难，因为您需要模拟或控制那些副作用。
// 这可能意味着您需要设置额外的测试环境，或者使用模拟对象（mocks）和存根（stubs）来隔离被测试的代码。此外，这样的测试往往更加敏感，因为它们可能依赖于外部因素（如文件系统状态、数据库连接等），这些因素在测试环境中可能不容易控制。

// 使用这种方式的原因：过去遗留问题的决策

// func main() {
// 	cfg := models.DefaultPostgresConfig()
// 	db, err := models.Open(cfg)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer db.Close()
// 	err = db.Ping()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Connected!")

// // SQL 创建数据最后一行不能够写逗号
// // user_id用于将第二个表与用户(第一个表进行绑定)
// // id是主键
// _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
// 	id SERIAL PRIMARY KEY,
// 	name TEXT,
// 	email TEXT NOT NULL
// );

// CREATE TABLE IF NOT EXISTS orders (
// 	id SERIAL PRIMARY KEY,
// 	user_id INT NOT NULL,
// 	amount INT,
// 	description TEXT
// );`)
// if err != nil {
// 	panic(err)
// }
// fmt.Println("Tables created.")

// 为什么需要将SQL语句硬编码执行，目的就是防止篡改，其中传入的动态参数无法被修改为可以执行的SQL语句
// 即插入的字符串将作为变量存储，无法变为可执行SQL语句
// 见讲义10.6
// name, email := "John Calhoun", "john@cahoun.io"
// row := db.QueryRow(`
// 	INSERT INTO users(name, email)
// 	VALUES ($1, $2) RETURNING id;`, name, email)
// // $i表示占位符，在MySQL中，对应的是占位符是?  	// 创建新的记录并获取其ID
// var id int
// err = row.Scan(&id) // 传入指针并返回需要插入的变量，并将错误保存在row.Err中
// if err != nil {
// 	panic(err)
// }
// fmt.Println("User created. id =", id)
//QueryRow 最长用于执行查找单个记录的任务在SQL database中，它也可以执行任何想要的SQL语句，并返回需要的数据

// id := 200
// row := db.QueryRow(`
// 	SELECT name, email
// 	FROM users
// 	WHERE id=$1;`, id)
// var name, email string
// err = row.Scan(&name, &email) // 如果没有查找到，则无法使用Scan函数
// if err == sql.ErrNoRows {
// 	fmt.Println("Error, no rows!")
// }
// if err != nil {
// 	panic(err)
// }
// fmt.Printf("User information: name=%s, email=%s\n", name, email)

// 创建fake订单
// userID := 1 // Pick an ID that exists in your DB
// for i := 1; i < 5; i++ {
// 	amount := i * 100
// 	desc := fmt.Sprintf("Fake order #%d", i)
// 	_, err := db.Exec(`
// 		INSERT INTO orders(user_id, amount, description)
// 		VALUES($1, $2, $3)`, userID, amount, desc)
// 	if err != nil {
// 		panic(err)
// 	}
// }
// fmt.Println("Created fake orders.")

// 为什么Query返回带有Err而QueryRow不带有？应该是设计原因，也有可能是Query创建*sql.Rows导致性能较差
// 	type Order struct {
// 		ID          int
// 		UserID      int
// 		Amou        int
// 		Description string
// 	}
// 	var orders []Order

// userID := 1
// rows, err := db.Query(`
//
//	SELECT id, amount, description
//	FROM orders
//	WHERE user_id=$1`, userID) // 也可以将所有数据存入一个结构体
//
//	if err != nil {
//		panic(err)
//	}
//
// defer rows.Close() // 因为sql.Rows读取的数据量很大，因此为了防止内存泄漏，需要通知垃圾收集器不需要该变量
// for rows.Next() {  // 获得返回查询的下一个错误，如果有数据，返回true,如果遇到错误或者没有下一个数据，返回false
//
//		// 只有使用rows.Next()才可以获得第一个数据
//		var order Order
//		order.UserID = userID
//		err := rows.Scan(&order.ID, &order.Amou, &order.Description)
//		if err != nil {
//			panic(err)
//		}
//		orders = append(orders, order)
//	}
//
// err = rows.Err()
//
//	if err != nil {
//		panic(err)
//	}
//
// fmt.Println("Orders:", orders)

// 	us := models.UserService{
// 		DB: db,
// 	}
// 	user, err := us.Create("bob4@bob.com", "bob123")
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(user)
// }

// Context的使用
// 在Golang中，不同的数据类型将会被不同看待，如ctxKey和string实际上是不同的类型
// Context的key以type any的形式存储，这允许我们存储任何数据类型，同时，当使用context中的data时
// 这意味着使用context的value时需要将其转化回original type类型
// type ctxKey string

// const (
// 	favoriteColorKey ctxKey = "favorite-color"
// )

// func main() {
// 	ctx := context.Background()

// 	// 报错：尽量定义自己的数据类型，以避免使用已经有的数据类型int, string, float
// 	// 以避免在设置key-value对时，键值对被其他应用写入的数据覆盖，因此，可以使用不导出的类型
// 	// 如ctxKey
// 	ctx = context.WithValue(ctx, favoriteColorKey, "blue")
// 	anyValue := ctx.Value(favoriteColorKey)
// 	// ctx = context.WithValue(ctx, "favorite-color", 0xFF0000)
// 	// value := ctx.Value("favorite-color").(string)
// 	stringValue, ok := anyValue.(string)
// 	if !ok {
// 		fmt.Println(anyValue, "is not a string")
// 		return
// 	}
// 	fmt.Println(stringValue, "is a string")
// }

// func main() {
// 	ctx := stdctx.Background()
// 	user := models.User{
// 		Email: "jons@calhoun.io",
// 	}
// 	ctx = context.WithUser(ctx, &user)
// 	retrievedUser := context.User(ctx)
// 	fmt.Println(retrievedUser)
// 	fmt.Println(retrievedUser.Email)
// }

// func main() {
// 	// email := models.Email{
// 	// 	// Comment this out to test the default sender logic we added.
// 	// 	From:    "apiTEST@demomailtrap.com", //域名设置为demomailtrap.com，便于向自己发送电子邮件
// 	// 	To:      "zxyl777@protonmail.com",
// 	// 	Subject: "This is a test email",
// 	// 	// Try sending emails with only one of these two fields set.
// 	// 	Plaintext: "This is the body of the email",
// 	// 	HTML:      `<h1>Hello there buddy!<!/h1><p>This is the email</p><p>Hope you enjoy it</p>`,
// 	// }
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}
// 	host := os.Getenv("SMTP_HOST")
// 	portStr := os.Getenv("SMTP_PORT")
// 	port, err := strconv.Atoi(portStr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	username := os.Getenv("SMTP_USERNAME")
// 	password := os.Getenv("SMTP_PASSWORD")

// 	es := models.NewEmailService(models.SMTPConfig{
// 		Host:     host,
// 		Port:     port,
// 		Username: username,
// 		Password: password,
// 	})
// 	err = es.ForgotPassword("zxyl777@protonmail.com", "https://localhost/reset-pw?token=abc123")
// 	// err := es.Send(email)

// 	// 发送邮件有两种方式
// 	// 第一种是先连接上。然后再发送，适合发送多封邮件
// 	// sender, err := dialer.Dial()
// 	// if err != nil {
// 	// 	//TODO: Handle the error correctly
// 	// 	panic(err)
// 	// }
// 	// defer sender.Close()
// 	// err = mail.Send(sender, msg)
// 	// if err != nil {
// 	// 	//TODO: Handle the error correctly
// 	// 	panic(err)
// 	// }

// 	// 一次成功，这种方式发送一次邮件
// 	// err := dialer.DialAndSend(msg)
// 	if err != nil {
// 		// todo: Handle the error correctly
// 		panic(err)
// 	}
// 	fmt.Println("Email sent")
// }

// 在发送电子邮箱时， DKIM-Signature通常用于消除垃圾邮件和网络钓鱼邮件。

// 记住，如果你正在使用Mailtrap的邮件测试功能，那么电子邮件实际上并不会被发送出去。
// 相反，它会出现在你的Mailtrap账户下的测试收件箱中。你还可以验证邮件的HTML源代码、文本内容以及其他方面的信息。

// func main() {
// 	gs := models.GalleryService{}
// 	fmt.Println(gs.Images(7))
// }

func main() {
	sketchURL := "http://localhost:3000/galleries/7/images/21.jpg"
	resp, err := http.Get(sketchURL) // 用于发起请求
	if err != nil {
		panic(err)
	}
	// fmt.Println(resp)
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}
