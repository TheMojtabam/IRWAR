package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/slipstream-panel/internal/dnstest"
	"github.com/yourusername/slipstream-panel/internal/runner"
	"github.com/yourusername/slipstream-panel/internal/store"
)

var page = `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Slipstream Panel</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;700&family=Vazirmatn:wght@300;400;500;700&display=swap" rel="stylesheet">
<style>
:root {
  --bg:        #0a0c10;
  --surface:   #111318;
  --card:      #161b22;
  --border:    #21262d;
  --border2:   #30363d;
  --green:     #39d353;
  --green-dim: #196127;
  --red:       #f85149;
  --yellow:    #e3b341;
  --blue:      #58a6ff;
  --purple:    #bc8cff;
  --text:      #e6edf3;
  --text2:     #8b949e;
  --text3:     #484f58;
  --mono:      'JetBrains Mono', monospace;
  --sans:      'Vazirmatn', sans-serif;
}

* { box-sizing: border-box; margin: 0; padding: 0; }

body {
  background: var(--bg);
  color: var(--text);
  font-family: var(--sans);
  min-height: 100vh;
  overflow-x: hidden;
}

/* Animated grid background */
body::before {
  content: '';
  position: fixed;
  inset: 0;
  background-image:
    linear-gradient(rgba(57,211,83,0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(57,211,83,0.03) 1px, transparent 1px);
  background-size: 40px 40px;
  pointer-events: none;
  z-index: 0;
}

/* ── HEADER ─────────────────────────────────────── */
header {
  position: relative;
  z-index: 10;
  border-bottom: 1px solid var(--border);
  background: rgba(10,12,16,0.95);
  backdrop-filter: blur(20px);
  padding: 0 2rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-family: var(--mono);
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--green);
  letter-spacing: -0.5px;
}

.logo-icon {
  width: 28px; height: 28px;
  background: var(--green);
  border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
  animation: pulse-glow 2s ease-in-out infinite;
}

@keyframes pulse-glow {
  0%,100% { box-shadow: 0 0 8px rgba(57,211,83,0.4); }
  50%      { box-shadow: 0 0 20px rgba(57,211,83,0.8); }
}

.logo-icon svg { width: 16px; height: 16px; fill: #0a0c10; }

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.bin-badge {
  font-family: var(--mono);
  font-size: 0.72rem;
  padding: 3px 10px;
  border-radius: 20px;
  border: 1px solid var(--border2);
  color: var(--text2);
  background: var(--surface);
}

.bin-badge.ok  { border-color: var(--green-dim); color: var(--green); }
.bin-badge.err { border-color: #6e1e1e; color: var(--red); }

/* ── MAIN ───────────────────────────────────────── */
main {
  position: relative;
  z-index: 1;
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1.5rem 4rem;
}

/* ── STATS BAR ──────────────────────────────────── */
.stats-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 1rem 1.2rem;
  transition: border-color 0.2s;
}

.stat-card:hover { border-color: var(--border2); }

.stat-label {
  font-size: 0.72rem;
  color: var(--text3);
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-bottom: 6px;
}

.stat-value {
  font-family: var(--mono);
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--text);
}

.stat-value.green  { color: var(--green); }
.stat-value.yellow { color: var(--yellow); }
.stat-value.red    { color: var(--red); }

/* ── SECTION HEADER ─────────────────────────────── */
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.section-title {
  font-family: var(--mono);
  font-size: 0.85rem;
  color: var(--text2);
  text-transform: uppercase;
  letter-spacing: 1.5px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title::before {
  content: '';
  display: inline-block;
  width: 3px; height: 14px;
  background: var(--green);
  border-radius: 2px;
}

/* ── BUTTONS ────────────────────────────────────── */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 6px;
  border: 1px solid transparent;
  font-family: var(--sans);
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.btn-primary {
  background: var(--green);
  color: #0a0c10;
  font-weight: 700;
}
.btn-primary:hover { background: #4be366; }

.btn-ghost {
  background: transparent;
  border-color: var(--border2);
  color: var(--text2);
}
.btn-ghost:hover { border-color: var(--text2); color: var(--text); }

.btn-danger {
  background: transparent;
  border-color: #6e1e1e;
  color: var(--red);
}
.btn-danger:hover { background: rgba(248,81,73,0.1); }

.btn-sm { padding: 4px 10px; font-size: 0.75rem; }
.btn-icon { padding: 6px; border-radius: 6px; }

.btn svg { width: 13px; height: 13px; flex-shrink: 0; }

/* ── INSTANCE CARDS ─────────────────────────────── */
.instances-grid {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.instance-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
  transition: border-color 0.2s, transform 0.15s;
}

.instance-card:hover { border-color: var(--border2); transform: translateY(-1px); }
.instance-card.running { border-color: rgba(57,211,83,0.25); }

.card-main {
  padding: 1rem 1.25rem;
  display: grid;
  grid-template-columns: auto 1fr auto auto auto;
  align-items: center;
  gap: 1rem;
}

/* Status dot */
.status-dot {
  width: 10px; height: 10px;
  border-radius: 50%;
  background: var(--text3);
  flex-shrink: 0;
}

.status-dot.running {
  background: var(--green);
  box-shadow: 0 0 6px rgba(57,211,83,0.6);
  animation: blink 2s ease-in-out infinite;
}

.status-dot.stopped { background: var(--text3); }

@keyframes blink {
  0%,100% { opacity: 1; }
  50%      { opacity: 0.4; }
}

/* Instance info */
.inst-info { min-width: 0; }

.inst-name {
  font-weight: 600;
  font-size: 0.95rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 3px;
}

.inst-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.meta-tag {
  font-family: var(--mono);
  font-size: 0.7rem;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--surface);
  color: var(--text2);
  border: 1px solid var(--border);
}

.meta-tag.port { color: var(--blue); border-color: rgba(88,166,255,0.2); background: rgba(88,166,255,0.05); }
.meta-tag.dns  { color: var(--purple); border-color: rgba(188,140,255,0.2); background: rgba(188,140,255,0.05); }
.meta-tag.timer { color: var(--yellow); border-color: rgba(227,179,65,0.2); background: rgba(227,179,65,0.05); }

/* Timer countdown */
.timer-display {
  font-family: var(--mono);
  font-size: 0.78rem;
  color: var(--text3);
  text-align: center;
  min-width: 70px;
}

.timer-display span {
  display: block;
  font-size: 0.65rem;
  color: var(--text3);
  margin-top: 1px;
}

/* Actions */
.card-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

/* Logs panel */
.logs-panel {
  display: none;
  border-top: 1px solid var(--border);
  background: #0d1117;
  padding: 1rem 1.25rem;
}

.logs-panel.open { display: block; }

.logs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.logs-title {
  font-family: var(--mono);
  font-size: 0.72rem;
  color: var(--text3);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.logs-content {
  font-family: var(--mono);
  font-size: 0.73rem;
  color: #8b949e;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 10px 12px;
  max-height: 200px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.6;
}

.logs-content::-webkit-scrollbar { width: 4px; }
.logs-content::-webkit-scrollbar-track { background: transparent; }
.logs-content::-webkit-scrollbar-thumb { background: var(--border2); border-radius: 2px; }

/* ── MODAL ──────────────────────────────────────── */
.modal-backdrop {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.75);
  backdrop-filter: blur(4px);
  z-index: 100;
  align-items: center;
  justify-content: center;
}

.modal-backdrop.open { display: flex; }

.modal {
  background: var(--card);
  border: 1px solid var(--border2);
  border-radius: 14px;
  width: 480px;
  max-width: 95vw;
  max-height: 90vh;
  overflow-y: auto;
  animation: modal-in 0.2s ease;
}

@keyframes modal-in {
  from { transform: scale(0.95) translateY(10px); opacity: 0; }
  to   { transform: scale(1) translateY(0); opacity: 1; }
}

.modal-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.modal-title {
  font-weight: 700;
  font-size: 1rem;
  color: var(--text);
}

.modal-body { padding: 1.5rem; }

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

/* ── FORM ───────────────────────────────────────── */
.form-group { margin-bottom: 1.1rem; }

.form-label {
  display: block;
  font-size: 0.8rem;
  color: var(--text2);
  margin-bottom: 6px;
  font-weight: 500;
}

.form-hint {
  font-size: 0.72rem;
  color: var(--text3);
  margin-top: 4px;
}

.form-input {
  width: 100%;
  background: var(--bg);
  border: 1px solid var(--border2);
  border-radius: 6px;
  padding: 9px 12px;
  color: var(--text);
  font-family: var(--mono);
  font-size: 0.85rem;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
  direction: ltr;
}

.form-input:focus {
  border-color: var(--green);
  box-shadow: 0 0 0 3px rgba(57,211,83,0.1);
}

.form-input::placeholder { color: var(--text3); }

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

/* Toggle */
.toggle-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toggle {
  position: relative;
  width: 40px; height: 22px;
  cursor: pointer;
}

.toggle input { display: none; }

.toggle-track {
  position: absolute;
  inset: 0;
  background: var(--border2);
  border-radius: 11px;
  transition: background 0.2s;
}

.toggle input:checked + .toggle-track { background: var(--green-dim); }

.toggle-thumb {
  position: absolute;
  top: 3px; left: 3px;
  width: 16px; height: 16px;
  background: var(--text2);
  border-radius: 50%;
  transition: transform 0.2s, background 0.2s;
}

.toggle input:checked ~ .toggle-thumb {
  transform: translateX(18px);
  background: var(--green);
}

/* ── EMPTY STATE ────────────────────────────────── */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: var(--text3);
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.3;
}

.empty-state h3 { font-size: 1rem; color: var(--text2); margin-bottom: 0.5rem; }
.empty-state p  { font-size: 0.85rem; }

/* ── TOAST ──────────────────────────────────────── */
.toast-container {
  position: fixed;
  bottom: 1.5rem;
  left: 50%;
  transform: translateX(-50%);
  z-index: 999;
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: center;
}

.toast {
  background: var(--card);
  border: 1px solid var(--border2);
  border-radius: 8px;
  padding: 10px 18px;
  font-size: 0.82rem;
  color: var(--text);
  box-shadow: 0 8px 32px rgba(0,0,0,0.5);
  animation: toast-in 0.25s ease;
  white-space: nowrap;
  font-family: var(--mono);
}

.toast.success { border-color: var(--green-dim); color: var(--green); }
.toast.error   { border-color: #6e1e1e; color: var(--red); }

@keyframes toast-in {
  from { opacity: 0; transform: translateY(10px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* ── RESPONSIVE ─────────────────────────────────── */
/* ── TEST MODAL ─────────────────────────────────────── */
.test-step {
  display: flex; align-items: flex-start; gap: 12px;
  padding: 10px 14px; border-radius: 8px;
  border: 1px solid var(--border); background: var(--surface);
  margin-bottom: 8px; transition: border-color 0.2s;
}
.test-step.ok  { border-color: var(--green-dim); }
.test-step.err { border-color: #6e1e1e; }
.step-icon { font-size: 1.1rem; flex-shrink: 0; margin-top: 1px; }
.step-body { flex: 1; min-width: 0; }
.step-name {
  font-weight: 600; font-size: 0.85rem; margin-bottom: 3px;
  display: flex; align-items: center; justify-content: space-between;
}
.step-query {
  font-family: var(--mono); font-size: 0.72rem; color: var(--text3);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-bottom: 2px;
}
.step-detail { font-size: 0.75rem; color: var(--text2); }
.step-ms {
  font-family: var(--mono); font-size: 0.72rem; padding: 2px 7px;
  border-radius: 4px; background: var(--bg); color: var(--text3); white-space: nowrap;
}
.step-ms.fast   { color: var(--green); }
.step-ms.medium { color: var(--yellow); }
.step-ms.slow   { color: var(--red); }
.test-summary {
  text-align: center; padding: 14px; border-radius: 8px;
  margin-bottom: 14px; font-weight: 700; font-size: 0.95rem;
}
.test-summary.ok      { background: rgba(57,211,83,0.08);  color: var(--green); border: 1px solid var(--green-dim); }
.test-summary.err     { background: rgba(248,81,73,0.08);  color: var(--red);   border: 1px solid #6e1e1e; }
.test-summary.running { background: rgba(88,166,255,0.08); color: var(--blue);  border: 1px solid rgba(88,166,255,0.2); }
.spinner {
  display: inline-block; width: 13px; height: 13px;
  border: 2px solid currentColor; border-top-color: transparent;
  border-radius: 50%; animation: spin 0.7s linear infinite;
  vertical-align: middle; margin-left: 6px;
}
@keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 768px) {
  .stats-bar { grid-template-columns: repeat(2,1fr); }
  .card-main { grid-template-columns: auto 1fr; }
  .card-actions { flex-wrap: wrap; }
  .timer-display { display: none; }
}
</style>
</head>
<body>

<header>
  <div class="logo">
    <div class="logo-icon">
      <svg viewBox="0 0 16 16"><path d="M8 1L14 4V8C14 11.3 11.3 14 8 15C4.7 14 2 11.3 2 8V4L8 1Z"/></svg>
    </div>
    Slipstream Panel
  </div>
  <div class="header-right">
    <div class="bin-badge" id="binBadge">checking…</div>
    <button class="btn btn-primary" onclick="openModal()">
      <svg viewBox="0 0 16 16" fill="currentColor"><path d="M8 2a.75.75 0 01.75.75v4.5h4.5a.75.75 0 010 1.5h-4.5v4.5a.75.75 0 01-1.5 0v-4.5h-4.5a.75.75 0 010-1.5h4.5v-4.5A.75.75 0 018 2z"/></svg>
      اضافه کردن Instance
    </button>
  </div>
</header>

<main>
  <!-- Stats -->
  <div class="stats-bar">
    <div class="stat-card">
      <div class="stat-label">کل Instance ها</div>
      <div class="stat-value" id="statTotal">0</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">در حال اجرا</div>
      <div class="stat-value green" id="statRunning">0</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">متوقف شده</div>
      <div class="stat-value red" id="statStopped">0</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">Auto-Restart</div>
      <div class="stat-value yellow" id="statAuto">0</div>
    </div>
  </div>

  <!-- Instances -->
  <div class="section-header">
    <div class="section-title">Instance ها</div>
    <div style="display:flex;gap:8px">
      <button class="btn btn-ghost btn-sm" onclick="startAll()">▶ همه را شروع کن</button>
      <button class="btn btn-ghost btn-sm" onclick="stopAll()">■ همه را متوقف کن</button>
    </div>
  </div>

  <div class="instances-grid" id="instancesGrid">
    <div class="empty-state">
      <div class="empty-icon">⚡</div>
      <h3>هیچ Instance‌ای وجود ندارد</h3>
      <p>روی دکمه «اضافه کردن» کلیک کنید تا اولین instance را بسازید</p>
    </div>
  </div>
</main>

<!-- Modal -->
<div class="modal-backdrop" id="modal">
  <div class="modal">
    <div class="modal-header">
      <div class="modal-title" id="modalTitle">Instance جدید</div>
      <button class="btn btn-ghost btn-icon btn-sm" onclick="closeModal()">✕</button>
    </div>
    <div class="modal-body">
      <input type="hidden" id="editId">

      <div class="form-group">
        <label class="form-label">نام (اختیاری)</label>
        <input class="form-input" id="fieldName" placeholder="مثلاً: Instance 1">
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">DNS Resolver</label>
          <input class="form-input" id="fieldResolver" placeholder="1.1.1.1" value="1.1.1.1">
          <div class="form-hint">مثلاً: 1.1.1.1 یا 8.8.8.8</div>
        </div>
        <div class="form-group">
          <label class="form-label">پورت SOCKS</label>
          <input class="form-input" id="fieldPort" type="number" placeholder="1080" value="1080">
          <div class="form-hint">پورت پروکسی لوکال</div>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">Domain (DNS Tunnel)</label>
        <input class="form-input" id="fieldDomain" placeholder="d.fastsoft98.ir">
        <div class="form-hint">مثلاً: d.fastsoft98.ir</div>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">ریست خودکار هر چند دقیقه</label>
          <input class="form-input" id="fieldRestart" type="number" placeholder="30" value="30" min="0">
          <div class="form-hint">0 = غیرفعال</div>
        </div>
        <div class="form-group">
          <div class="toggle-wrapper" style="margin-top:22px">
            <label class="form-label" style="margin:0">Auto-start هنگام بوت</label>
            <label class="toggle">
              <input type="checkbox" id="fieldAutoRestart" checked>
              <div class="toggle-track"></div>
              <div class="toggle-thumb"></div>
            </label>
          </div>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">آرگومان‌های اضافی (اختیاری)</label>
        <input class="form-input" id="fieldExtra" placeholder="--keep-alive-interval 500">
        <div class="form-hint">پارامترهای اضافی برای slipstream</div>
      </div>
    </div>
    <div class="modal-footer">
      <button class="btn btn-ghost" onclick="closeModal()">انصراف</button>
      <button class="btn btn-primary" onclick="saveInstance()">ذخیره</button>
    </div>
  </div>
</div>

<!-- Test Modal -->
<div class="modal-backdrop" id="testModal">
  <div class="modal" style="width:520px">
    <div class="modal-header">
      <div class="modal-title">🔍 تست DNS Handshake</div>
      <button class="btn btn-ghost btn-icon btn-sm" onclick="closeTestModal()">✕</button>
    </div>
    <div class="modal-body">
      <div style="display:flex;gap:8px;margin-bottom:14px">
        <div style="flex:1">
          <div class="form-label">DNS Resolver</div>
          <input class="form-input" id="testResolver" placeholder="1.1.1.1">
        </div>
        <div style="flex:1.5">
          <div class="form-label">Domain</div>
          <input class="form-input" id="testDomain" placeholder="d.fastsoft98.ir">
        </div>
        <div style="display:flex;align-items:flex-end">
          <button class="btn btn-primary" id="testRunBtn" onclick="runTest()">تست</button>
        </div>
      </div>
      <div id="testResult"></div>
    </div>
  </div>
</div>

<div class="toast-container" id="toasts"></div>

<script>
// ── State ─────────────────────────────────────────────────────────────────────
let instances = [];
let countdowns = {}; // id -> interval

// ── API ───────────────────────────────────────────────────────────────────────
async function api(method, path, body) {
  const r = await fetch(path, {
    method,
    headers: body ? {'Content-Type':'application/json'} : {},
    body: body ? JSON.stringify(body) : undefined
  });
  return r.json();
}

// ── Init ──────────────────────────────────────────────────────────────────────
async function init() {
  const s = await api('GET', '/api/status');
  const badge = document.getElementById('binBadge');
  if (s.bin && s.bin !== 'NOT FOUND') {
    badge.textContent = '✓ ' + s.bin.split('/').pop();
    badge.className = 'bin-badge ok';
  } else {
    badge.textContent = '✗ binary not found';
    badge.className = 'bin-badge err';
  }
  await refresh();
  setInterval(refresh, 4000);
}

async function refresh() {
  instances = await api('GET', '/api/instances');
  render();
}

// ── Render ────────────────────────────────────────────────────────────────────
function render() {
  const grid = document.getElementById('instancesGrid');
  const running = instances.filter(i=>i.status==='running').length;
  const stopped = instances.length - running;
  const auto    = instances.filter(i=>i.auto_restart).length;

  document.getElementById('statTotal').textContent   = instances.length;
  document.getElementById('statRunning').textContent = running;
  document.getElementById('statStopped').textContent = stopped;
  document.getElementById('statAuto').textContent    = auto;

  if (instances.length === 0) {
    grid.innerHTML = ` + "`" + `
      <div class="empty-state">
        <div class="empty-icon">⚡</div>
        <h3>هیچ Instance‌ای وجود ندارد</h3>
        <p>روی دکمه «اضافه کردن» کلیک کنید تا اولین instance را بسازید</p>
      </div>` + "`" + `;
    return;
  }

  grid.innerHTML = instances.map(inst => renderCard(inst)).join('');

  // Re-init countdown timers
  instances.forEach(inst => {
    if (inst.status === 'running' && inst.restart_minutes > 0) {
      startCountdown(inst.id, inst.restart_minutes * 60);
    }
  });
}

function renderCard(inst) {
  const isRunning = inst.status === 'running';
  const restartMin = inst.restart_minutes;
  const timerHTML = isRunning && restartMin > 0
    ? ` + "`" + `<div class="timer-display">
         <span>ریست در</span>
         <span id="cd_${inst.id}" style="color:var(--yellow);font-size:0.9rem">--:--</span>
       </div>` + "`" + `
    : ` + "`" + `<div class="timer-display"><span style="opacity:0.3">${restartMin > 0 ? restartMin+'min' : 'no timer'}</span></div>` + "`" + `;

  return ` + "`" + `
  <div class="instance-card ${isRunning ? 'running' : ''}" id="card_${inst.id}">
    <div class="card-main">
      <div class="status-dot ${inst.status}"></div>
      <div class="inst-info">
        <div class="inst-name">${esc(inst.name)}</div>
        <div class="inst-meta">
          <span class="meta-tag port">SOCKS :${inst.socks_port}</span>
          <span class="meta-tag dns">-r ${esc(inst.resolver)}</span>
          <span class="meta-tag dns">-d ${esc(inst.domain)}</span>
          ${restartMin > 0 ? ` + "`" + `<span class="meta-tag timer">↺ ${restartMin}min</span>` + "`" + ` : ''}
        </div>
      </div>
      ${timerHTML}
      <div class="card-actions">
        ${isRunning
          ? ` + "`" + `<button class="btn btn-ghost btn-sm" onclick="doRestart('${inst.id}')">↺ ریست</button>
             <button class="btn btn-danger btn-sm" onclick="doStop('${inst.id}')">■ متوقف</button>` + "`" + `
          : ` + "`" + `<button class="btn btn-primary btn-sm" onclick="doStart('${inst.id}')">▶ شروع</button>` + "`" + `
        }
        <button class="btn btn-ghost btn-icon btn-sm" title="تست DNS" onclick="doTest('${inst.id}')">🔍</button>
        <button class="btn btn-ghost btn-icon btn-sm" title="لاگ‌ها" onclick="toggleLogs('${inst.id}')">📋</button>
        <button class="btn btn-ghost btn-icon btn-sm" title="ویرایش" onclick="editInstance('${inst.id}')">✏️</button>
        <button class="btn btn-danger btn-icon btn-sm" title="حذف" onclick="doDelete('${inst.id}')">🗑</button>
      </div>
    </div>
    <div class="logs-panel" id="logs_${inst.id}">
      <div class="logs-header">
        <span class="logs-title">لاگ‌های خروجی</span>
        <div style="display:flex;gap:6px">
          <button class="btn btn-ghost btn-sm" onclick="loadLogs('${inst.id}')">↻ بارگذاری مجدد</button>
          <button class="btn btn-ghost btn-sm" onclick="clearLogs('${inst.id}')">پاک کردن</button>
        </div>
      </div>
      <div class="logs-content" id="logcontent_${inst.id}">کلیک کنید تا بارگذاری شود…</div>
    </div>
  </div>` + "`" + `;
}

function esc(s) {
  return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

// ── Countdown ─────────────────────────────────────────────────────────────────
function startCountdown(id, totalSec) {
  if (countdowns[id]) clearInterval(countdowns[id]);
  let remaining = totalSec;
  const update = () => {
    const el = document.getElementById('cd_' + id);
    if (!el) { clearInterval(countdowns[id]); return; }
    remaining = Math.max(0, remaining - 4);
    const m = String(Math.floor(remaining / 60)).padStart(2,'0');
    const s = String(remaining % 60).padStart(2,'0');
    el.textContent = m + ':' + s;
  };
  update();
  countdowns[id] = setInterval(update, 4000);
}

// ── Actions ───────────────────────────────────────────────────────────────────
async function doStart(id) {
  const r = await api('POST', ` + "`" + `/api/instances/${id}/start` + "`" + `);
  toast(r.ok ? 'شروع شد ✓' : r.msg, r.ok ? 'success' : 'error');
  await refresh();
}

async function doStop(id) {
  const r = await api('POST', ` + "`" + `/api/instances/${id}/stop` + "`" + `);
  toast(r.ok ? 'متوقف شد' : r.msg, r.ok ? 'success' : 'error');
  await refresh();
}

async function doRestart(id) {
  const r = await api('POST', ` + "`" + `/api/instances/${id}/restart` + "`" + `);
  toast(r.ok ? '↺ ریست شد' : r.msg, r.ok ? 'success' : 'error');
  await refresh();
}

async function doDelete(id) {
  if (!confirm('آیا مطمئن هستید؟')) return;
  await api('DELETE', ` + "`" + `/api/instances/${id}` + "`" + `);
  toast('حذف شد', 'success');
  await refresh();
}

async function startAll() {
  for (const inst of instances) {
    if (inst.status !== 'running') await api('POST', ` + "`" + `/api/instances/${inst.id}/start` + "`" + `);
  }
  toast('همه شروع شدند ✓', 'success');
  await refresh();
}

async function stopAll() {
  for (const inst of instances) {
    if (inst.status === 'running') await api('POST', ` + "`" + `/api/instances/${inst.id}/stop` + "`" + `);
  }
  toast('همه متوقف شدند', 'success');
  await refresh();
}

// ── Logs ──────────────────────────────────────────────────────────────────────
async function toggleLogs(id) {
  const panel = document.getElementById('logs_' + id);
  if (panel.classList.contains('open')) {
    panel.classList.remove('open');
  } else {
    panel.classList.add('open');
    await loadLogs(id);
  }
}

async function loadLogs(id) {
  const el = document.getElementById('logcontent_' + id);
  if (!el) return;
  el.textContent = 'در حال بارگذاری…';
  const r = await api('GET', ` + "`" + `/api/instances/${id}/logs` + "`" + `);
  el.textContent = r.logs || '(خالی)';
  el.scrollTop = el.scrollHeight;
}

async function clearLogs(id) {
  await api('POST', ` + "`" + `/api/instances/${id}/clear_logs` + "`" + `);
  const el = document.getElementById('logcontent_' + id);
  if (el) el.textContent = '(پاک شد)';
  toast('لاگ‌ها پاک شدند', 'success');
}

// ── Modal ─────────────────────────────────────────────────────────────────────
function openModal(id) {
  document.getElementById('editId').value = '';
  document.getElementById('modalTitle').textContent = 'Instance جدید';
  document.getElementById('fieldName').value = '';
  document.getElementById('fieldResolver').value = '1.1.1.1';
  document.getElementById('fieldPort').value = nextPort();
  document.getElementById('fieldDomain').value = '';
  document.getElementById('fieldRestart').value = '30';
  document.getElementById('fieldAutoRestart').checked = true;
  document.getElementById('fieldExtra').value = '';
  document.getElementById('modal').classList.add('open');
}

function editInstance(id) {
  const inst = instances.find(i => i.id === id);
  if (!inst) return;
  document.getElementById('editId').value = id;
  document.getElementById('modalTitle').textContent = 'ویرایش: ' + inst.name;
  document.getElementById('fieldName').value = inst.name;
  document.getElementById('fieldResolver').value = inst.resolver;
  document.getElementById('fieldPort').value = inst.socks_port;
  document.getElementById('fieldDomain').value = inst.domain;
  document.getElementById('fieldRestart').value = inst.restart_minutes;
  document.getElementById('fieldAutoRestart').checked = inst.auto_restart;
  document.getElementById('fieldExtra').value = inst.extra_args || '';
  document.getElementById('modal').classList.add('open');
}

function closeModal() {
  document.getElementById('modal').classList.remove('open');
}

document.getElementById('modal').addEventListener('click', e => {
  if (e.target === e.currentTarget) closeModal();
});

function nextPort() {
  const used = instances.map(i => parseInt(i.socks_port));
  let port = 1080;
  while (used.includes(port)) port++;
  return port;
}

async function saveInstance() {
  const id = document.getElementById('editId').value;
  const body = {
    name:            document.getElementById('fieldName').value || ` + "`" + `Instance ${instances.length+1}` + "`" + `,
    resolver:        document.getElementById('fieldResolver').value.trim() || '1.1.1.1',
    domain:          document.getElementById('fieldDomain').value.trim(),
    socks_port:      parseInt(document.getElementById('fieldPort').value) || 1080,
    restart_minutes: parseInt(document.getElementById('fieldRestart').value) || 0,
    auto_restart:    document.getElementById('fieldAutoRestart').checked,
    extra_args:      document.getElementById('fieldExtra').value.trim(),
  };

  if (!body.domain) { toast('Domain الزامی است', 'error'); return; }

  let r;
  if (id) {
    r = await api('PUT', ` + "`" + `/api/instances/${id}` + "`" + `, body);
  } else {
    r = await api('POST', '/api/instances', body);
  }

  if (r.error) { toast(r.error, 'error'); return; }
  toast(id ? 'ذخیره شد ✓' : 'ساخته شد ✓', 'success');
  closeModal();
  await refresh();
}

// ── Toast ─────────────────────────────────────────────────────────────────────
function toast(msg, type='success') {
  const el = document.createElement('div');
  el.className = ` + "`" + `toast ${type}` + "`" + `;
  el.textContent = msg;
  document.getElementById('toasts').appendChild(el);
  setTimeout(() => el.remove(), 3000);
}

// ── DNS Test ──────────────────────────────────────────────────────────────────
function doTest(id) {
  const inst = instances.find(i => i.id === id);
  if (!inst) return;
  document.getElementById('testResolver').value = inst.resolver;
  document.getElementById('testDomain').value   = inst.domain;
  document.getElementById('testResult').innerHTML = '';
  document.getElementById('testModal').classList.add('open');
}

function closeTestModal() {
  document.getElementById('testModal').classList.remove('open');
}

document.getElementById('testModal').addEventListener('click', e => {
  if (e.target === e.currentTarget) closeTestModal();
});

async function runTest() {
  const resolver = document.getElementById('testResolver').value.trim();
  const domain   = document.getElementById('testDomain').value.trim();
  if (!resolver || !domain) { toast('resolver و domain رو پر کن', 'error'); return; }

  const btn = document.getElementById('testRunBtn');
  btn.disabled = true;
  btn.textContent = '...';

  document.getElementById('testResult').innerHTML = ` + "`" + `
    <div class="test-summary running">
      در حال تست<span class="spinner"></span>
    </div>
    <div id="testSteps"></div>` + "`" + `;

  try {
    const r = await api('POST', '/api/test', {resolver, domain});
    renderTestResult(r);
  } catch(e) {
    document.getElementById('testResult').innerHTML =
      ` + "`" + `<div class="test-summary err">❌ خطا در ارتباط با سرور</div>` + "`" + `;
  }

  btn.disabled = false;
  btn.textContent = 'تست';
}

function renderTestResult(r) {
  const summary = r.ok
    ? ` + "`" + `<div class="test-summary ok">✅ DNS Tunnel قابل دسترس است</div>` + "`" + `
    : ` + "`" + `<div class="test-summary err">❌ مشکل در DNS Tunnel — جزئیات را بررسی کنید</div>` + "`" + `;

  const steps = (r.steps || []).map(s => {
    const icon   = s.ok ? '✅' : '❌';
    const cls    = s.ok ? 'ok' : 'err';
    const ms     = s.ms || 0;
    const msCls  = ms < 200 ? 'fast' : ms < 600 ? 'medium' : 'slow';
    return ` + "`" + `
    <div class="test-step ${cls}">
      <div class="step-icon">${icon}</div>
      <div class="step-body">
        <div class="step-name">
          <span>${esc(s.step)}</span>
          <span class="step-ms ${msCls}">${ms}ms</span>
        </div>
        <div class="step-query">${esc(s.query)}</div>
        <div class="step-detail">${esc(s.detail)}</div>
      </div>
    </div>` + "`" + `;
  }).join('');

  document.getElementById('testResult').innerHTML = summary + steps;
}

init();
</script>
</body>
</html>
`

