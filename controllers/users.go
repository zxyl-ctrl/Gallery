package controllers

import (
	"Gallery/context"
	"Gallery/models"

	"Gallery/errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Users struct {
	Templates struct { // 用于将用户与渲染模板绑定
		New            Template
		SignIn         Template
		ForgotPassword Template
		CheckYourEmail Template
		ResetPassword  Template
	}
	UserService          *models.UserService
	SessionService       *models.SessionService
	PasswordResetService *models.PasswordResetService
	EmailService         *models.EmailService
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, r, data)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Create(data.Email, data.Password) // 将signup的数据存入数据库
	if err != nil {
		// fmt.Println(err)
		// http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		if errors.Is(err, models.ErrEmailTaken) {
			err = errors.Public(err, "That email address is already associated with an account.")
		}
		u.Templates.New.Execute(w, r, data, err)
		return
	}

	session, err := u.SessionService.Create(user.ID) // 为这个用户创建会话
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound) // 重定向的目的是为了防止用户困惑是否注册成功
		// 虽然在重定向可以使用任何状态码，但是对于许多web浏览器，只有当301 Moved Permanenty和302 Found状态码使用时，才可以被重定向
		// 作为结果，最好使用这两个状态码
		// 举例：301和302举例用于
		return
	}
	setCookie(w, CookieSession, session.Token)          // 设置cookie
	http.Redirect(w, r, "/galleries", http.StatusFound) // 注册成功，直接重定向

	// fmt.Fprintf(w, "User created: %+v", user)

	// err := r.ParseForm() // 解析http.Request传输过来的参数
	// if err != nil {
	// 	http.Error(w, "Unable to pasrse form submission. ", http.StatusBadRequest)
	// 	return
	// }
	// fmt.Fprintf(w, "<p>Email: %s</p>", r.PostForm.Get("email"))
	// fmt.Fprintf(w, "<p>Password: %s</p>", r.PostForm.Get("password"))
	// 类似的，可以使用r.FormValue("email")方法 但是这种方法忽略了ParseForm可能带来的错误
	// 某些情况下，一个key可能有多个value,但r.FormValue只会将第一个value赋值给key
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Authenticate(data.Email, data.Password) // 认证
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	session, err := u.SessionService.Create(user.ID) // 登陆进入创建Session
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

// func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
// 	// tokenCookie, err := r.Cookie("session") // 如果没有查找到，会返回http.ErrNoCookie错误
// 	token, err := readCookie(r, CookieSession)
// 	if err != nil {
// 		fmt.Println(err)
// 		http.Redirect(w, r, "/signin", http.StatusFound)
// 		return
// 	}
// 	user, err := u.SessionService.User(token)
// 	if err != nil {
// 		fmt.Println(err)
// 		http.Redirect(w, r, "/signin", http.StatusFound)
// 		return
// 	}
// 	fmt.Fprintf(w, "Current user :%s\n", user.Email)
// }

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	// 使用RequireUser去重定向了
	// if user == nil {
	// 	http.Redirect(w, r, "/signin", http.StatusFound)
	// 	return
	// }
	fmt.Fprintf(w, "Current user: %s\n", user.Email)
}

// 中间件,自己测试用
func MakeMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h(w, r)
		fmt.Println("Request time:", time.Since(start))
		fmt.Println("IP:", r.RemoteAddr)
	}
}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSession)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = u.SessionService.Delete(token)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	deleteCookie(w, CookieSession)
	http.Redirect(w, r, "/signin", http.StatusFound)
}

// 中间件，与认证相关
type UserMiddleware struct {
	SessionService *models.SessionService
}

