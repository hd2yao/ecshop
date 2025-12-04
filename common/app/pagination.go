package app

const (
	// DefaultPageSize 默认每页数量
	DefaultPageSize = 10
	// MaxPageSize 最大每页数量
	MaxPageSize = 100
)

// Pagination 分页结构
type Pagination struct {
	Page      int `json:"page"`       // 当前页码
	PageSize  int `json:"page_size"`  // 每页数量
	TotalRows int `json:"total_rows"` // 总记录数
}

// NewPagination 创建分页实例（从请求参数中解析）
// page: 页码，如果 <= 0 则默认为 1
// pageSize: 每页数量，如果 <= 0 则使用默认值，如果 > MaxPageSize 则限制为 MaxPageSize
func NewPagination(page, pageSize int) *Pagination {
	p := &Pagination{}

	// 处理页码
	if page < 1 {
		p.Page = 1
	} else {
		p.Page = page
	}

	// 处理每页数量
	if pageSize < 1 {
		p.PageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	} else {
		p.PageSize = pageSize
	}

	return p
}

// GetPage 获取当前页码
func (p *Pagination) GetPage() int {
	return p.Page
}

// GetPageSize 获取每页数量
func (p *Pagination) GetPageSize() int {
	return p.PageSize
}

// SetTotalRows 设置总记录数
func (p *Pagination) SetTotalRows(total int) {
	p.TotalRows = total
}

// Offset 计算偏移量
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}
