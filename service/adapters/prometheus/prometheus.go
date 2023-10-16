package prometheus

import (
	"runtime/debug"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const (
	labelHandlerName = "handlerName"

	labelTopic = "topic"

	labelVcsRevision = "vcsRevision"
	labelVcsTime     = "vcsTime"
	labelGo          = "go"

	labelPublicKey = "publicKey"

	labelRelayAddress = "relayAddress"
	labelState        = "state"

	labelResult                          = "result"
	labelResultValueSuccess              = "success"
	labelResultValueError                = "error"
	labelResultValueInvalidPointerPassed = "invalidPointerPassed"
)

type Prometheus struct {
	applicationHandlerCallsCounter          *prometheus.CounterVec
	applicationHandlerCallDurationHistogram *prometheus.HistogramVec

	subscriptionQueueLengthGauge           *prometheus.GaugeVec
	numberOfPublicKeyDownloadersGauge      prometheus.Gauge
	numberOfPublicKeyDownloaderRelaysGauge *prometheus.GaugeVec
	relayConnectionStateGauge              *prometheus.GaugeVec
	twitterAPICallsToPostTweetCounter      *prometheus.CounterVec

	registry *prometheus.Registry

	logger logging.Logger
}

func NewPrometheus(logger logging.Logger) (*Prometheus, error) {
	applicationHandlerCallsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "application_handler_calls_total",
			Help: "Total number of calls to application handlers.",
		},
		[]string{labelHandlerName, labelResult},
	)
	applicationHandlerCallDurationHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "application_handler_calls_duration",
			Help: "Duration of calls to application handlers in seconds.",
		},
		[]string{labelHandlerName, labelResult},
	)
	subscriptionQueueLengthGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "subscription_queue_length",
			Help: "Number of events in the subscription queue.",
		},
		[]string{labelTopic},
	)
	versionGague := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "version",
			Help: "This metric exists just to put a commit label on it.",
		},
		[]string{labelVcsRevision, labelVcsTime, labelGo},
	)
	numberOfPublicKeyDownloadersGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "public_key_downloader_count",
			Help: "Number of running public key downloaders.",
		},
	)
	numberOfPublicKeyDownloaderRelaysGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "public_key_downloader_relays_count",
			Help: "Number of relays for a public key downloader.",
		},
		[]string{labelPublicKey},
	)
	relayConnectionStateGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "relay_connection_state",
			Help: "State of relay connection.",
		},
		[]string{labelRelayAddress, labelState},
	)
	twitterAPICallsToPostTweetCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "twitter_api_calls_to_post_tweet",
			Help: "Total number of calls to Twitter API to post tweets.",
		},
		[]string{labelResult},
	)

	reg := prometheus.NewRegistry()
	for _, v := range []prometheus.Collector{
		applicationHandlerCallsCounter,
		applicationHandlerCallDurationHistogram,
		subscriptionQueueLengthGauge,
		versionGague,
		numberOfPublicKeyDownloadersGauge,
		numberOfPublicKeyDownloaderRelaysGauge,
		relayConnectionStateGauge,
		twitterAPICallsToPostTweetCounter,

		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	} {
		if err := reg.Register(v); err != nil {
			return nil, errors.Wrap(err, "error registering a collector")
		}
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		var vcsRevision, vcsTime string
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.revision" {
				vcsRevision = setting.Value
			}
			if setting.Key == "vcs.time" {
				vcsTime = setting.Value
			}
		}
		versionGague.With(prometheus.Labels{labelGo: buildInfo.GoVersion, labelVcsRevision: vcsRevision, labelVcsTime: vcsTime}).Set(1)
	}

	return &Prometheus{
		applicationHandlerCallsCounter:          applicationHandlerCallsCounter,
		applicationHandlerCallDurationHistogram: applicationHandlerCallDurationHistogram,

		subscriptionQueueLengthGauge:           subscriptionQueueLengthGauge,
		numberOfPublicKeyDownloadersGauge:      numberOfPublicKeyDownloadersGauge,
		numberOfPublicKeyDownloaderRelaysGauge: numberOfPublicKeyDownloaderRelaysGauge,
		relayConnectionStateGauge:              relayConnectionStateGauge,
		twitterAPICallsToPostTweetCounter:      twitterAPICallsToPostTweetCounter,

		registry: reg,

		logger: logger.New("prometheus"),
	}, nil
}

