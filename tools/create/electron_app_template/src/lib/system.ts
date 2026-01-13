// Simple request helper for Electron app
import { ENV_API_BASE } from "../../shared/constants";

export const API_BASE = (() => {
    // Default
    let base = 'https://localhost:8080';

    // Check environment
    // Use try-catch or optional chaining safely
    const electronEnv = (window as any).electronAPI?.getEnv?.() || {};
    const isDev = (import.meta as any).env?.DEV || electronEnv.NODE_ENV === 'development';
    const envApiBase = electronEnv[ENV_API_BASE];

    if (envApiBase) {
        base = envApiBase;
    } else if (isDev) {
        base = 'http://localhost:8008';
    }

    console.log(`[API] Base URL: ${base} (isDev: ${isDev})`);
    return base;
})();

interface ApiResponse<T = any> {
    code: number;
    msg?: string;
    data: T;
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const electronAPI = (window as any).electronAPI;

    // Use IPC if available (handles CORS and cookies via Main process)
    if (electronAPI && electronAPI.request) {
        console.log(`[IPC Request] ${options.method || 'GET'} ${path}`);
        const response = await electronAPI.request({
            path,
            options: {
                method: options.method,
                headers: options.headers,
                body: options.body
            }
        });

        console.log(`[IPC Response] ${response.status}`, response);

        if (!response.ok) {
            throw new Error(`Request failed: ${response.status} ${response.statusText}`);
        }

        const json = response.data;
        // Standard API response check
        if (json && typeof json === 'object' && 'code' in json) {
            if (json.code !== 0) {
                console.error(`[Response Error] Code: ${json.code}, Msg: ${json.msg}`);
                throw new Error(json.msg || `Server error code: ${json.code}`);
            }
            return json.data;
        }
        return json as T;
    }

    // Fallback to fetch (web mode)
    const url = `${API_BASE}${path}`;
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    console.log(`[Request] ${options.method || 'GET'} ${url}`, { credentials: 'include', headers });

    const response = await fetch(url, {
        ...options,
        headers,
        credentials: 'include'
    });

    console.log(`[Response] ${response.status} ${response.statusText}`);

    if (!response.ok) {
        throw new Error(`Request failed: ${response.status} ${response.statusText}`);
    }

    let json: ApiResponse<T>;
    try {
        json = await response.json();
        console.log(`[Response Data]`, json);
    } catch (e) {
        console.error('[Response Error] Invalid JSON', e);
        throw new Error('Invalid JSON response');
    }

    if (json.code !== 0) {
        console.error(`[Response Error] Code: ${json.code}, Msg: ${json.msg}`);
        throw new Error(json.msg || `Server error code: ${json.code}`);
    }

    return json.data;
}

export function get<T>(path: string) {
    return request<T>(path, { method: 'GET' });
}

export function postJSON<T>(path: string, body: any) {
    return request<T>(path, {
        method: 'POST',
        body: JSON.stringify(body),
    });
}
