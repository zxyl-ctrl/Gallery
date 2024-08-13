package controllers

import (
	"fmt"
	"net/http"
)

const (
	CookieSession = "session"
)

func newCookie(name, value string) *http.Cookie {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",  // Path定义路径，其所有子路径都可以拥有该Cookie
		HttpOnly: true, // 仅HTTP可以访问cookie，防止js访问造成的XSS攻击
	}
	return &cookie
}

func setCookie(w http.ResponseWriter, name, value string) {
	cookie := newCookie(name, value)
	// 不合理的Cookie会被直接丢弃，且SetCookie不会返回错误，这是因为SetCookie函数只向header添加键值对
	http.SetCookie(w, cookie)
}

func readCookie(r *http.Request, name string) (string, error) {
	c, err := r.Cookie(name)
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	return c.Value, nil
}

func deleteCookie(w http.ResponseWriter, name string) {
	// 创建一个空的cookie并赋值
	cookie := newCookie(name, "")
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}
