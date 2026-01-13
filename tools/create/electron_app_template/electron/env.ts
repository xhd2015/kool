import { ENV_APP_RESOURCE_URL } from "../shared/constants";

export const isDev = process.env.NODE_ENV === 'development';

export function getResourceUrl() {
    if (process.env[ENV_APP_RESOURCE_URL]) {
        return process.env[ENV_APP_RESOURCE_URL];
    }
    if (isDev) {
        return 'http://localhost:5173';
    }
    return null;
}
