package main

// dashboardHTML is the entire single-file frontend. It fetches /api/status,
// renders a live status page styled with the dui.css system via shell.js.
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Status — dhanur.me</title>
  <meta name="description" content="Live status dashboard for all dhanur.me services." />

  <script>
  window.SiteNavConfig = {
    mode: "navbar",
    nav: [{ name: "Home", url: "https://dhanur.me/", icon: "fa-solid fa-house" }],
    sidebarNav: [],
    activePath: window.location.pathname,
    badge: { text: "STATUS", class: "badge-neutral" },
    showSearch: false,
    showAppsGrid: true,
    showThemeToggle: true
  };
  </script>
  <script src="https://dhanur.me/js/shell.min.js"></script>

  <style>
    /* ── Pulse animation for live indicator ── */
    @keyframes pulse-ring {
      0%   { box-shadow: 0 0 0 0 color-mix(in oklab, var(--color-success) 55%, transparent); }
      70%  { box-shadow: 0 0 0 7px transparent; }
      100% { box-shadow: 0 0 0 0 transparent; }
    }
    .pulse { animation: pulse-ring 2s cubic-bezier(0.455,0.03,0.515,0.955) infinite; }

    /* ── Spark bar (90 ticks) ── */
    .spark-bar { display: flex; gap: 2px; align-items: flex-end; height: 28px; }
    .spark-tick {
      flex: 1 1 0;
      min-width: 2px;
      border-radius: 2px;
      transition: height 0.3s ease;
    }
    .tick-up      { background: var(--color-success); opacity: 0.8; }
    .tick-degraded{ background: var(--color-warning); opacity: 0.8; }
    .tick-down    { background: var(--color-error);   opacity: 0.8; }
    .tick-unknown { background: color-mix(in oklab, var(--color-base-content) 20%, transparent); }

    /* ── Status pill ── */
    .pill {
      display: inline-flex; align-items: center; gap: 6px;
      font-size: 11px; font-weight: 700; letter-spacing: 0.05em;
      text-transform: uppercase; padding: 3px 10px;
      border-radius: 999px;
    }
    .pill-up       { background: color-mix(in oklab, var(--color-success) 14%, transparent); color: var(--color-success); border: 1px solid color-mix(in oklab, var(--color-success) 30%, transparent); }
    .pill-degraded { background: color-mix(in oklab, var(--color-warning) 14%, transparent); color: var(--color-warning); border: 1px solid color-mix(in oklab, var(--color-warning) 30%, transparent); }
    .pill-down     { background: color-mix(in oklab, var(--color-error) 14%, transparent);   color: var(--color-error);   border: 1px solid color-mix(in oklab, var(--color-error) 30%, transparent); }
    .pill-unknown  { background: color-mix(in oklab, var(--color-base-content) 8%, transparent); color: var(--color-base-content); border: 1px solid color-mix(in oklab, var(--color-base-content) 15%, transparent); }

    /* ── Latency badge ── */
    .latency-chip {
      font-size: 10px; font-family: ui-monospace, monospace;
      color: color-mix(in oklab, var(--color-base-content) 45%, transparent);
      padding: 2px 6px;
      background: color-mix(in oklab, var(--color-base-content) 6%, transparent);
      border-radius: 4px;
    }
  </style>
