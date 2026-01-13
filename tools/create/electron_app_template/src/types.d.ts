declare global {
    interface ElectronAPI {
        getEnv: () => { [key: string]: string | undefined };
        openExternal: (url: string) => Promise<void>;
    }

    interface Window {
        electronAPI: ElectronAPI;
    }
}

export { };
