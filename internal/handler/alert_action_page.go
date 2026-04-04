package handler

import (
	"fmt"
	"html"
	"strings"

	"github.com/sreagent/sreagent/internal/model"
)

// renderActionPage generates the HTML page for alert actions.
func renderActionPage(event *model.AlertEvent, token, preAction string, preDuration int) string {
	// Build labels display
	var labelsHTML strings.Builder
	for k, v := range event.Labels {
		if k == "alertname" || k == "severity" {
			continue
		}
		labelsHTML.WriteString(fmt.Sprintf(`<span class="label">%s: %s</span>`, html.EscapeString(k), html.EscapeString(v)))
	}
	if labelsHTML.Len() == 0 {
		labelsHTML.WriteString(`<span class="label">无额外标签</span>`)
	}

	// Severity display
	severityClass := "severity-info"
	severityText := string(event.Severity)
	switch event.Severity {
	case model.SeverityCritical:
		severityClass = "severity-critical"
	case model.SeverityWarning:
		severityClass = "severity-warning"
	}

	// Status display
	statusText := string(event.Status)
	statusMap := map[model.AlertEventStatus]string{
		model.EventStatusFiring:       "告警中",
		model.EventStatusAcknowledged: "已认领",
		model.EventStatusAssigned:     "已分配",
		model.EventStatusSilenced:     "已静默",
		model.EventStatusResolved:     "已恢复",
		model.EventStatusClosed:       "已关闭",
	}
	if mapped, ok := statusMap[event.Status]; ok {
		statusText = mapped
	}

	// Pre-select action in JavaScript
	preActionJS := ""
	if preAction != "" {
		preActionJS = fmt.Sprintf(`document.addEventListener('DOMContentLoaded', function() {
			var sel = document.getElementById('action');
			if (sel) { sel.value = '%s'; toggleDuration(); }
		});`, html.EscapeString(preAction))
	}

	preDurationValue := ""
	if preDuration > 0 {
		preDurationValue = fmt.Sprintf("%d", preDuration)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>告警操作 - %s</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; color: #333; min-height: 100vh; padding: 16px; }
.container { max-width: 600px; margin: 0 auto; }
.card { background: #fff; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); padding: 20px; margin-bottom: 16px; }
.card-header { display: flex; align-items: center; gap: 8px; margin-bottom: 16px; }
.card-header h2 { font-size: 18px; flex: 1; word-break: break-all; }
.severity-critical { background: #ff4d4f; color: #fff; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.severity-warning { background: #faad14; color: #fff; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.severity-info { background: #1890ff; color: #fff; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.info-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #f0f0f0; font-size: 14px; }
.info-row:last-child { border-bottom: none; }
.info-label { color: #999; min-width: 80px; }
.info-value { text-align: right; word-break: break-all; }
.labels-wrap { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 12px; }
.label { background: #f0f5ff; color: #1890ff; padding: 2px 8px; border-radius: 4px; font-size: 12px; }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 14px; color: #666; margin-bottom: 6px; }
.form-group select, .form-group input, .form-group textarea { width: 100%%; padding: 10px 12px; border: 1px solid #d9d9d9; border-radius: 8px; font-size: 14px; outline: none; transition: border-color 0.2s; }
.form-group select:focus, .form-group input:focus, .form-group textarea:focus { border-color: #1890ff; }
.form-group textarea { resize: vertical; min-height: 60px; }
.btn { width: 100%%; padding: 12px; border: none; border-radius: 8px; font-size: 16px; font-weight: 600; cursor: pointer; transition: opacity 0.2s; }
.btn:active { opacity: 0.8; }
.btn-primary { background: #1890ff; color: #fff; }
.btn-primary:hover { background: #40a9ff; }
.btn:disabled { background: #d9d9d9; cursor: not-allowed; }
#duration-group { display: none; }
.brand { text-align: center; color: #bbb; font-size: 12px; margin-top: 24px; }
</style>
</head>
<body>
<div class="container">
  <div class="card">
    <div class="card-header">
      <span class="%s">%s</span>
      <h2>%s</h2>
    </div>
    <div class="info-row"><span class="info-label">状态</span><span class="info-value">%s</span></div>
    <div class="info-row"><span class="info-label">触发时间</span><span class="info-value">%s</span></div>
    <div class="info-row"><span class="info-label">来源</span><span class="info-value">%s</span></div>
    <div class="info-row"><span class="info-label">触发次数</span><span class="info-value">%d</span></div>
    <div class="labels-wrap">%s</div>
  </div>

  <div class="card">
    <form method="POST" action="/alert-action/%s" id="action-form">
      <div class="form-group">
        <label for="action">选择操作</label>
        <select name="action" id="action" onchange="toggleDuration()" required>
          <option value="">-- 请选择 --</option>
          <option value="acknowledge">认领告警</option>
          <option value="silence">静默告警</option>
          <option value="resolve">标记已解决</option>
          <option value="close">关闭告警</option>
        </select>
      </div>

      <div class="form-group" id="duration-group">
        <label for="duration">静默时长（分钟）</label>
        <input type="number" name="duration" id="duration" value="%s" min="1" max="10080" placeholder="60">
      </div>

      <div class="form-group">
        <label for="operator_name">操作人</label>
        <input type="text" name="operator_name" id="operator_name" placeholder="请输入姓名" required>
      </div>

      <div class="form-group">
        <label for="note">备注</label>
        <textarea name="note" id="note" placeholder="可选备注信息"></textarea>
      </div>

      <button type="submit" class="btn btn-primary" id="submit-btn">提交</button>
    </form>
  </div>

  <div class="brand">SREAgent Alert Platform</div>
</div>

<script>
function toggleDuration() {
  var action = document.getElementById('action').value;
  var dg = document.getElementById('duration-group');
  dg.style.display = action === 'silence' ? 'block' : 'none';
}
%s
document.getElementById('action-form').addEventListener('submit', function() {
  var btn = document.getElementById('submit-btn');
  btn.disabled = true;
  btn.textContent = '提交中...';
});
</script>
</body>
</html>`,
		html.EscapeString(event.AlertName),
		severityClass,
		html.EscapeString(strings.ToUpper(severityText)),
		html.EscapeString(event.AlertName),
		html.EscapeString(statusText),
		html.EscapeString(event.FiredAt.Format("2006-01-02 15:04:05")),
		html.EscapeString(event.Source),
		event.FireCount,
		labelsHTML.String(),
		html.EscapeString(token),
		html.EscapeString(preDurationValue),
		preActionJS,
	)
}

// renderErrorPage generates an error HTML page.
func renderErrorPage(title, message string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s - SREAgent</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; color: #333; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 16px; }
.card { background: #fff; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); padding: 32px; text-align: center; max-width: 400px; width: 100%%; }
.icon { font-size: 48px; margin-bottom: 16px; }
h2 { font-size: 20px; margin-bottom: 8px; color: #ff4d4f; }
p { font-size: 14px; color: #999; line-height: 1.6; }
</style>
</head>
<body>
<div class="card">
  <div class="icon">&#9888;</div>
  <h2>%s</h2>
  <p>%s</p>
</div>
</body>
</html>`,
		html.EscapeString(title),
		html.EscapeString(title),
		html.EscapeString(message),
	)
}

// renderResultPage generates a result HTML page after action execution.
func renderResultPage(success bool, title, detail string) string {
	icon := "&#10004;"
	color := "#52c41a"
	if !success {
		icon = "&#10008;"
		color = "#ff4d4f"
	}

	detailHTML := ""
	if detail != "" {
		detailHTML = fmt.Sprintf(`<p>%s</p>`, html.EscapeString(detail))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s - SREAgent</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; color: #333; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 16px; }
.card { background: #fff; border-radius: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); padding: 32px; text-align: center; max-width: 400px; width: 100%%; }
.icon { font-size: 48px; margin-bottom: 16px; color: %s; }
h2 { font-size: 20px; margin-bottom: 8px; }
p { font-size: 14px; color: #999; line-height: 1.6; margin-top: 8px; }
.back-link { display: inline-block; margin-top: 16px; color: #1890ff; text-decoration: none; font-size: 14px; }
</style>
</head>
<body>
<div class="card">
  <div class="icon">%s</div>
  <h2>%s</h2>
  %s
  <a href="javascript:history.back()" class="back-link">返回</a>
</div>
</body>
</html>`,
		html.EscapeString(title),
		color,
		icon,
		html.EscapeString(title),
		detailHTML,
	)
}
