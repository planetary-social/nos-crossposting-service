package prometheus

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/nos-crossposting-service/internal/logging"
	"github.com/planetary-social/nos-crossposting-service/service/adapters"
	"github.com/planetary-social/nos-crossposting-service/service/adapters/twitter"
	"github.com/planetary-social/nos-crossposting-service/service/app"
	"github.com/planetary-social/nos-crossposting-service/service/domain"
	"github.com/planetary-social/nos-crossposting-service/service/domain/accounts"
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

	labelErrorDescription = "errorDescription"

	labelAction               = "action"
	labelActionValuePostTweet = "postTweet"
	labelActionValueGetUser   = "getUser"

	labelAccountID = "accountID"
)

type Prometheus struct {
	applicationHandlerCallsCounter          *prometheus.CounterVec
	applicationHandlerCallDurationHistogram *prometheus.HistogramVec

	subscriptionQueueLengthGauge           *prometheus.GaugeVec
	numberOfPublicKeyDownloadersGauge      prometheus.Gauge
	numberOfPublicKeyDownloaderRelaysGauge *prometheus.GaugeVec
	relayConnectionStateGauge              *prometheus.GaugeVec
	twitterAPICallsCounter                 *prometheus.CounterVec
	purplePagesLookupResultCounter         *prometheus.CounterVec
	tweetCreatedCountPerAccountGauge       *prometheus.GaugeVec
	numberOfAccountsGauge                  prometheus.Gauge
	numberOfLinkedPublicKeysGauge          prometheus.Gauge

	registry *prometheus.Registry

	logger logging.Logger
}

