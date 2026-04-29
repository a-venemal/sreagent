package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

// MetricsHandler exposes app metrics in Prometheus exposition format.
// Uses the default gatherer (which includes Go runtime metrics via promhttp's init).
func MetricsHandler(c *gin.Context) {
	gatherer := prometheus.DefaultGatherer
	mfs, err := gatherer.Gather()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	contentType := expfmt.Negotiate(c.Request.Header)
	c.Header("Content-Type", string(contentType))

	enc := expfmt.NewEncoder(c.Writer, contentType)
	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			return
		}
	}
}
