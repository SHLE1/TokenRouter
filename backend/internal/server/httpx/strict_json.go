package httpx

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// BindJSONStrict 在本次绑定中拒绝未知字段。
// credentials、extra 等显式 map 仍由所属领域校验，结构体字段继续使用原 binding 标签。
func BindJSONStrict(c *gin.Context, target any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("request must contain one JSON object")
	}
	return binding.Validator.ValidateStruct(target)
}
