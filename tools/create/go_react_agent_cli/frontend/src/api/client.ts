import { getRoutePrefix } from '../routePrefix';

export function apiFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
    if (input instanceof Request) {
        const nextURL = apiUrl(input.url);
        const request = nextURL === input.url ? input : new Request(nextURL, input);
        return fetch(request, init);
    }
    return fetch(apiUrl(input), init);
}

export function apiEventSource(url: string | URL, eventSourceInitDict?: EventSourceInit): EventSource {
    return new EventSource(apiUrl(url), eventSourceInitDict);
}

export function apiUrl(input: string | URL): string {
    const raw = String(input);
    const prefix = getRoutePrefix();
    if (!prefix) {
        return raw;
    }

    if (raw.startsWith('/')) {
        return prefixRootPath(raw, prefix);
    }

    try {
        const url = new URL(raw, window.location.origin);
        if (url.origin === window.location.origin) {
            url.pathname = prefixRootPath(url.pathname, prefix);
            return url.toString();
        }
    } catch {
        // Non-URL inputs are left for fetch/EventSource to handle.
    }

    return raw;
}

function prefixRootPath(pathname: string, prefix: string): string {
    if (pathname === prefix || pathname.startsWith(`${prefix}/`)) {
        return pathname;
    }
    return `${prefix}${pathname}`;
}