func (p *Prometheus) Registry() *prometheus.Registry {
	return p.registry
}

func (p *Prometheus) StartApplicationCall(handlerName string) app.ApplicationCall {
	return NewApplicationCall(p, handlerName, p.logger)
}

func (p *Prometheus) ReportNumberOfPublicKeyDownloaders(n int) {
	p.numberOfPublicKeyDownloadersGauge.Set(float64(n))
}

func (p *Prometheus) ReportNumberOfPublicKeyDownloaderRelays(publicKey domain.PublicKey, n int) {
	p.numberOfPublicKeyDownloaderRelaysGauge.With(prometheus.Labels{labelPublicKey: publicKey.Hex()}).Set(float64(n))
}

func (p *Prometheus) ReportRelayConnectionState(relayAddress domain.RelayAddress, state app.RelayConnectionState) {
	switch state {
	case app.RelayConnectionStateDisconnected:
		p.relayConnectionStateGauge.With(prometheus.Labels{labelRelayAddress: relayAddress.String(), labelState: app.RelayConnectionStateInitializing.String()}).Set(0)
		p.relayConnectionStateGauge.With(prometheus.Labels{labelRelayAddress: relayAddress.String(), labelState: app.RelayConnectionStateConnected.String()}).Set(0)
	case app.RelayConnectionStateConnected:
		p.relayConnectionStateGauge.With(prometheus.Labels{labelRelayAddress: relayAddress.String(), labelState: app.RelayConnectionStateInitializing.String()}).Set(0)
		p.relayConnectionStateGauge.With(prometheus.Labels{labelRelayAddress: relayAddress.String(), labelState: app.RelayConnectionStateDisconnected.String()}).Set(0)
	case app.RelayConnectionStateInitializing:
		p.relayConnectionStateGauge.With(prometheus.Labels{labelRelayAddress: relayAddress.String(), labelState: app.RelayConnectionStateDisconnected.String()}).Set(0)
		p.relayConnectionStateGauge.With(prometheus.Labels{labelRelayAddress: relayAddress.String(), labelState: app.RelayConnectionStateConnected.String()}).Set(0)
	}
	p.relayConnectionStateGauge.With(prometheus.Labels{labelRelayAddress: relayAddress.String(), labelState: state.String()}).Set(1)
}

func (p *Prometheus) ReportCallingTwitterAPIToPostATweet(err error) {
	var labels prometheus.Labels
	if err == nil {
		labels = prometheus.Labels{labelResult: labelResultValueSuccess}
	} else {
		labels = prometheus.Labels{labelResult: labelResultValueError}
	}
	p.twitterAPICallsToPostTweetCounter.With(labels).Inc()
}

func (p *Prometheus) ReportSubscriptionQueueLength(topic string, n int) {
	p.subscriptionQueueLengthGauge.With(prometheus.Labels{labelTopic: topic}).Set(float64(n))
}

type ApplicationCall struct {
	handlerName string
	p           *Prometheus
	start       time.Time
	logger      logging.Logger
}

func NewApplicationCall(p *Prometheus, handlerName string, logger logging.Logger) *ApplicationCall {
	return &ApplicationCall{
		p:           p,
		handlerName: handlerName,
		logger:      logger,
		start:       time.Now(),
	}
}

func (a *ApplicationCall) End(err *error) {
	duration := time.Since(a.start)

	l := a.logger.
		WithField("handlerName", a.handlerName).
		WithField("duration", duration)

	if err == nil {
		l.Error().Message("application call with an invalid error pointer")
	} else {
		l.Debug().WithError(*err).Message("application call")
	}

	labels := a.getLabels(err)
	a.p.applicationHandlerCallsCounter.With(labels).Inc()
	a.p.applicationHandlerCallDurationHistogram.With(labels).Observe(duration.Seconds())
}

func (a *ApplicationCall) getLabels(err *error) prometheus.Labels {
	labels := prometheus.Labels{
		labelHandlerName: a.handlerName,
	}

	if err == nil {
		labels[labelResult] = labelResultValueInvalidPointerPassed
	} else {
		if *err == nil {
			labels[labelResult] = labelResultValueSuccess
		} else {
			labels[labelResult] = labelResultValueError
		}
	}

	return labels
}
