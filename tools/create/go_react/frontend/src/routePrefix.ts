declare global {
    interface Window {
        __KOOL_ROUTE_PREFIX__?: string;
    }
}

export function getRoutePrefix(): string {
    const configured = normalizePrefix(window.__KOOL_ROUTE_PREFIX__ ?? '');
    if (configured) {
        return configured;
    }
    return normalizePrefix(import.meta.env.BASE_URL ?? '');
}

export function normalizePrefix(prefix: string): string {
    prefix = prefix.trim();
    if (!prefix || prefix === '/') {
        return '';
    }
    return `/${prefix.replace(/^\/+|\/+$/g, '')}`;
}
