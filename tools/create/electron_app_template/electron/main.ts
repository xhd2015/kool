import { app, BrowserWindow, ipcMain, shell } from 'electron';
import path from 'path';
import { isDev, getResourceUrl } from './env';

let mainWindow: BrowserWindow | null = null;

function getAssetPath(relativePath: string) {
    if (isDev) {
        return path.join(app.getAppPath(), 'public', relativePath);
    }
    return path.join(app.getAppPath(), 'dist', relativePath);
}

function loadWindowContent(win: BrowserWindow) {
    const resourceUrl = getResourceUrl();

    if (resourceUrl) {
        win.loadURL(resourceUrl);
        console.log("DEBUG: Loaded window content from:", resourceUrl);
    } else {
        win.loadFile(path.join(app.getAppPath(), 'dist/index.html'));
        console.log("DEBUG: Loaded window content from file:", path.join(app.getAppPath(), 'dist/index.html'));
    }
}

function createWindow() {
    const iconPath = getAssetPath('icon.png');

    mainWindow = new BrowserWindow({
        width: 1200,
        height: 800,
        icon: iconPath,
        webPreferences: {
            preload: path.join(__dirname, 'preload.js'),
            nodeIntegration: false,
            contextIsolation: true,
        },
        show: false, // Don't show until ready-to-show
    });

    // Show window when ready to avoid flickering
    mainWindow.once('ready-to-show', () => {
        console.log("DEBUG: ready-to-show event fired");
        mainWindow?.show();
    });

    loadWindowContent(mainWindow);

    mainWindow.on('close', (event) => {
        if (!app.isQuitting) {
            // Uncomment if you want to minimize to tray instead of quitting
            // event.preventDefault();
            // mainWindow?.hide();
        }
        return false;
    });
}

// Custom property to track quitting
declare global {
    namespace Electron {
        interface App {
            isQuitting: boolean;
        }
    }
}
app.isQuitting = false;

app.whenReady().then(() => {
    ipcMain.handle('open-external', async (event, url) => {
        await shell.openExternal(url);
    });

    createWindow();

    app.on('activate', () => {
        if (BrowserWindow.getAllWindows().length === 0) createWindow();
        else mainWindow?.show();
    });
});

app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
        app.quit();
    }
});

// Clean up
app.on('before-quit', () => {
    console.log("DEBUG: before-quit");
    app.isQuitting = true;
});