var (
	db  *store.Store
	mgr *runner.Manager
)

func main() {
	dir := "/opt/slipstream-panel"
	if v := os.Getenv("PANEL_DIR"); v != "" {
		dir = v
	}
	port := "9090"
	if v := os.Getenv("PANEL_PORT"); v != "" {
		port = v
	}
	os.MkdirAll(dir, 0755)
	os.MkdirAll(filepath.Join(dir, "logs"), 0755)

	var err error
	db, err = store.New(filepath.Join(dir, "instances.json"))
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	mgr = runner.New(filepath.Join(dir, "logs"))
	log.Printf("binary: %s", mgr.Bin())

	for _, inst := range db.List() {
		if inst.AutoRestart {
			if err := mgr.Start(inst); err != nil {
				log.Printf("autostart %s: %v", inst.ID, err)
			}
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page))
	})
	mux.HandleFunc("/api/status",     apiStatus)
	mux.HandleFunc("/api/instances",  apiInstances)
	mux.HandleFunc("/api/instances/", apiInstance)
	mux.HandleFunc("/api/test",       apiTestRaw)

	log.Printf("panel on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func jw(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
func je(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

func apiStatus(w http.ResponseWriter, r *http.Request) {
	b := mgr.Bin()
	if b == "" { b = "NOT FOUND" }
	jw(w, map[string]string{"bin": b})
}

func apiInstances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		type row struct {
			store.Instance
			Status string `json:"status"`
		}
		list := db.List()
		out := make([]row, len(list))
		for i, v := range list {
			out[i] = row{v, string(mgr.Status(v.ID))}
		}
		jw(w, out)
	case "POST":
		var body store.Instance
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			je(w, 400, "bad json"); return
		}
		if body.Domain == "" { je(w, 400, "domain required"); return }
		if body.SocksPort == 0 { body.SocksPort = 1080 }
		if body.Name == "" { body.Name = fmt.Sprintf("Instance %d", len(db.List())+1) }
		inst, err := db.Create(body)
		if err != nil { je(w, 400, err.Error()); return }
		jw(w, map[string]string{"ok": "true", "id": inst.ID})
	default:
		w.WriteHeader(405)
	}
}

