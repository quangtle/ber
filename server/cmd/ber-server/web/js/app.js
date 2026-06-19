let currentServer = 'localhost:8080';
let videos = [];

const player = document.getElementById('videoPlayer');
const libraryList = document.getElementById('libraryList');
const statusEl = document.getElementById('status');
const serverAddr = document.getElementById('serverAddr');
const scanBtn = document.getElementById('scanBtn');
const emptyState = document.getElementById('emptyState');
const videoInfo = document.getElementById('videoInfo');

function apiUrl(path) {
  return `http://${currentServer}/api${path}`;
}

async function apiFetch(path, opts = {}) {
  const url = apiUrl(path);
  const res = await fetch(url, {
    ...opts,
    headers: { 'Content-Type': 'application/json', ...opts.headers },
  });
  const data = await res.json();
  if (!data.ok) throw new Error(data.error || 'request failed');
  return data.data;
}

async function loadLibrary() {
  try {
    videos = await apiFetch('/library');
    renderLibrary();
    statusEl.textContent = 'connected';
    statusEl.className = 'status-indicator connected';
  } catch (e) {
    statusEl.textContent = 'disconnected';
    statusEl.className = 'status-indicator disconnected';
    console.error('Failed to load library:', e);
  }
}

function renderLibrary() {
  libraryList.innerHTML = '';
  if (videos.length === 0) {
    const item = document.createElement('li');
    item.textContent = 'No videos found. Click Scan to add files.';
    item.style.cursor = 'default';
    item.style.color = 'var(--text-muted)';
    libraryList.appendChild(item);
    return;
  }
  videos.forEach(v => {
    const item = document.createElement('li');
    item.textContent = v.title;
    item.dataset.id = v.id;
    item.addEventListener('click', () => playVideo(v));
    libraryList.appendChild(item);
  });
}

function playVideo(video) {
  emptyState.style.display = 'none';
  player.style.display = 'block';

  const streamUrl = `http://${currentServer}/api/stream/${video.id}`;
  player.src = streamUrl;
  player.load();
  player.play();

  const duration = video.duration ? `${Math.round(video.duration)}s` : 'N/A';
  const resolution = video.width && video.height ? `${video.width}x${video.height}` : 'N/A';
  videoInfo.textContent = `${video.title} — ${resolution} — ${duration} — ${video.codec || 'N/A'}`;
}

async function scanLibrary() {
  scanBtn.disabled = true;
  scanBtn.textContent = 'Scanning...';
  try {
    await apiFetch('/library/scan', { method: 'POST' });
    await loadLibrary();
  } catch (e) {
    console.error('Scan failed:', e);
  }
  scanBtn.disabled = false;
  scanBtn.textContent = 'Scan';
}

serverAddr.addEventListener('change', () => {
  currentServer = serverAddr.value.trim() || 'localhost:8080';
  loadLibrary();
});

scanBtn.addEventListener('click', scanLibrary);

player.addEventListener('error', () => {
  videoInfo.textContent = 'Playback error — format may not be supported in this browser';
});

loadLibrary();
setInterval(loadLibrary, 15000);
