import { app, BrowserWindow, ipcMain, dialog } from 'electron';
import path from 'path';
import os from 'os';
import { promises as fs, existsSync, statSync } from 'fs';
import { glob } from 'glob';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const isDev = process.argv.includes('--dev');
const userDataPath = path.join(os.homedir(), '.nanocode');
const settingsFile = path.join(userDataPath, 'settings.json');
const sessionsDir = path.join(userDataPath, 'sessions');
const sessionsDirResolved = path.resolve(sessionsDir);

function ensureString(value, fieldName) {
  if (typeof value !== 'string') {
    throw new Error(`${fieldName} must be a string`);
  }
  return value;
}

function getSafeProjectDir(projectKey) {
  const key = ensureString(projectKey, 'projectKey').trim();
  if (!key) {
    throw new Error('projectKey cannot be empty');
  }

  const projectDir = path.resolve(sessionsDirResolved, key);
  const isInsideSessionsDir =
    projectDir === sessionsDirResolved ||
    projectDir.startsWith(`${sessionsDirResolved}${path.sep}`);

  if (!isInsideSessionsDir) {
    throw new Error('Invalid project key');
  }

  return projectDir;
}

async function ensureDir(dir) {
  try {
    await fs.mkdir(dir, { recursive: true });
  } catch (err) {
    if (err.code !== 'EEXIST') throw err;
  }
}

async function loadSettings() {
  try {
    await ensureDir(userDataPath);
    const raw = await fs.readFile(settingsFile, 'utf-8');
    return JSON.parse(raw);
  } catch {
    return { providers: {}, activeProviderId: null };
  }
}

async function saveSettings(settings) {
  await ensureDir(userDataPath);
  await fs.writeFile(settingsFile, JSON.stringify(settings, null, 2), 'utf-8');
}

async function listSessions(projectKey) {
  const projectDir = getSafeProjectDir(projectKey);
  try {
    const files = await fs.readdir(projectDir);
    const sessions = [];
    for (const file of files) {
      if (!file.endsWith('.json')) continue;
      const id = path.basename(file, '.json');
      const filePath = path.join(projectDir, file);
      const raw = await fs.readFile(filePath, 'utf-8');
      const data = JSON.parse(raw);
      sessions.push({
        id: data.id,
        name: data.name,
        createdAt: data.createdAt,
        projectPath: data.projectPath,
      });
    }
    return sessions.sort((a, b) => b.createdAt - a.createdAt);
  } catch {
    return [];
  }
}

async function loadSession(projectKey, id) {
  const projectDir = getSafeProjectDir(projectKey);
  const safeId = ensureString(id, 'id');
  const filePath = path.join(projectDir, `${safeId}.json`);
  try {
    const raw = await fs.readFile(filePath, 'utf-8');
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

async function saveSession(projectKey, session) {
  const projectDir = getSafeProjectDir(projectKey);
  if (!session || typeof session !== 'object') {
    throw new Error('session must be an object');
  }
  const safeId = ensureString(session.id, 'session.id');
  await ensureDir(projectDir);
  const filePath = path.join(projectDir, `${safeId}.json`);
  await fs.writeFile(filePath, JSON.stringify(session, null, 2), 'utf-8');
}

async function deleteSession(projectKey, id) {
  const projectDir = getSafeProjectDir(projectKey);
  const safeId = ensureString(id, 'id');
  const filePath = path.join(projectDir, `${safeId}.json`);
  try {
    await fs.unlink(filePath);
  } catch (err) {
    if (err.code !== 'ENOENT') throw err;
  }
}

function createWindow() {
  const mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
    frame: false,
    backgroundColor: '#0d0d0d',
    webPreferences: {
      preload: path.join(__dirname, 'preload.cjs'),
      nodeIntegration: false,
      contextIsolation: true,
    },
  });

  if (isDev) {
    mainWindow.loadURL('http://localhost:5173');
    mainWindow.webContents.openDevTools({ mode: 'detach' });
  } else {
    mainWindow.loadFile(path.join(__dirname, '../dist/index.html'));
  }

  ipcMain.on('minimize', () => mainWindow.minimize());
  ipcMain.on('maximize', () => {
    mainWindow.isMaximized() ? mainWindow.unmaximize() : mainWindow.maximize();
  });
  ipcMain.on('close', () => mainWindow.close());
}

function setupIPC() {
  ipcMain.handle('select-folder', async () => {
    const result = await dialog.showOpenDialog({ properties: ['openDirectory'] });
    if (result.canceled || result.filePaths.length === 0) return null;
    return result.filePaths[0];
  });

  ipcMain.handle('resolve-path', async (_event, p) => {
    if (typeof p !== 'string') return null;
    try {
      if (path.isAbsolute(p) && existsSync(p)) return p;
      return null;
    } catch {
      return null;
    }
  });

  ipcMain.handle('load-settings', loadSettings);
  ipcMain.handle('save-settings', (_event, s) => saveSettings(s));
  ipcMain.handle('list-sessions', (_event, pk) => listSessions(pk));
  ipcMain.handle('load-session', (_event, pk, id) => loadSession(pk, id));
  ipcMain.handle('save-session', (_event, pk, s) => saveSession(pk, s));
  ipcMain.handle('delete-session', (_event, pk, id) => deleteSession(pk, id));

  ipcMain.handle('glob', async (_event, pattern, options) => {
    if (typeof pattern !== 'string') {
      return [];
    }
    try {
      const files = await glob(pattern, {
        ...options,
        windowsPathsNoEscape: true,
      });
      return files;
    } catch (err) {
      console.error('Glob error:', err);
      return [];
    }
  });

  ipcMain.handle('stat', async (_event, filePath) => {
    if (typeof filePath !== 'string') {
      throw new Error('filePath must be a string');
    }
    const stats = statSync(filePath);
    return {
      mtimeMs: stats.mtimeMs,
      isDirectory: stats.isDirectory(),
    };
  });

  ipcMain.handle('get-platform', () => process.platform);
}

app.whenReady().then(() => {
  setupIPC();
  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});
