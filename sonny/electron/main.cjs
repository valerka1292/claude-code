const { app, BrowserWindow, ipcMain } = require('electron');
const fs = require('fs');
const os = require('os');
const path = require('path');

const isDev = process.env.NODE_ENV === 'development';
const providersPath = path.join(os.homedir(), '.sonny', 'providers.json');
const historyDir = path.join(os.homedir(), '.sonny', 'history');

function ensureProvidersDir() {
  fs.mkdirSync(path.dirname(providersPath), { recursive: true });
}

function readProviders() {
  ensureProvidersDir();
  if (!fs.existsSync(providersPath)) {
    return { activeProviderId: null, providers: {} };
  }

  try {
    const raw = fs.readFileSync(providersPath, 'utf-8');
    const parsed = JSON.parse(raw);
    const activeProviderId = typeof parsed?.activeProviderId === 'string' ? parsed.activeProviderId : null;
    const providers = typeof parsed?.providers === 'object' && parsed.providers !== null ? parsed.providers : {};
    return { activeProviderId, providers };
  } catch {
    return { activeProviderId: null, providers: {} };
  }
}

function writeProviders(data) {
  ensureProvidersDir();
  const safeData = {
    activeProviderId: typeof data?.activeProviderId === 'string' ? data.activeProviderId : null,
    providers: typeof data?.providers === 'object' && data.providers !== null ? data.providers : {},
  };
  try {
    fs.writeFileSync(providersPath, JSON.stringify(safeData, null, 2), 'utf-8');
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to write providers file: ${message}`);
  }
}

function registerProvidersIpc() {
  ipcMain.handle('providers:getAll', () => readProviders());
  ipcMain.handle('providers:save', (_, data) => {
    writeProviders(data);
    return readProviders();
  });
}

function ensureHistoryDir() {
  fs.mkdirSync(historyDir, { recursive: true });
}

function readChat(chatId) {
  ensureHistoryDir();
  const filePath = path.join(historyDir, `${chatId}.json`);
  if (!fs.existsSync(filePath)) {
    return null;
  }

  try {
    const raw = fs.readFileSync(filePath, 'utf-8');
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function writeChat(chatId, data) {
  ensureHistoryDir();
  const filePath = path.join(historyDir, `${chatId}.json`);
  try {
    fs.writeFileSync(filePath, JSON.stringify(data, null, 2), 'utf-8');
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to write chat "${chatId}": ${message}`);
  }
}

function listChats() {
  ensureHistoryDir();
  const files = fs.readdirSync(historyDir);

  return files
    .filter((fileName) => fileName.endsWith('.json'))
    .map((fileName) => {
      const id = path.basename(fileName, '.json');
      const raw = readChat(id);
      if (!raw || typeof raw !== 'object') {
        return null;
      }

      return {
        id: typeof raw.id === 'string' ? raw.id : id,
        title: typeof raw.title === 'string' ? raw.title : 'Untitled Chat',
        updatedAt: typeof raw.updatedAt === 'number' ? raw.updatedAt : 0,
      };
    })
    .filter(Boolean)
    .sort((a, b) => b.updatedAt - a.updatedAt);
}

function registerHistoryIpc() {
  ipcMain.handle('history:list', () => listChats());
  ipcMain.handle('history:get', (_, chatId) => readChat(chatId));
  ipcMain.handle('history:save', (_, chatId, data) => {
    writeChat(chatId, data);
    return listChats();
  });
  ipcMain.handle('history:delete', (_, chatId) => {
    ensureHistoryDir();
    const filePath = path.join(historyDir, `${chatId}.json`);
    if (fs.existsSync(filePath)) {
      fs.unlinkSync(filePath);
    }
    return listChats();
  });
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1400,
    height: 900,
    minWidth: 800,
    minHeight: 600,
    frame: false,
    backgroundColor: '#0f0f0f',
    webPreferences: {
      preload: path.join(__dirname, 'preload.cjs'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  if (isDev) {
    win.loadURL('http://localhost:3000');
    win.webContents.openDevTools();
  } else {
    win.loadFile(path.join(__dirname, '../dist/index.html'));
  }

  ipcMain.on('minimize-window', () => win.minimize());
  ipcMain.on('maximize-window', () => {
    if (win.isMaximized()) {
      win.unmaximize();
    } else {
      win.maximize();
    }
  });
  ipcMain.on('close-window', () => win.close());
}

app.whenReady().then(() => {
  registerProvidersIpc();
  registerHistoryIpc();
  createWindow();
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});
