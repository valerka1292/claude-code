const { app, BrowserWindow, ipcMain } = require('electron');
const fsPromises = require('fs/promises');
const os = require('os');
const path = require('path');

const isDev = process.env.NODE_ENV === 'development';
const providersPath = path.join(os.homedir(), '.sonny', 'providers.json');
const historyDir = path.join(os.homedir(), '.sonny', 'history');
const historyIndexPath = path.join(historyDir, 'index.json');

const SAFE_ID_PATTERN = /^[a-zA-Z0-9_-]+$/;

function validateChatId(chatId) {
  if (typeof chatId !== 'string' || !SAFE_ID_PATTERN.test(chatId)) {
    console.error(`[Main] CRITICAL: Invalid chatId attempted: ${String(chatId)}`);
    throw new Error(`Invalid chatId: ${String(chatId)}`);
  }
  return chatId;
}

async function pathExists(targetPath) {
  try {
    await fsPromises.access(targetPath);
    return true;
  } catch {
    return false;
  }
}

async function ensureDir(dirPath) {
  await fsPromises.mkdir(dirPath, { recursive: true });
}

// Atomic write: write to .tmp then rename
async function atomicWriteFile(targetPath, content) {
  const tmpPath = `${targetPath}.tmp`;
  try {
    await ensureDir(path.dirname(targetPath));
    await fsPromises.writeFile(tmpPath, content, 'utf-8');
    await fsPromises.rename(tmpPath, targetPath);
  } catch (error) {
    console.error(`[Main] Atomic write failed for ${targetPath}:`, error);
    await fsPromises.unlink(tmpPath).catch(() => {});
    throw error;
  }
}

async function readProviders() {
  console.log('[Main] Reading providers...');
  if (!(await pathExists(providersPath))) {
    return { activeProviderId: null, providers: {} };
  }

  try {
    const raw = await fsPromises.readFile(providersPath, 'utf-8');
    const parsed = JSON.parse(raw);
    const activeProviderId = typeof parsed?.activeProviderId === 'string' ? parsed.activeProviderId : null;
    const providers = typeof parsed?.providers === 'object' && parsed.providers !== null ? parsed.providers : {};
    return { activeProviderId, providers };
  } catch (error) {
    console.error('[Main] Failed to read providers:', error);
    return { activeProviderId: null, providers: {} };
  }
}

async function writeProviders(data) {
  console.log('[Main] Writing providers (atomic)...');
  const safeData = {
    activeProviderId: typeof data?.activeProviderId === 'string' ? data.activeProviderId : null,
    providers: typeof data?.providers === 'object' && data.providers !== null ? data.providers : {},
  };
  await atomicWriteFile(providersPath, JSON.stringify(safeData, null, 2));
}

async function readHistoryIndex() {
  if (!(await pathExists(historyIndexPath))) return null;
  try {
    const raw = await fsPromises.readFile(historyIndexPath, 'utf-8');
    return JSON.parse(raw);
  } catch (error) {
    console.error('[Main] Failed to read history index:', error);
    return null;
  }
}

async function writeHistoryIndex(chats) {
  console.log(`[Main] Updating history index (count: ${chats.length})`);
  await atomicWriteFile(historyIndexPath, JSON.stringify(chats, null, 2));
}

async function readChat(chatId) {
  validateChatId(chatId);
  const filePath = path.join(historyDir, `${chatId}.json`);
  if (!(await pathExists(filePath))) return null;

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
  const filePath = path.join(historyDir, `${chatId}.json`);
  console.log(`[Main] Writing chat ${chatId} (messages: ${data?.messages?.length ?? 0})`);
  await atomicWriteFile(filePath, JSON.stringify(data, null, 2));
}

async function listChats() {
  await ensureDir(historyDir);
  const index = await readHistoryIndex();
  if (index) return index.sort((a, b) => b.updatedAt - a.updatedAt);

  console.log('[Main] Rebuilding history index from disk...');
  const files = await fsPromises.readdir(historyDir);
  const jsonFiles = files.filter(f => f.endsWith('.json') && f !== 'index.json');
  
  // Basic concurrency limit
  const chats = [];
  for (const file of jsonFiles) {
    const id = path.basename(file, '.json');
    const raw = await readChat(id);
    if (raw) {
      chats.push({
        id: raw.id || id,
        title: raw.title || 'Untitled Chat',
        updatedAt: raw.updatedAt || 0
      });
    }
  }

  const sorted = chats.sort((a, b) => b.updatedAt - a.updatedAt);
  await writeHistoryIndex(sorted);
  return sorted;
}

function registerIpc() {
  // Providers
  ipcMain.removeHandler('providers:getAll');
  ipcMain.handle('providers:getAll', async () => {
    console.log('[IPC] providers:getAll');
    return readProviders();
  });

  ipcMain.removeHandler('providers:save');
  ipcMain.handle('providers:save', async (_, data) => {
    console.log('[IPC] providers:save');
    await writeProviders(data);
    return readProviders();
  });

  // History
  ipcMain.removeHandler('history:list');
  ipcMain.handle('history:list', async () => {
    console.log('[IPC] history:list');
    return listChats();
  });

  ipcMain.removeHandler('history:get');
  ipcMain.handle('history:get', async (_, chatId) => {
    console.log(`[IPC] history:get -> ${chatId}`);
    return readChat(chatId);
  });

  ipcMain.removeHandler('history:save');
  ipcMain.handle('history:save', async (_, chatId, data) => {
    console.log(`[IPC] history:save -> ${chatId}`);
    await writeChat(chatId, data);
    const list = await listChats();
    const updated = list.filter(c => c.id !== chatId);
    const final = [{ id: chatId, title: data.title || 'Untitled Chat', updatedAt: Date.now() }, ...updated];
    await writeHistoryIndex(final);
    return final;
  });

  ipcMain.removeHandler('history:delete');
  ipcMain.handle('history:delete', async (_, chatId) => {
    validateChatId(chatId); // FIXED: Added validation
    console.log(`[IPC] history:delete -> ${chatId}`);
    const filePath = path.join(historyDir, `${chatId}.json`);
    if (await pathExists(filePath)) {
      await fsPromises.unlink(filePath);
    }
    const list = (await readHistoryIndex() || []).filter(c => c.id !== chatId);
    await writeHistoryIndex(list);
    return list;
  });

  // Window
  ipcMain.removeAllListeners('minimize-window');
  ipcMain.on('minimize-window', (e) => BrowserWindow.fromWebContents(e.sender)?.minimize());

  ipcMain.removeAllListeners('maximize-window');
  ipcMain.on('maximize-window', (e) => {
    const win = BrowserWindow.fromWebContents(e.sender);
    if (!win) return;
    if (win.isMaximized()) win.unmaximize();
    else win.maximize();
  });

  ipcMain.removeAllListeners('close-window');
  ipcMain.on('close-window', (e) => BrowserWindow.fromWebContents(e.sender)?.close());
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1400, height: 900,
    minWidth: 800, minHeight: 600,
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
  } else {
    win.loadFile(path.join(__dirname, '../dist/index.html'));
  }
}

app.whenReady().then(() => {
  registerIpc();
  createWindow();
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) createWindow();
});
