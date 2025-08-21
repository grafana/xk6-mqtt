package mqtt

import (
	_ "embed"
	"io"
	"os"
	"testing"

	"github.com/grafana/xk6-mqtt/internal/broker"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.k6.io/k6/js/modules"
	"go.k6.io/k6/js/modulestest"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/metrics"
)

func Test_module(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t)

	root := new(rootModule)
	mod := root.NewModuleInstance(runtime.VU)

	exports := mod.Exports()
	require.NotNil(t, exports)

	require.Nil(t, exports.Default)
	require.Contains(t, exports.Named, "Client")
}

type assertRootModule struct {
	tb testing.TB
}

func newAssertRoot(tb testing.TB) *assertRootModule {
	tb.Helper()

	return &assertRootModule{tb: tb}
}

func (r *assertRootModule) NewModuleInstance(_ modules.VU) modules.Instance {
	return &assertModule{instance: require.New(r.tb)}
}

type assertModule struct {
	instance *require.Assertions
}

func (m *assertModule) Exports() modules.Exports {
	return modules.Exports{
		Default: m.instance,
	}
}

func newTestRuntime(t *testing.T) *modulestest.Runtime {
	t.Helper()

	runtime := modulestest.NewRuntime(t)
	runtime.BuiltinMetrics = metrics.RegisterBuiltinMetrics(runtime.VU.InitEnvField.Registry)
	runtime.VU.InitEnvField.BuiltinMetrics = runtime.BuiltinMetrics

	err := runtime.SetupModuleSystem(
		map[string]any{
			ImportPath:    new(rootModule),
			"k6/x/assert": newAssertRoot(t),
		},
		nil,
		nil,
	)

	require.NoError(t, err)

	env := map[string]string{
		broker.EnvBrokerAddress: os.Getenv(broker.EnvBrokerAddress),
	}

	require.NoError(t, runtime.VU.Runtime().Set("__ENV", env))

	return runtime
}

func newTestVUState(t *testing.T) *lib.State {
	t.Helper()

	samples := make(chan metrics.SampleContainer, 1000)

	t.Cleanup(func() {
		// close(samples)
	})

	registry := metrics.NewRegistry()

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.Out = io.Discard

	return &lib.State{
		Options: lib.Options{
			SystemTags: &metrics.DefaultSystemTagSet,
		},
		Samples:        samples,
		BuiltinMetrics: metrics.RegisterBuiltinMetrics(registry),
		Tags:           lib.NewVUStateTags(registry.RootTagSet()),
		Logger:         logger,
	}
}
