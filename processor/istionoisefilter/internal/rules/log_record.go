package rules

import (
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

type logAttrs struct {
	kymaModule     string
	serverAddress  string
	httpMethod     string
	httpDirection  string
	userAgent      string
	urlPath        string
	namespace      string
	deploymentName string
	daemonsetName  string
}

func extractLogAttrs(log plog.LogRecord, resourceAttrs pcommon.Map) logAttrs {
	attrs := log.Attributes()

	return logAttrs{
		kymaModule:     getStringAttrOrEmpty(attrs, "kyma.module"),
		serverAddress:  getStringAttrOrEmpty(attrs, "server.address"),
		httpMethod:     getStringAttrOrEmpty(attrs, "http.request.method"),
		httpDirection:  getStringAttrOrEmpty(attrs, "http.direction"),
		userAgent:      getStringAttrOrEmpty(attrs, "user_agent.original"),
		urlPath:        getStringAttrOrEmpty(attrs, "url.path"),
		namespace:      getStringAttrOrEmpty(resourceAttrs, "k8s.namespace.name"),
		deploymentName: getStringAttrOrEmpty(resourceAttrs, "k8s.deployment.name"),
		daemonsetName:  getStringAttrOrEmpty(resourceAttrs, "k8s.daemonset.name"),
	}
}

func ShouldDropLogRecord(log plog.LogRecord, resourceAttrs pcommon.Map) bool {
	attrs := extractLogAttrs(log, resourceAttrs)

	// a magic attribute that indicates that is an Istio proxy access log
	if attrs.kymaModule != "istio" {
		return false
	}

	switch {
	case isTelemetryMouduleComponentAccessLog(attrs):
		return true
	case regexTelemetryGatewayHost.MatchString(attrs.serverAddress):
		return true
	case isMetricScrapeAccessLog(attrs):
		return true
	case isAvailabilityServiceProbeAccessLog(attrs):
		return true
	default:
		return false
	}
}

func isTelemetryMouduleComponentAccessLog(attrs logAttrs) bool {
	if attrs.namespace != "kyma-system" {
		return false
	}

	if _, found := telemetryModuleAgents[attrs.daemonsetName]; found {
		return true
	}

	if _, found := telemetryModuleGateways[attrs.deploymentName]; found {
		return true
	}

	return false
}

func isAvailabilityServiceProbeAccessLog(attrs logAttrs) bool {
	if attrs.httpMethod != "GET" {
		return false
	}

	if attrs.httpDirection != "outbound" {
		return false
	}

	return strings.HasPrefix(attrs.serverAddress, healthzHostPrefix) && strings.HasSuffix(attrs.urlPath, healthzPath)
}

func isMetricScrapeAccessLog(attrs logAttrs) bool {
	if attrs.httpMethod != "GET" {
		return false
	}

	if attrs.httpDirection != "inbound" {
		return false
	}

	return isRMAUserAgent(attrs.userAgent) || isMetricAgentUserAgent(attrs.userAgent)
}