func apiInstance(w http.ResponseWriter, r *http.Request) {
	tail := strings.TrimPrefix(r.URL.Path, "/api/instances/")
	parts := strings.SplitN(tail, "/", 2)
	id := parts[0]
	action := ""
	if len(parts) == 2 { action = parts[1] }
	if id == "" { je(w, 400, "missing id"); return }

	switch action {
	case "start":
		inst, ok := db.Get(id)
		if !ok { je(w, 404, "not found"); return }
		if err := mgr.Start(inst); err != nil { je(w, 500, err.Error()); return }
		jw(w, map[string]bool{"ok": true})
	case "stop":
		mgr.Stop(id)
		jw(w, map[string]bool{"ok": true})
	case "restart":
		inst, ok := db.Get(id)
		if !ok { je(w, 404, "not found"); return }
		if err := mgr.Restart(inst); err != nil { je(w, 500, err.Error()); return }
		jw(w, map[string]bool{"ok": true})
	case "logs":
		b, _ := os.ReadFile(mgr.LogPath(id))
		lines := strings.Split(string(b), "\n")
		if len(lines) > 200 { lines = lines[len(lines)-200:] }
		jw(w, map[string]string{"logs": strings.Join(lines, "\n")})
	case "clear_logs":
		mgr.ClearLog(id)
		jw(w, map[string]bool{"ok": true})
	case "test":
		inst, ok := db.Get(id)
		if !ok { je(w, 404, "not found"); return }
		jw(w, dnstest.Run(inst.Resolver, inst.Domain))
	case "":
		switch r.Method {
		case "PUT":
			var body store.Instance
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				je(w, 400, "bad json"); return
			}
			was := mgr.Status(id) == runner.Running
			if was { mgr.Stop(id) }
			if err := db.Update(id, body); err != nil { je(w, 500, err.Error()); return }
			if was {
				if inst, ok := db.Get(id); ok { mgr.Start(inst) }
			}
			jw(w, map[string]bool{"ok": true})
		case "DELETE":
			mgr.Stop(id)
			db.Delete(id)
			jw(w, map[string]bool{"ok": true})
		default:
			w.WriteHeader(405)
		}
	default:
		je(w, 404, "unknown: "+action)
	}
}

func apiTestRaw(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" { w.WriteHeader(405); return }
	var body struct {
		Resolver string `json:"resolver"`
		Domain   string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Domain == "" {
		je(w, 400, "resolver and domain required"); return
	}
	jw(w, dnstest.Run(body.Resolver, body.Domain))
}
