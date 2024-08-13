package models

import (
	"Gallery/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
)

type Session struct {
	ID     int
	UserID int

	// Token is only set when creating a new session. When lookup a session
	// this will be left empty, as we only store the hash of a session token
	// in our database and we cannot reverse it into a raw token
	Token     string // 这个值不进行存储，否则攻击者会获取该值并伪造用户
	TokenHash string
}

type SessionService struct {
	DB *sql.DB
	// BytesPerToken is used to determine how many bytes to use when generating
	// each  session token. If this value is not set or is less than the
	// MinBytesPerToken const it will be ignored and MinBytesPerToken will be
	// used
	BytesPerToken int
}

const (
	// The minimum number of bytes to be used for each session token.
	MinBytesPerToken = 32
)

// Create will create a new session for the user provided. The session token
// will be returned as the Token field on the Session type, but only the hashed
// session token is stored in the database.

// 不使用ON CONFLICT的Create函数
// func (ss *SessionService) Create(userID int) (*Session, error) {
// 	bytesPerToken := ss.BytesPerToken
// 	if bytesPerToken < MinBytesPerToken {
// 		bytesPerToken = MinBytesPerToken
// 	}
// 	token, err := rand.String(bytesPerToken)
// 	if err != nil {
// 		return nil, fmt.Errorf("create: %w", err)
// 	}
// 	session := Session{
// 		UserID:    userID,
// 		Token:     token,
// 		TokenHash: ss.hash(token),
// 	}
// 	// 更新session，如果不存在，则创建
// 	row := ss.DB.QueryRow(`
// 		UPDATE sessions
// 		SET token_hash = $2
// 		WHERE user_id = $1
// 		RETURNING id;`, session.UserID, session.TokenHash)
// 	err = row.Scan(&session.ID)
// 	if err == sql.ErrNoRows {
// 		// If no session exists, we will get ErrNo Rows. That means we need to
// 		// create a session object for that user.
// 		row = ss.DB.QueryRow(`
// 			INSERT INTO sessions (user_id, token_hash)
// 			VALUES ($1, $2)
// 			RETURNING id;`, session.UserID, session.TokenHash)
// 		//The error will be overwritten with either a new error,or nil
// 		err = row.Scan(&session.ID)
// 	}
// 	if err != nil {
// 		return nil, fmt.Errorf("create: %w", err)
// 	}

// 	return &session, nil
// }

func (ss *SessionService) Create(userID int) (*Session, error) {
	bytesPerToken := ss.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	session := Session{
		UserID:    userID,
		Token:     token,
		TokenHash: ss.hash(token),
	}
	row := ss.DB.QueryRow(`
	INSERT INTO sessions (user_id, token_hash)
	VALUES ($1, $2) ON CONFLICT (user_id) DO
	UPDATE
	SET token_hash = $2
	RETURNING id;`, session.UserID, session.TokenHash)
	err = row.Scan(&session.ID)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	return &session, nil
}

// 创建Session有三种选项
// (1) 接收一个hashed session token作为参数，直接将其写入数据库
// (2) 接受一个origin session token作为参数，然后使用Create方法hashing它
// (3) 在Create创建session token并hash it，之后返回original session token

// Return a *Session when we query via session token, then let the caller query for the user seperately.
// Query for the user directly from the SessionService
// Query for the user directly from the UserService

// 不使用INNER的版本
// func (ss *SessionService) User(token string) (*User, error) {
// 	// TODO: Querying Users via Session Token
// 	// 1. Hash the session token
// 	tokenHash := ss.hash(token)
// 	// 2. Query for the session with that hash
// 	var userID int
// 	row := ss.DB.QueryRow(`
// 		SELECT user_id
// 		FROM sessions
// 		WHERE token_hash = $1;`, tokenHash)
// 	err := row.Scan(&userID)
// 	if err != nil {
// 		return nil, fmt.Errorf("user: %w", err)
// 	}
// 	// 3. Using the UserID from the session, we need to query for that user
// 	var user User
// 	row = ss.DB.QueryRow(`
// 		SELECT email, password_hash
// 		FROM users WHERE id = $1`, userID)
// 	err = row.Scan(&user.Email, &user.PasswordHash)
// 	if err != nil {
// 		return nil, fmt.Errorf("user: %w", err)
// 	}
// 	// 4. Return the user
// 	return &user, nil
// }

