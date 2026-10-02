package composite

import "github.com/TokenFlux/TokenRouter/internal/search"

// ApplySearchAdminReadSettings 将搜索读取结果写入综合快照。
func (s *Snapshot) ApplySearchAdminReadSettings(value *search.AdminReadSettings) {
	s.WebSearchEmulationEnabled = value.WebSearchEmulationEnabled
}
