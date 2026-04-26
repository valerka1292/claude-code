const { contextBridge, ipcRenderer } = require('electron');

function ensureString(value, fieldName) {
  if (typeof value !== 'string') {
    throw new Error(`${fieldName} must be a string`);
  }
  return value;
}

contextBridge.exposeInMainWorld('electronAPI', {
  minimize: () => ipcRenderer.send('minimize'),
  maximize: () => ipcRenderer.send('maximize'),
  close: () => ipcRenderer.send('close'),
  platform: ipcRenderer.invoke('get-platform'),

  selectFolder: () => ipcRenderer.invoke('select-folder'),
  resolvePath: (p) => ipcRenderer.invoke('resolve-path', ensureString(p, 'path')),

  loadSettings: () => ipcRenderer.invoke('load-settings'),
  saveSettings: (data) => ipcRenderer.invoke('save-settings', data),

  listSessions: (projectKey) => ipcRenderer.invoke('list-sessions', ensureString(projectKey, 'projectKey')),
  loadSession: (projectKey, id) =>
    ipcRenderer.invoke(
      'load-session',
      ensureString(projectKey, 'projectKey'),
      ensureString(id, 'id')
    ),
  saveSession: (projectKey, session) =>
    ipcRenderer.invoke('save-session', ensureString(projectKey, 'projectKey'), session),
  deleteSession: (projectKey, id) =>
    ipcRenderer.invoke(
      'delete-session',
      ensureString(projectKey, 'projectKey'),
      ensureString(id, 'id')
    ),

  glob: (pattern, options) =>
    ipcRenderer.invoke('glob', ensureString(pattern, 'pattern'), options),

  stat: (filePath, cwd) =>
    ipcRenderer.invoke(
      'stat',
      ensureString(filePath, 'filePath'),
      cwd === undefined ? undefined : ensureString(cwd, 'cwd')
    ),
});
