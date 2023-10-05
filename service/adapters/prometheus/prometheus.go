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
	labelHandlerName          = "handlerName"
	labelRelayDownloaderState = "state"
	labelTopic                = "topic"
	labelVcsRevision          = "vcsRevision"
	labelVcsTime              = "vcsTime"
	labelGo                   = "go"

	labelResult                     = "result"
	labelResultSuccess              = "success"
	labelResultError                = "error"
	labelResultInvalidPointerPassed = "invalidPointerPassed"
)

type Prometheus struct {
	applicationHandlerCallsCounter          *prometheus.CounterVec
	applicationHandlerCallDurationHistogram *prometheus.HistogramVec
	relayDownloaderStateGauge               *prometheus.GaugeVec
	subscriptionQueueLengthGauge            *prometheus.GaugeVec

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
	relayDownloaderStateGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "relay_downloader_count",
			Help: "Number of running relay downloaders.",
		},
		[]string{labelRelayDownloaderState},
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

	reg := prometheus.NewRegistry()
	for _, v := range []prometheus.Collector{
		applicationHandlerCallsCounter,
		applicationHandlerCallDurationHistogram,
		relayDownloaderStateGauge,
		subscriptionQueueLengthGauge,
		versionGague,
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
		relayDownloaderStateGauge:               relayDownloaderStateGauge,
		subscriptionQueueLengthGauge:            subscriptionQueueLengthGauge,

		registry: reg,

		logger: logger.New("prometheus"),
	}, nil
}

func (p *Prometheus) StartApplicationCall(handlerName string) app.ApplicationCall {
	return NewApplicationCall(p, handlerName, p.logger)
}

//func (p *Prometheus) MeasureRelayDownloadersState(n int, state app.RelayConnectionState) {
//	p.relayDownloaderStateGauge.With(prometheus.Labels{labelRelayDownloaderState: state.String()}).Set(float64(n))
//}

func (p *Prometheus) ReportSubscriptionQueueLength(topic string, n int) {
	p.subscriptionQueueLengthGauge.With(prometheus.Labels{labelTopic: topic}).Set(float64(n))
}

func (p *Prometheus) Registry() *prometheus.Registry {
	return p.registry
}

func (p *Prometheus) ReportNumberOfPublicKeyDownloaders(n int) {
	//TODO implement me
}

func (p *Prometheus) ReportNumberOfPublicKeyDownloaderRelays(publicKey domain.PublicKey, n int) {
	//TODO implement me
}

func (p *Prometheus) ReportRelayConnectionState(relayAddress domain.RelayAddress, state app.RelayConnectionState) {
	//TODO implement me
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
		labels[labelResult] = labelResultInvalidPointerPassed
	} else {
		if *err == nil {
			labels[labelResult] = labelResultSuccess
		} else {
			labels[labelResult] = labelResultError
		}
	}

	return labels
}
