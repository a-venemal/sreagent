package model

// PrometheusRuleFile represents a Prometheus/VMAlert-compatible rule file.
type PrometheusRuleFile struct {
	Groups []PrometheusRuleGroup `yaml:"groups" json:"groups"`
}

// PrometheusRuleGroup represents a group of alerting rules.
type PrometheusRuleGroup struct {
	Name  string           `yaml:"name" json:"name"`
	Rules []PrometheusRule `yaml:"rules" json:"rules"`
}

// PrometheusRule represents a single Prometheus alerting rule.
type PrometheusRule struct {
	Alert       string            `yaml:"alert" json:"alert"`
	Expr        string            `yaml:"expr" json:"expr"`
	For         string            `yaml:"for" json:"for"`
	Labels      map[string]string `yaml:"labels" json:"labels"`
	Annotations map[string]string `yaml:"annotations" json:"annotations"`
}
