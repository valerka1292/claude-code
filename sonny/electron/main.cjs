const { app, BrowserWindow, ipcMain } = require('electron');
const fsPromises = require('fs/promises');
const os = require('os');
const path = require('path');

const isDev = process.env.NODE_ENV === 'development';
const providersPath = path.join(os.homedir(), '.sonny', 'providers.json');
const historyDir = path.join(os.homedir(), '.sonny', 'history');
const historyIndexPath = path.join(historyDir, 'index.json');

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

const SAFE_ID_PATTERN = /^[a-zA-Z0-9_-]+$/;

function validateChatId(chatId) {
  if (typeof chatId !== 'string' || !SAFE_ID_PATTERN.test(chatId)) {
    throw new Error(`Invalid chatId: ${String(chatId)}`);
  }
  return chatId;
}

function registerProvidersIpc() {
  ipcMain.removeHandler('providers:getAll');
  ipcMain.removeHandler('providers:save');

  ipcMain.handle('providers:getAll', async () => {
    console.log('[IPC] providers:getAll');
    return readProviders();
  });
  ipcMain.handle('providers:save', async (_, data) => {
    console.log('[IPC] providers:save');
    await writeProviders(data);
    return readProviders();
  });
}

async function ensureHistoryDir() {
  await fsPromises.mkdir(historyDir, { recursive: true });
}

async function readHistoryIndex() {
  await ensureHistoryDir();
  if (!(await pathExists(historyIndexPath))) {
    return null;
  }

  try {
    const raw = await fsPromises.readFile(historyIndexPath, 'utf-8');
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return null;
    }

    return parsed
      .filter((item) => item && typeof item === 'object')
      .map((item) => ({
        id: typeof item.id === 'string' ? item.id : '',
        title: typeof item.title === 'string' ? item.title : 'Untitled Chat',
        updatedAt: typeof item.updatedAt === 'number' ? item.updatedAt : 0,
      }))
      .filter((item) => item.id.length > 0);
  } catch (error) {
    console.error('[Main] Failed to read history index:', error);
    return null;
  }
}

async function writeHistoryIndex(chats) {
  await ensureHistoryDir();
  console.log(`[Main] Writing history index, count: ${chats.length}`);
  await fsPromises.writeFile(historyIndexPath, JSON.stringify(chats, null, 2), 'utf-8');
}

async function readChat(chatId) {
  validateChatId(chatId);
  await ensureHistoryDir();
  const filePath = path.join(historyDir, `${chatId}.json`);
  if (!(await pathExists(filePath))) {
    console.warn(`[Main] Chat file not found: ${filePath}`);
    return null;
  }

  try {
    const raw = await fsPromises.readFile(filePath, 'utf-8');
    return JSON.parse(raw);
  } catch (error) {
    console.error(`[Main] Failed to read chat ${chatId}:`, error);
    return null;
  }
}

async function writeChat(chatId, data) {
  validateChatId(chatId);
  await ensureHistoryDir();
  const filePath = path.join(historyDir, `${chatId}.json`);
  console.log(`[Main] Writing chat file: ${filePath}, messages: ${data?.messages?.length ?? 0}`);
  try {
    await fsPromises.writeFile(filePath, JSON.stringify(data, null, 2), 'utf-8');
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to write chat "${chatId}": ${message}`);
  }
}

async function listChats() {
  await ensureHistoryDir();
  const index = await readHistoryIndex();
  if (index) {
    console.log(`[Main] Returning history from index, count: ${index.length}`);
    return index.sort((a, b) => b.updatedAt - a.updatedAt);
  }

  console.log('[Main] Index not found, scanning directory...');
  const files = await fsPromises.readdir(historyDir);
  const chats = await Promise.all(files
    .filter((fileName) => fileName.endsWith('.json'))
    .filter((fileName) => fileName !== 'index.json')
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

  const normalized = chats.filter(Boolean).sort((a, b) => b.updatedAt - a.updatedAt);
  await writeHistoryIndex(normalized);
  return normalized;
}

function registerHistoryIpc() {
  ipcMain.removeHandler('history:list');
  ipcMain.removeHandler('history:get');
  ipcMain.removeHandler('history:save');
  ipcMain.removeHandler('history:delete');

  ipcMain.handle('history:list', async () => {
    console.log('[IPC] history:list');
    return listChats();
  });
  ipcMain.handle('history:get', async (_, chatId) => {
    console.log(`[IPC] history:get ${chatId}`);
    return readChat(chatId);
  });
  ipcMain.handle('history:save', async (_, chatId, data) => {
    console.log(`[IPC] history:save ${chatId}`);
    await writeChat(chatId, data);
    const current = (await readHistoryIndex()) ?? [];
    const nextItem = {
      id: typeof data?.id === 'string' ? data.id : chatId,
      title: typeof data?.title === 'string' ? data.title : 'Untitled Chat',
      updatedAt: typeof data?.updatedAt === 'number' ? data.updatedAt : Date.now(),
    };
    const withoutCurrent = current.filter((chat) => chat.id !== chatId);
    const nextList = [...withoutCurrent, nextItem].sort((a, b) => b.updatedAt - a.updatedAt);
    await writeHistoryIndex(nextList);
    return nextList;
  });
  ipcMain.handle('history:delete', async (_, chatId) => {
    console.log(`[IPC] history:delete ${chatId}`);
    await ensureHistoryDir();
    const filePath = path.join(historyDir, `${chatId}.json`);
    if (await pathExists(filePath)) {
      await fsPromises.unlink(filePath);
    }
    const current = (await readHistoryIndex()) ?? [];
    const nextList = current.filter((chat) => chat.id !== chatId);
    await writeHistoryIndex(nextList);
    return nextList;
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
    // win.webContents.openDevTools();
  } else {
    win.loadFile(path.join(__dirname, '../dist/index.html'));
  }

  ipcMain.removeAllListeners('minimize-window');
  ipcMain.removeAllListeners('maximize-window');
  ipcMain.removeAllListeners('close-window');

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
