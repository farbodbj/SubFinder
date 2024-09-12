package v2rayprobe

import (
	"ConfigProbe/pkg/v2rayprobe/litespeedtest/web"
	"context"
	"github.com/xxf098/lite-proxy/web/render"
	"net/url"
	"runtime"
	"time"
)

type V2rayProbe interface {
	TestV2RayPing(link string, sortDesc bool, timeout time.Duration) (render.Nodes, error)
	TestV2RaySpeed(link string, sortDesc bool, timeout time.Duration) (render.Nodes, error)
	TestV2RayComplete(link string, sortDesc bool, timeout time.Duration) (render.Nodes, error)
}

type v2rayProbeImpl struct {
	concurrency ConcurrencyOpt
	outputMode  OutputMode
}

func NewV2rayProbe(concurrency ConcurrencyOpt, outputMode OutputMode) V2rayProbe {
	return &v2rayProbeImpl{
		concurrency: concurrency,
		outputMode:  outputMode,
	}
}

func (v2ray *v2rayProbeImpl) TestV2RayComplete(link string, sortDesc bool, timeout time.Duration) (render.Nodes, error) {
	return v2ray.doTestWithOpts(Full, link, sortDesc, timeout)
}

func (v2ray *v2rayProbeImpl) TestV2RaySpeed(link string, sortDesc bool, timeout time.Duration) (render.Nodes, error) {
	return v2ray.doTestWithOpts(Speed, link, sortDesc, timeout)
}

func (v2ray *v2rayProbeImpl) TestV2RayPing(link string, sortDesc bool, timeout time.Duration) (render.Nodes, error) {
	return v2ray.doTestWithOpts(Ping, link, sortDesc, timeout)
}

func (v2ray *v2rayProbeImpl) doTestWithOpts(probingMethod ProbeMethod, link string, sortDesc bool, timeout time.Duration) (render.Nodes, error) {
	ctx := context.Background()

	speedTestMode := Map(probingMethod, probeMethodToSpeedTestModeMapper)
	opts := v2ray.createProfileTestOptions(
		link,
		speedTestMode,
		PingMethod(GOOGLE_PING),
		Map2(speedTestMode, sortDesc, speedTestModeToSortMethodMapper),
		v2ray.concurrency,
		v2ray.outputMode,
		timeout,
	)
	return v2ray.doTest(ctx, opts)
}

func probeMethodToSpeedTestModeMapper(probingMethod ProbeMethod) SpeedTestMode {
	switch probingMethod {
	case Speed:
		return SPEED_ONLY
	case Ping:
		return PING_ONLY
	case Full:
		return FULL_TEST
	default:
		return SPEED_ONLY
	}
}

func speedTestModeToSortMethodMapper(speedTestMode SpeedTestMode, isDesc bool) SortMethod {
	if isDesc {
		switch speedTestMode {
		case FULL_TEST:
			return SPEED_DESC
		case SPEED_ONLY:
			return SPEED_DESC
		case PING_ONLY:
			return PING_DESC
		default:
			return SPEED_DESC
		}
	} else {
		switch speedTestMode {
		case FULL_TEST:
			return SPEED_ASC
		case SPEED_ONLY:
			return SPEED_ASC
		case PING_ONLY:
			return PING_ASC
		default:
			return SPEED_ASC
		}
	}
}

func (v2ray *v2rayProbeImpl) createProfileTestOptions(
	sublink string,
	speedTestMode SpeedTestMode,
	pingMethode PingMethod,
	sortMode SortMethod,
	concurrency ConcurrencyOpt,
	outputMode OutputMode,
	timeout time.Duration,
) web.ProfileTestOptions {

	return web.ProfileTestOptions{
		GroupName:     "Default",
		SpeedTestMode: string(speedTestMode),
		PingMethod:    string(pingMethode),
		SortMethod:    string(sortMode),
		Concurrency:   v2ray.calculateConcurrency(concurrency),
		TestMode:      2,
		Subscription:  sublink,
		Language:      "en",
		FontSize:      24,
		Theme:         "rainbow",
		Unique:        true,
		Timeout:       timeout,
		OutputMode:    int(outputMode),
	}
}

func (v2ray *v2rayProbeImpl) calculateConcurrency(concurrency ConcurrencyOpt) int {
	if int(concurrency) == -1 {
		return runtime.NumCPU()
	} else {
		return int(concurrency)
	}
}

func (v2ray *v2rayProbeImpl) validateLinkOrPanic(link string) {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		panic(err)
	}
}

func (v2ray *v2rayProbeImpl) doTest(ctx context.Context, opts web.ProfileTestOptions) (render.Nodes, error) {
	nodes, err := web.TestContext(ctx, opts, &web.EmptyMessageWriter{})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}
