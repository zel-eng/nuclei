package compiler

import (
	"context"
	"testing"
	"time"

	"github.com/projectdiscovery/goja"
	"github.com/projectdiscovery/nuclei/v3/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestSessionInjectsProxyAndTimeoutContext(t *testing.T) {
	runtime := goja.New()
	program, err := goja.Compile("test.js", `
		function run() {
			return { success: true };
		}
		run();
	`, false)
	require.NoError(t, err)

	var sawProxy string
	var sawTimeout time.Duration

	opts := &ExecuteOptions{
		ExecutionId: "session-proxy-test",
		ProxyURL:    "http://127.0.0.1:8080",
		TimeoutVariants: &types.Timeouts{
			TcpReadTimeout:             12 * time.Second,
			JsCompilerExecutionTimeout: 5 * time.Second,
		},
		Callback: func(rt *goja.Runtime) error {
			if v, ok := rt.GetContextValue("proxyURL"); ok {
				sawProxy, _ = v.(string)
			}
			if v, ok := rt.GetContextValue("timeoutVariants"); ok {
				if tv, ok := v.(*types.Timeouts); ok && tv != nil {
					sawTimeout = tv.TcpReadTimeout
				}
			}
			return nil
		},
	}

	_, err = executeWithRuntime(context.Background(), runtime, program, NewExecuteArgs(), opts, nil)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:8080", sawProxy)
	require.Equal(t, 12*time.Second, sawTimeout)

	_, ok := runtime.GetContextValue("proxyURL")
	require.False(t, ok, "proxyURL should be cleaned up after execution")
	_, ok = runtime.GetContextValue("timeoutVariants")
	require.False(t, ok, "timeoutVariants should be cleaned up after execution")
}