</head>
<body>
<div class="navbar site-topbar fixed top-0 left-0 right-0 z-50 h-16">
    <!-- Mobile hamburger -->
    <div class="flex-none lg:hidden">
        <label id="hamburger-toggle" for="my-drawer-2" aria-label="Open menu" class="btn btn-ghost btn-circle btn-sm transition-colors duration-200">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path></svg>
        </label>
    </div>

    <!-- Logo -->
    <div class="flex-1 flex justify-center items-center lg:flex lg:justify-start">
        <a href="https://dhanur.me/" class="btn btn-ghost hover:bg-transparent hover:border-transparent normal-case text-xl font-bold text-base-content site-logo-link">
            dhanur.me
        </a>
    </div>

    <div class="flex-none lg:hidden" aria-hidden="true"><div class="btn btn-circle btn-sm invisible"></div></div>

    <!-- Desktop Menu -->
    <div class="flex-none hidden lg:flex">
        <ul class="menu menu-horizontal px-1 items-center">
            
            <!-- Apps Grid -->
            <li data-nav-chrome="apps" class="ml-1">
                <div class="relative p-0" data-dropdown="apps">
                    <div tabindex="0" role="button" class="btn btn-ghost btn-square tooltip tooltip-bottom" data-tooltip-label="Apps" data-tooltip-position="bottom" aria-label="Apps">
                        <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"></path></svg>
                    </div>
                    <div class="dropdown-panel z-50 mt-2 p-4 bg-base-100 border border-base-content/10 rounded-box w-64 right-0 mr-2 md:mr-4">
                        <div class="grid grid-cols-3 gap-2" data-apps-grid>
                            <!-- Apps grid populated by manifest.js -->
                            <div class="col-span-3 flex justify-center py-4 opacity-50"><span class="loading loading-spinner loading-sm"></span></div>
                        </div>
                    </div>
                </div>
            </li>

            <!-- Account -->
            <li data-nav-chrome="account" class="ml-1">
                <div class="relative p-0" data-dropdown="account">
                    <div tabindex="0" role="button" class="btn btn-ghost btn-circle tooltip tooltip-bottom" data-tooltip-label="Account" data-tooltip-position="bottom" aria-label="Account">
                        <div data-auth="nav-guest-avatar" class="w-9 h-9 rounded-full bg-base-300 flex items-center justify-center">
                            <i class="fa-solid fa-user text-base-content/50 text-sm"></i>
                        </div>
                        <div data-auth="nav-authed-avatar" class="hidden w-9 h-9 rounded-full ring-2 ring-primary ring-offset-base-100 ring-offset-1 overflow-hidden">
                            <img src="" alt="Profile" class="w-full h-full object-cover" />
                        </div>
                    </div>
                    <div class="dropdown-panel z-50 mt-2 bg-base-100 border border-base-content/10 rounded-box w-64 right-0 mr-2 md:mr-4 overflow-hidden">
                        
                        <!-- authed header -->
                        <div data-auth="nav-authed-header" class="hidden px-3 pt-3 pb-2 border-b border-base-content/10 cursor-default select-none">
                            <div class="flex items-center gap-2.5">
                                <div class="w-8 h-8 rounded-full overflow-hidden shrink-0"><img data-auth="nav-authed-header-avatar" class="w-full h-full object-cover" src="" alt="Profile" /></div>
                                <div class="flex-1 min-w-0">
                                    <div class="font-semibold truncate text-sm" data-auth="nav-name"></div>
                                    <div class="text-xs opacity-60 truncate" data-auth="nav-email"></div>
                                </div>
                                <span class="badge badge-sm hidden" data-auth="nav-role"></span>
                            </div>
                        </div>

                        <!-- guest header -->
                        <div data-auth="nav-guest-header" class="px-3 pt-3 pb-2 border-b border-base-content/10 cursor-default select-none">
                            <div class="flex items-center gap-2.5">
                                <div class="w-8 h-8 rounded-full bg-base-300 flex items-center justify-center shrink-0">
                                    <i class="fa-solid fa-user text-base-content/40"></i>
                                </div>
                                <div class="flex-1 min-w-0">
                                    <div class="font-semibold text-sm">Guest</div>
                                    <div class="text-xs opacity-60">Not signed in</div>
                                </div>
                            </div>
                        </div>

                        <!-- Credits display -->
                        <div class="hidden px-3 py-1.5 border-b border-base-content/10" data-auth="credits-row">
                            <div class="flex items-center justify-between text-xs">
                                <span class="flex items-center gap-1.5 opacity-70">
                                    <span>🪙</span>
                                    <span data-auth="credits-balance">—</span>
                                    <span class="opacity-60">credits</span>
                                </span>
                                <span class="opacity-40 text-[10px]" data-auth="credits-reset"></span>
                            </div>
                        </div>

                        <ul class="menu p-2">
                            <li data-nav-chrome="theme">
                                <div class="flex items-center gap-2.5 px-2.5 py-2">
                                    <i class="fa-solid fa-circle-half-stroke w-4 text-center opacity-60"></i>
                                    <div id="theme-toggle" class="theme-switcher w-full bg-base-300/70 text-xs font-medium">
                                        <button data-theme-mode="light" class="theme-switcher-btn rounded-md px-2 py-1.5 cursor-pointer"><i class="fa-solid fa-sun mr-1"></i>Light</button>
                                        <button data-theme-mode="dark" class="theme-switcher-btn rounded-md px-2 py-1.5 cursor-pointer"><i class="fa-solid fa-moon mr-1"></i>Dark</button>
                                        <button data-theme-mode="auto" class="theme-switcher-btn rounded-md px-2 py-1.5 cursor-pointer"><i class="fa-solid fa-circle-half-stroke mr-1"></i>Auto</button>
                                    </div>
                                </div>
                            </li>
                            <li data-auth="login-item">
                                <a class="flex items-center gap-3" href="https://auth.dhanur.me" data-auth="login-btn"><i class="fa-solid fa-right-to-bracket w-4 text-center"></i><span>Sign In</span></a>
                            </li>
                            <li class="hidden" data-auth="account-item">
                                <a class="flex items-center gap-3" href="https://auth.dhanur.me"><i class="fa-solid fa-gear w-4 text-center"></i><span>Account Settings</span></a>
                            </li>
                            <li class="hidden border-t border-base-content/10 mt-1 pt-1" data-auth="logout-item">
                                <button type="button" class="w-full text-left flex items-center gap-3 text-error/80 hover:text-error" data-auth="logout-btn"><i class="fa-solid fa-right-from-bracket w-4 text-center"></i><span>Sign Out</span></button>
                            </li>
                        </ul>
                    </div>
                </div>
            </li>
        </ul>
    </div>
