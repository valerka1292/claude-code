const { app, BrowserWindow, ipcMain } = require('electron');
const fsPromises = require('fs/promises');
const os = require('os');
const path = require('path');

const isDev = process.env.NODE_ENV === 'development';
const providersPath = path.join(os.homedir(), '.sonny', 'providers.json');
const historyDir = path.join(os.homedir(), '.sonny', 'history');

async function pathExists(targetPath) {
  try {
    await fsPromises.access(targetPath);
    return true;
  } catch {
    return false;
  }
}

async function ensureProvidersDir() {
  await fsPromises.mkdir(path.dirname(providersPath), { recursive: true });
}

async function readProviders() {
  await ensureProvidersDir();
  if (!(await pathExists(providersPath))) {
    return { activeProviderId: null, providers: {} };
  }

  try {
    const raw = await fsPromises.readFile(providersPath, 'utf-8');
    const parsed = JSON.parse(raw);
    const activeProviderId = typeof parsed?.activeProviderId === 'string' ? parsed.activeProviderId : null;
    const providers = typeof parsed?.providers === 'object' && parsed.providers !== null ? parsed.providers : {};
    return { activeProviderId, providers };
  } catch {
    return { activeProviderId: null, providers: {} };
  }
}

async function writeProviders(data) {
  await ensureProvidersDir();
  const safeData = {
    activeProviderId: typeof data?.activeProviderId === 'string' ? data.activeProviderId : null,
    providers: typeof data?.providers === 'object' && data.providers !== null ? data.providers : {},
  };
  try {
    await fsPromises.writeFile(providersPath, JSON.stringify(safeData, null, 2), 'utf-8');
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to write providers file: ${message}`);
  }
}

function registerProvidersIpc() {
  ipcMain.handle('providers:getAll', async () => readProviders());
  ipcMain.handle('providers:save', async (_, data) => {
    await writeProviders(data);
    return readProviders();
  });
}

async function ensureHistoryDir() {
  await fsPromises.mkdir(historyDir, { recursive: true });
}

async function readChat(chatId) {
  await ensureHistoryDir();
  const filePath = path.join(historyDir, `${chatId}.json`);
  if (!(await pathExists(filePath))) {
    return null;
  }

  try {
    const raw = await fsPromises.readFile(filePath, 'utf-8');
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

async function writeChat(chatId, data) {
  await ensureHistoryDir();
  const filePath = path.join(historyDir, `${chatId}.json`);
  try {
    await fsPromises.writeFile(filePath, JSON.stringify(data, null, 2), 'utf-8');
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to write chat "${chatId}": ${message}`);
  }
}

async function listChats() {
  await ensureHistoryDir();
  const files = await fsPromises.readdir(historyDir);
  const chats = await Promise.all(files
    .filter((fileName) => fileName.endsWith('.json'))
    .map(async (fileName) => {
      const id = path.basename(fileName, '.json');
      const raw = await readChat(id);
      if (!raw || typeof raw !== 'object') {
        return null;
      }

      return {
        id: typeof raw.id === 'string' ? raw.id : id,
        title: typeof raw.title === 'string' ? raw.title : 'Untitled Chat',
        updatedAt: typeof raw.updatedAt === 'number' ? raw.updatedAt : 0,
      };
    }));

  return chats.filter(Boolean).sort((a, b) => b.updatedAt - a.updatedAt);
}

function registerHistoryIpc() {
  ipcMain.handle('history:list', async () => listChats());
  ipcMain.handle('history:get', async (_, chatId) => readChat(chatId));
  ipcMain.handle('history:save', async (_, chatId, data) => {
    await writeChat(chatId, data);
    return listChats();
  });
  ipcMain.handle('history:delete', async (_, chatId) => {
    await ensureHistoryDir();
    const filePath = path.join(historyDir, `${chatId}.json`);
    if (await pathExists(filePath)) {
      await fsPromises.unlink(filePath);
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
