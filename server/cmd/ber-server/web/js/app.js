let currentServer = window.location.host;
let videos = [];

const player = document.getElementById('videoPlayer');
const libraryList = document.getElementById('libraryList');
const statusEl = document.getElementById('status');
const scanBtn = document.getElementById('scanBtn');
const importBtn = document.getElementById('importBtn');
const importModal = document.getElementById('importModal');
const importPath = document.getElementById('importPath');
const importCancelBtn = document.getElementById('importCancelBtn');
const importConfirmBtn = document.getElementById('importConfirmBtn');
const importError = document.getElementById('importError');

const browseBtn = document.getElementById('browseBtn');
const browseModal = document.getElementById('browseModal');
const browserPath = document.getElementById('browserPath');
const browserList = document.getElementById('browserList');
const browserUpBtn = document.getElementById('browserUpBtn');
const browserCancelBtn = document.getElementById('browserCancelBtn');
const browserSelectBtn = document.getElementById('browserSelectBtn');

const emptyState = document.getElementById('emptyState');
const videoInfo = document.getElementById('videoInfo');

let currentBrowsePath = '';

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
    item.textContent = 'No videos found. Click Import or Scan to add files.';
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

// --- Folder browser ---
async function loadBrowser(dir) {
  currentBrowsePath = dir;
  browserPath.textContent = dir;
  browserList.innerHTML = '<li class="browser-loading">Loading...</li>';
  try {
    const data = await apiFetch(`/browse?path=${encodeURIComponent(dir)}`);
    browserList.innerHTML = '';
    data.entries.forEach(e => {
      if (!e.is_dir) return;
      const item = document.createElement('li');
      item.textContent = e.name;
      item.className = 'browser-dir';
      item.addEventListener('click', () => loadBrowser(e.path));
      browserList.appendChild(item);
    });
    if (data.entries.filter(e => e.is_dir).length === 0) {
      browserList.innerHTML = '<li class="browser-empty">(no subfolders)</li>';
    }
    browserUpBtn.style.display = data.parent ? 'inline' : 'none';
  } catch (e) {
    browserList.innerHTML = `<li class="browser-error">Error: ${e.message}</li>`;
  }
}

// --- Event handlers ---
scanBtn.addEventListener('click', scanLibrary);

importBtn.addEventListener('click', () => {
  importPath.value = '';
  importError.textContent = '';
  importModal.classList.remove('hidden');
  importPath.focus();
});

importCancelBtn.addEventListener('click', () => {
  importModal.classList.add('hidden');
});

importModal.addEventListener('click', (e) => {
  if (e.target === importModal) importModal.classList.add('hidden');
});

importPath.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') importConfirmBtn.click();
  if (e.key === 'Escape') importModal.classList.add('hidden');
});

browseBtn.addEventListener('click', async () => {
  browseModal.classList.remove('hidden');
  currentBrowsePath = '';
  loadBrowser('');
});

browserCancelBtn.addEventListener('click', () => {
  browseModal.classList.add('hidden');
});

browserSelectBtn.addEventListener('click', () => {
  importPath.value = currentBrowsePath;
  browseModal.classList.add('hidden');
});

browserUpBtn.addEventListener('click', () => {
  // go to parent: strip trailing slash, take dirname
  let parent = currentBrowsePath.replace(/\\/g, '/').replace(/\/$/, '');
  // if at a drive root (e.g. "C:" or "C"), show drive list
  if (/^[A-Za-z]$/.test(parent) || /^[A-Za-z]:$/.test(parent)) {
    loadBrowser('');
    return;
  }
  parent = parent.substring(0, parent.lastIndexOf('/'));
  if (!parent) parent = '';
  loadBrowser(parent);
});

browseModal.addEventListener('click', (e) => {
  if (e.target === browseModal) browseModal.classList.add('hidden');
});

importConfirmBtn.addEventListener('click', async () => {
  const path = importPath.value.trim();
  if (!path) { importError.textContent = 'Enter a file or folder path'; return; }
  importConfirmBtn.disabled = true;
  importConfirmBtn.textContent = 'Importing...';
  try {
    // if path has a video extension, add single file; otherwise scan as folder
    const videoExts = ['.mp4','.mkv','.avi','.mov','.wmv','.flv','.webm','.mpeg','.mpg','.ts','.mts','.ogv'];
    const ext = path.substring(path.lastIndexOf('.')).toLowerCase();
    if (videoExts.includes(ext)) {
      await apiFetch('/library/add', { method: 'POST', body: JSON.stringify({ path }) });
    } else {
      await apiFetch('/library/scan-dir', { method: 'POST', body: JSON.stringify({ path }) });
    }
    importModal.classList.add('hidden');
    await loadLibrary();
  } catch (e) {
    importError.textContent = e.message || 'Import failed';
  }
  importConfirmBtn.disabled = false;
  importConfirmBtn.textContent = 'Import';
});

player.addEventListener('error', () => {
  videoInfo.textContent = 'Playback error — format may not be supported in this browser';
});

loadLibrary();
setInterval(loadLibrary, 15000);