</div>


<main class="w-full max-w-4xl mx-auto p-4 lg:p-8">

  <!-- Page header -->
  <div class="mt-14 mb-10">
    <div class="flex items-center gap-3 mb-3">
      <span id="global-dot" class="w-3 h-3 rounded-full bg-base-content/20"></span>
      <h1 class="text-4xl lg:text-5xl font-extrabold tracking-tight" id="global-title">Checking…</h1>
    </div>
    <p class="text-base-content/60 text-sm" id="global-sub">Fetching service health…</p>
  </div>

  <!-- Service cards -->
  <div id="cards" class="flex flex-col gap-4"></div>

  <!-- Footer -->
  <div class="mt-10 pt-6 border-t border-base-content/10 flex items-center justify-between text-[11px] text-base-content/35 font-mono">
    <span>up.dhanur.me · refreshing every 30s</span>
    <span id="last-checked">—</span>
  </div>
</main>

<script>
const STATUS_COLORS = {
  up:       'pill-up',
  degraded: 'pill-degraded',
  down:     'pill-down',
  unknown:  'pill-unknown',
};
const STATUS_LABELS = {
  up:       'Operational',
  degraded: 'Degraded',
  down:     'Down',
  unknown:  'Unknown',
};
const GLOBAL_TEXT = {
  up:       ['All Systems Operational', 'Everything looks good.'],
  degraded: ['Partial Outage', 'Some services are experiencing issues.'],
  down:     ['Major Outage', 'One or more services are unreachable.'],
  unknown:  ['Checking…', 'Fetching service health…'],
};
const GLOBAL_DOT = {
  up:       'bg-success pulse',
  degraded: 'bg-warning pulse',
  down:     'bg-error pulse',
  unknown:  'bg-base-content/20',
};

