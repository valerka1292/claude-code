const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  minimize: () => ipcRenderer.send('minimize'),
  maximize: () => ipcRenderer.send('maximize'),
  close: () => ipcRenderer.send('close'),
  platform: ipcRenderer.invoke('get-platform'),

  selectFolder: () => ipcRenderer.invoke('select-folder'),
  resolvePath: (p) => ipcRenderer.invoke('resolve-path', p),

  loadSettings: () => ipcRenderer.invoke('load-settings'),
  saveSettings: (data) => ipcRenderer.invoke('save-settings', data),

  listSessions: (projectKey) => ipcRenderer.invoke('list-sessions', projectKey),
  loadSession: (projectKey, id) => ipcRenderer.invoke('load-session', projectKey, id),
  saveSession: (projectKey, session) => ipcRenderer.invoke('save-session', projectKey, session),
  deleteSession: (projectKey, id) => ipcRenderer.invoke('delete-session', projectKey, id),

  glob: (pattern, options) => ipcRenderer.invoke('glob', pattern, options),

  stat: (filePath) => ipcRenderer.invoke('stat', filePath),
});