func NewPrometheus(logger logging.Logger) (*Prometheus, error) {
	applicationHandlerCallsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "application_handler_calls_total",
			Help: "Total number of calls to application handlers.",
		},
		[]string{labelHandlerName, labelResult, labelErrorDescription},
	)
	applicationHandlerCallDurationHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "application_handler_calls_duration",
			Help: "Duration of calls to application handlers in seconds.",
		},
		[]string{labelHandlerName, labelResult, labelErrorDescription},
	)
	subscriptionQueueLengthGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "subscription_queue_length",
			Help: "Number of events in the subscription queue.",
		},
		[]string{labelTopic},
	)
	versionGauge := prometheus.NewGaugeVec(
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
	twitterAPICallsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "twitter_api_calls",
			Help: "Total number of calls to Twitter API to post tweets.",
		},
		[]string{labelResult, labelAction, labelErrorDescription},
	)
	purplePagesLookupResultCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "purple_pages_lookups",
			Help: "Number of purple pages lookups.",
		},
		[]string{labelResult, labelErrorDescription, labelRelayAddress},
	)
	tweetCreatedCountPerAccountGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tweet_created_per_account",
			Help: "Tracks number of tweet created events in the queue per account id.",
		},
		[]string{labelAccountID},
	)
	numberOfAccountsGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "accounts_count",
			Help: "Number of accounts.",
		},
	)
	numberOfLinkedPublicKeysGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "linked_public_keys_count",
			Help: "Number of linked public keys.",
		},
	)

	reg := prometheus.NewRegistry()
	for _, v := range []prometheus.Collector{
		applicationHandlerCallsCounter,
		applicationHandlerCallDurationHistogram,
		subscriptionQueueLengthGauge,
		versionGauge,
		numberOfPublicKeyDownloadersGauge,
		numberOfPublicKeyDownloaderRelaysGauge,
		relayConnectionStateGauge,
		twitterAPICallsCounter,
		purplePagesLookupResultCounter,
		tweetCreatedCountPerAccountGauge,
		numberOfAccountsGauge,
		numberOfLinkedPublicKeysGauge,

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
		versionGauge.With(prometheus.Labels{labelGo: buildInfo.GoVersion, labelVcsRevision: vcsRevision, labelVcsTime: vcsTime}).Set(1)
	}

	return &Prometheus{
		applicationHandlerCallsCounter:          applicationHandlerCallsCounter,
		applicationHandlerCallDurationHistogram: applicationHandlerCallDurationHistogram,

		subscriptionQueueLengthGauge:           subscriptionQueueLengthGauge,
		numberOfPublicKeyDownloadersGauge:      numberOfPublicKeyDownloadersGauge,
		numberOfPublicKeyDownloaderRelaysGauge: numberOfPublicKeyDownloaderRelaysGauge,
		relayConnectionStateGauge:              relayConnectionStateGauge,
		twitterAPICallsCounter:                 twitterAPICallsCounter,
		purplePagesLookupResultCounter:         purplePagesLookupResultCounter,
		tweetCreatedCountPerAccountGauge:       tweetCreatedCountPerAccountGauge,
		numberOfAccountsGauge:                  numberOfAccountsGauge,
		numberOfLinkedPublicKeysGauge:          numberOfLinkedPublicKeysGauge,

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

func (p *Prometheus) ReportRelayConnectionState(m map[domain.RelayAddress]app.RelayConnectionState) {
	p.relayConnectionStateGauge.Reset()

	for relayAddress, state := range m {
		p.relayConnectionStateGauge.With(prometheus.Labels{labelRelayAddress: relayAddress.String(), labelState: state.String()}).Set(1)
	}
}

func (p *Prometheus) ReportCallingTwitterAPIToPostATweet(err error) {
	labels := prometheus.Labels{
		labelAction:           labelActionValuePostTweet,
		labelErrorDescription: p.getTwitterErrorDescription(err),
	}
	if err == nil {
		labels[labelResult] = labelResultValueSuccess
	} else {
		labels[labelResult] = labelResultValueError
	}
	p.twitterAPICallsCounter.With(labels).Inc()
}

func (p *Prometheus) ReportCallingTwitterAPIToGetAUser(err error) {
	labels := prometheus.Labels{
		labelAction:           labelActionValueGetUser,
		labelErrorDescription: p.getTwitterErrorDescription(err),
	}
	if err == nil {
		labels[labelResult] = labelResultValueSuccess
	} else {
		labels[labelResult] = labelResultValueError
	}
	p.twitterAPICallsCounter.With(labels).Inc()
}

func (p *Prometheus) ReportSubscriptionQueueLength(topic string, n int) {
	p.subscriptionQueueLengthGauge.With(prometheus.Labels{labelTopic: topic}).Set(float64(n))
}

func (p *Prometheus) ReportPurplePagesLookupResult(address domain.RelayAddress, err *error) {
	labels := prometheus.Labels{
		labelResult:           labelResultValueSuccess,
		labelErrorDescription: "none",
		labelRelayAddress:     address.String(),
	}
	if *err != nil {
		labels[labelResult] = labelResultValueError
		labels[labelErrorDescription] = p.getPurplePagesErrorDescription(*err)
	}
	p.purplePagesLookupResultCounter.With(labels).Inc()
}

func (p *Prometheus) ReportTweetCreatedCountPerAccount(m map[accounts.AccountID]int) {
	p.tweetCreatedCountPerAccountGauge.Reset()

	for accountId, count := range m {
		p.tweetCreatedCountPerAccountGauge.
			With(prometheus.Labels{labelAccountID: accountId.String()}).
			Set(float64(count))
	}
}

func (p *Prometheus) ReportNumberOfAccounts(count int) {
	p.numberOfAccountsGauge.Set(float64(count))
}

func (p *Prometheus) ReportNumberOfLinkedPublicKeys(count int) {
	p.numberOfLinkedPublicKeysGauge.Set(float64(count))
}

func (p *Prometheus) getTwitterErrorDescription(err error) string {
	if err == nil {
		return "none"
	}

	var twitterError twitter.TwitterError
	if errors.As(err, &twitterError) {
		return fmt.Sprintf("twitter/%s", twitterError.Description())
	}

	return "unknown"
}

func (p *Prometheus) getPurplePagesErrorDescription(err error) string {
	if errors.Is(err, adapters.ErrRelayListNotFoundInPurplePages) {
		return "notFound"
	}

	return "unknown"
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
		labels[labelErrorDescription] = "invalidPointer"
	} else {
		if *err == nil {
			labels[labelResult] = labelResultValueSuccess
		} else {
			labels[labelResult] = labelResultValueError
		}
		labels[labelErrorDescription] = a.getErrorDescription(*err)
	}

	return labels
}

func (a *ApplicationCall) getErrorDescription(err error) string {
	if err == nil {
		return "none"
	}

	if errors.Is(err, twitter.ErrExceededLimiterLimit) {
		return "twitter/limiter"
	}

	if errors.Is(err, twitter.TwitterError{}) {
		return "twitter/error"
	}

	return "unknown"
}
