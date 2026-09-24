declare global {
    interface Window {
        __KOOL_ROUTE_PREFIX__?: string;
    }
}

/**
 * The path prefix this page is served under, e.g. "/demo".
 *
 * The server injects the value per request, because one process can serve a
 * prefixed front door and a direct domain at the same time; the build-time base
 * is only the fallback (Vite dev sets it from --base).
 */
export function getRoutePrefix(): string {
    const injected = injectedRoutePrefix();
    if (injected !== undefined) {
        // An explicit empty injected value means the root route: trust it over
        // the build base, which may still carry a prefix (Vite dev --base).
        return normalizePrefix(injected);
    }
    return normalizePrefix(import.meta.env.BASE_URL ?? '');
}

/** injectedRoutePrefix is the page global, or undefined outside a browser. */
function injectedRoutePrefix(): string | undefined {
    if (typeof window === 'undefined') {
        return undefined;
    }
    return typeof window.__KOOL_ROUTE_PREFIX__ === 'string' ? window.__KOOL_ROUTE_PREFIX__ : undefined;
}

export function normalizePrefix(prefix: string): string {
    prefix = prefix.trim();
    if (!prefix || prefix === '/') {
        return '';
    }
    return `/${prefix.replace(/^\/+|\/+$/g, '')}`;
}