func (ss *SessionService) User(token string) (*User, error) {
	tokenHash := ss.hash(token)
	var user User
	// SELECT 表示
	row := ss.DB.QueryRow(
		`SELECT users.id,users.email,users.password_hash
		FROM sessions
		JOIN users ON users.id = sessions.user_id
		WHERE sessions.token_hash = $1;`, tokenHash)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}
	return &user, nil
}

// 验证用户session的方法：从请求的cookies获得session token；
// 如果一个session token存在，对其hash，如果不存在，则表明用户没有login
// 一旦有了hashed token，我们就会使用相同的token hash查找数据库，这将会帮助我们决定当前的user

// 不建议使用bcrypy下的hash函数，因为它会为为每一个session添加 salts
// 此时，当我们对同一列相同的session token进行hash，由于salt值的不同，将会得到不同的哈希值

// 我们可以在cookie中存储用户的ID，并使用那些ID去查找Session来实现功能。者可以访问到哈希令牌，但是有显著的缺点
// 1. 性能问题：每次都需要通过用户ID去查找会话和相应的哈希令牌，这会增加额外的数据库查询操作，导致处理速度变慢。
// 2. 增加复杂性：这种方法增加了系统的复杂性，因为需要维护用户ID与会话的映射关系，并在每次请求时都进行查找。
// 3. 未来支持多会话的挑战：如果未来需要支持同一用户拥有多个活动会话，这种方法会变得更加复杂和难以管理。
// 4. 哈希函数不适用：更重要的是，我们为了使用这种方法而选择的哈希函数可能并不适合我们的需求。

// 在这个场景中，同样不适合使用bcrypt进行Hash，原因如下：
// 随机生成：我们的会话令牌是随机生成的，这意味着不需要像密码那样进行慢速哈希来防止暴力破解。
// 每次登录更换：每次用户登录时，会话令牌都会被替换，这进一步降低了使用专为密码设计的哈希函数的必要性。
// 不跨网站重用：会话令牌不像密码那样会在多个网站间重用，因此不需要同样的安全性保障。

// 为什么不使用HMAC
// 虽然HMAC的工作原理是使用一个密钥，类似于给密码添加salts，不同之处在于相同的密钥会被应用到所有需要哈希的值上。
// 一旦应用了这个密钥，就会使用像SHA256这样的哈希算法来进行实际的哈希计算。
// HMAC的主要好处是，如果攻击者能够访问我们的数据库但无法获取到我们的密钥，那么他们甚至无法尝试彩虹表攻击。虽然这听起来不错，但实际上，我们的会话令牌是随机生成的，并且在每次登录时都会被替换，所以彩虹表攻击几乎不可能成功。
// 除此之外，HMAC并没有为我们增加额外的安全性，因为很可能攻击者在能够访问我们数据库的同时也能获取到我们的密钥，这使得HMAC的使用变得毫无意义。

func (ss *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	// base64 encode the data into a string
	return base64.URLEncoding.EncodeToString(tokenHash[:]) // 注意类型转换，接受的类型为slice
}

func (ss *SessionService) Delete(token string) error {
	tokenHash := ss.hash(token)
	_, err := ss.DB.Exec(`
		DELETE FROM sessions
		WHERE token_hash = $1;`, tokenHash) // 使用Exec大的主要原因是我们并不关心返回值
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

// 如何使得用户等出，
// Delete or invalidate the session in the database
// Delete the user's session cookie