function latencyText(ms) {
  if (ms < 0) return 'n/a';
  if (ms < 1000) return ms + ' ms';
  return (ms / 1000).toFixed(2) + ' s';
}

function tickClass(status) {
  switch (status) {
    case 'up':       return 'tick-up';
    case 'degraded': return 'tick-degraded';
    case 'down':     return 'tick-down';
    default:         return 'tick-unknown';
  }
}

function tickHeight(status) {
  switch (status) {
    case 'up':       return '100%';
    case 'degraded': return '55%';
    case 'down':     return '20%';
    default:         return '35%';
  }
}

function renderSpark(history) {
  // history is newest-first; we want to display oldest-left
  const slots = history.slice().reverse();
  // Pad to historyLen if shorter
  while (slots.length < 90) slots.unshift(null);
  return slots.map(h => {
    const cls  = h ? tickClass(h.status) : 'tick-unknown';
    const hpct = h ? tickHeight(h.status) : '35%';
    return '<div class="spark-tick ' + cls + '" style="height:' + hpct + '"></div>';
  }).join('');
}

function renderCard(svc) {
  const cur  = svc.current;
  const pill = STATUS_COLORS[cur.status] || 'pill-unknown';
  const lbl  = STATUS_LABELS[cur.status] || 'Unknown';

  return '<div class="card bg-base-100 border border-base-content/10 rounded-2xl shadow-sm">' +
    '<div class="card-body py-4 px-5 gap-3">' +
      // Row 1: icon + name + status + latency
      '<div class="flex items-center justify-between gap-3">' +
        '<div class="flex items-center gap-3">' +
          '<div class="flex items-center justify-center w-9 h-9 rounded-xl bg-base-200 border border-base-content/10 flex-shrink-0">' +
            '<i class="' + svc.config.icon + ' text-base-content/70"></i>' +
          '</div>' +
          '<div>' +
            '<a href="' + svc.config.homepageUrl + '" target="_blank" rel="noopener"' +
              ' class="font-bold text-[15px] text-base-content hover:underline decoration-1 underline-offset-2">' +
              svc.config.displayName +
            '</a>' +
            '<div class="text-[10px] text-base-content/40 font-mono mt-0.5">' + svc.config.category + '</div>' +
          '</div>' +
        '</div>' +
        '<div class="flex items-center gap-2">' +
          '<span class="latency-chip">' + latencyText(cur.latencyMs) + '</span>' +
          '<span class="pill ' + pill + '">' + lbl + '</span>' +
        '</div>' +
      '</div>' +
      // Row 2: spark + uptime
      '<div>' +
        '<div class="spark-bar">' + renderSpark(svc.history) + '</div>' +
        '<div class="flex justify-between text-[10px] text-base-content/35 mt-1 font-mono">' +
          '<span>45 min ago</span>' +
          '<span>' + svc.uptimePercent.toFixed(1) + '% uptime</span>' +
        '</div>' +
      '</div>' +
    '</div>' +
  '</div>';
}

async function refresh() {
  try {
    const res  = await fetch('/api/status');
    if (!res.ok) throw new Error('HTTP ' + res.status);
    const data = await res.json();

    // Update global header
    const [title, sub] = GLOBAL_TEXT[data.overall] || GLOBAL_TEXT.unknown;
    document.getElementById('global-title').textContent = title;
    document.getElementById('global-sub').textContent   = sub;

    const dot = document.getElementById('global-dot');
    dot.className = 'w-3 h-3 rounded-full ' + (GLOBAL_DOT[data.overall] || GLOBAL_DOT.unknown);

    // Render cards
    document.getElementById('cards').innerHTML =
      data.services.map(renderCard).join('');

    // Footer timestamp
    document.getElementById('last-checked').textContent =
      'Last checked: ' + new Date(data.updatedAt).toLocaleTimeString();

  } catch (e) {
    console.error('Failed to fetch status:', e);
  }
}

// Initial load + polling every 30 s (matching server-side interval)
refresh();
setInterval(refresh, 30_000);
</script>
</body>
</html>`
