import { contextBridge, ipcRenderer } from 'electron';

contextBridge.exposeInMainWorld('electronAPI', {
    getEnv: () => ({
        NODE_ENV: process.env.NODE_ENV
    }),
    openExternal: (url: string) => ipcRenderer.invoke('open-external', url),
});
