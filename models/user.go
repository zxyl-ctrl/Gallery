package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"golang.org/x/crypto/bcrypt"
)

// 最好使用uint,有两倍以上的量程
type User struct {
	ID           int
	Email        string
	PasswordHash string
}

// 为了创建数据库连接并存储和读取数据，需要数据库连接，电子邮件，密码
// 关于数据库连接有两种方法：第一种是将*sql.DB作为函数的参数并于其他数据库交互；
// 第二种方法是创建一个带有*sql.DB的结构，并添加与数据库进行交互的方法

// Method1
// func CreateUser(db *sql.DB, email, password string) (*User, error) {
// 	// Create and return the user using 'db'
// }

// Method2
// type UserService struct {
// 	DB *sql.DB
// }
// func (us *UserService) Create(email, password string) (*User, error) {
// 	// Create and return the user using `us.DB`
// }

type UserService struct {
	DB *sql.DB
}

func (us *UserService) Create(email, password string) (*User, error) {
	// Create and return the user using `us.DB`
	email = strings.ToLower(email)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	passwordHash := string(hashedBytes)

	user := User{
		Email:        email,
		PasswordHash: passwordHash,
	}
	row := us.DB.QueryRow(`
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2) RETURNING id`, email, passwordHash)
	err = row.Scan(&user.ID)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				return nil, ErrEmailTaken
			}
		}
		// fmt.Printf("Type = %T\n", err) // 由于错误的包装，可以使用%T来获得错误类型
		// fmt.Printf("Error = %v\n", err)
		return nil, fmt.Errorf("Create user: %w", err)
	}
	return &user, nil
}

// 当创建或者更新的资源与models存储的不同时，有以下选择，为那个操作创建一个新的struct，编写需要的字段
// 就是函数中传递字段的区别
// type NewUser struct {
// 	Email    string
// 	Password string
// }
// func (us *UserService) Create(nu NewUser) (*User, error) {} // 适合保留接口用作拓展
// func (us *UserService) Create(email, password string) (*User, error) {} // 适合创建零散的数据

func (us UserService) Authenticate(email, password string) (*User, error) {
	email = strings.ToLower(email)
	user := User{
		Email: email,
	}
	row := us.DB.QueryRow(`
	SELECT id, password_hash
	FROM users WHERE email=$1`, email)
	err := row.Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	return &user, nil
}

//1. 数据标准化：清理数据使它在每一刻都能够保证相同

func (us *UserService) UpdatePassword(userID int, password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	passwordHash := string(hashedBytes)
	_, err = us.DB.Exec(`
		UPDATE users
		SET password_hash = $2
		WHERE id = $1`, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}
