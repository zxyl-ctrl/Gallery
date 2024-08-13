package migrations

import "embed"

//go:embed *.sql
var FS embed.FS

// 通过嵌入，可以将所有文件移动到./目录(顶级目录下)
// 注意，这个嵌入//go ...之间不能够有空格//和go之间
