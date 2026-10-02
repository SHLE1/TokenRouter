package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/TokenFlux/TokenRouter/internal/gateway/requeststate"
	"github.com/stretchr/testify/require"
)

// TestHealthObservationKeepsRawAndEffectiveModel 检查空模型与缺省模型的区别，供应商解析使用原始模型，冷却使用规范模型。
func TestHealthObservationKeepsRawAndEffectiveModel(t *testing.T) {
	ctx := requeststate.WithHealthModel(context.Background(), []string{"stored-model"})
	ctx = requeststate.WithOpenAIImagesEndpoint(requeststate.WithThinkingEnabled(ctx, false))
	for _, models := range [][]string{nil, {""}, {" mapped-model "}} {
		input := HealthObservationFromContext(ctx, http.StatusNotFound, nil, []byte("failure"), models)
		require.Equal(t, len(models) > 0, input.ModelProvided)
		require.Equal(t, requeststate.HealthModel(ctx, models), input.EffectiveModel)
		if len(models) > 0 {
			require.Equal(t, models[0], input.Model)
		}
		require.True(t, input.ImagesEndpoint)
		require.NotNil(t, input.Thinking)
		require.False(t, *input.Thinking)
	}
}