// next暗示可能会有多个handler嵌套
func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First try to read the cookie. If we run into an error reading it,
		// proceed with the request. The goal of this middleware isn't to limit
		// access. It only sets the user in the context if it can.
		token, err := readCookie(r, CookieSession)
		if err != nil {
			// Cannot lookup the user with no cookie, so proceed without a user being
			// set, then return.
			next.ServeHTTP(w, r)
			return
		}
		if token == "" { // 处理请求为空的情况
			fmt.Println("The sessionToken is Empty!")
			http.Error(w, "The sessionToken is Empty!", http.StatusBadRequest)
		}
		// If we have a token, try to lookup the user with that token.
		user, err := umw.SessionService.User(token)
		if err != nil {
			// Invalid or expired token. In either case we can still proceed, we just
			// cannot set a user.
			next.ServeHTTP(w, r)
			return
		}
		// If we get to this point, we have a user that we can store in the context!
		// Get the context
		ctx := r.Context()
		// We need to derive a new context to store values in it. Be certain that
		// we import our own context package, and not the one from the standard
		// library.
		ctx = context.WithUser(ctx, user)
		// Next we need to get a request that uses our new context. This is done
		// in a way similar to how contexts work - we call a WithContext function
		// and it returns us a new request with the context set.
		r = r.WithContext(ctx)
		// Finally we call the handler that our middleware was applied to with the
		// updated request.
		next.ServeHTTP(w, r)
	})
}

// 中间件流程
// 1. 通过用户的cookies查找会话令牌
// 2. 使用会话令牌查找有效的会话
// 3. 将会话关联的用户存储在上下文中
// 4. 执行包装的HTTP处理器

// 一些endpoint需要用户登陆后才能够访问，当用户查看他们的画廊列表时，需要知道当前的用户是谁，以便渲染正确的画廊
// 在编辑画廊时，需要确保用户有权访问该画廊，这两种情况下，都必须有用户存在才能处理Web请求
func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// 中间件对于简化HTTP处理器确实很有帮助，但也要注意不要过度使用它们。
// 有时，在HTTP处理器中重复一些代码反而更清晰地表明了测试该处理器所需的要求。与大多数事情一样，这是一个权衡取舍的过程。

func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	// Be sure to revert this back to the ForgotPassword template
	// before proceeding to the next lesson.
	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	pwReset, err := u.PasswordResetService.Create(data.Email)
	if err != nil {
		//TODO: Handle other cases in the future.Forinstance,
		//if a user doesn't exist with the email address.
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	vals := url.Values{ // 设置URL值
		"token": {pwReset.Token},
	}
	// TODO：Make the URL here configurable
	// 这里不能够使用相对路径，因为相对路径是相对于邮箱的
	resetURL := "http://localhost:3000/reset-pw?" + vals.Encode() // 这里不要用https，没有使用openssl无法使用https
	// 这里，可以使用异步调用后台服务来提高效率，如发送电子邮件
	// 选择同步方式的目的：在忘记密码的情况下，我们通常需要确保用户已经成功接收到重置密码的链接或指令，
	// 然后才能继续处理他们的请求。如果我们异步地发送电子邮件，那么可能会存在一种情况：用户收到了一个确认消息，
	// 但实际上他们并没有收到电子邮件，因为他们的邮件服务器可能暂时不可用，或者邮件被误判为垃圾邮件等。这将导致用户体验的混乱和潜在的安全问题。
	// 因此，为了确保用户能够顺利地重置密码，我们在发送电子邮件之后，等待确认发送成功，然后再响应用户的请求。
	// 这样可以确保用户在继续之前已经接收到了必要的信息，从而提高应用程序的可靠性和用户满意度。
	err = u.EmailService.ForgotPassword(data.Email, resetURL)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Somthing went wrong.", http.StatusInternalServerError)
		return
	}
	// Don't render the token here! We need them to confirm they have access to
	// their email to get the token. Sharing it here would be a massive security
	// hole.
	u.Templates.CheckYourEmail.Execute(w, r, data)
}

func (u Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
	}
	data.Token = r.FormValue("token")
	u.Templates.ResetPassword.Execute(w, r, data)
}

// 从URL获得Token和密码后的操作
// Attempt to consume the token
// Update the user's password
// Create a new session
// Sign the user in
// Redirect them to the /users/me page

func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token    string
		Password string
	}
	data.Token = r.FormValue("token")
	data.Password = r.FormValue("password")

	user, err := u.PasswordResetService.Consume(data.Token)
	if err != nil {
		fmt.Println(err)
		// TODO: Distinguish between server errors and invalid token errors.
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	err = u.UserService.UpdatePassword(user.ID, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	// Sign the user in now that they have reset their password.
	// Any errors from this point onward should redirect to the sign in page.
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "users/me", http.StatusFound)
}
