package util

const (
	SINGLE = iota	// 单图
	MULTI			// 多图
	UGOKU			// 动图
)

type Illust struct {
	Name string
	IllustID int
	AuthorID int
	AuthorName string
	Type int
	ImageURL []string
}

func NewIllust() Illust {
	illust := Illust{Name:"", IllustID:0, AuthorID:0, AuthorName:"", Type:SINGLE, ImageURL:make([]string, 1)}
	return illust
}
