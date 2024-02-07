package gin

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func handler1(c *gin.Context) {}
func handler2(c *gin.Context) {}
func handler3(c *gin.Context) {}

func TestGetFuncName(t *testing.T) {
	type args struct {
		handlers []gin.HandlerFunc
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should return the function name of the last handler",
			args: args{
				handlers: []gin.HandlerFunc{handler1, handler2, handler3},
			},
			want: "handler3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFuncName(tt.args.handlers...); got != tt.want {
				t.Errorf("getFuncName() = %v, want %v", got, tt.want)
			}
		})
	}
}
