import { apiFetch } from './client';

/** Allocate the next id (POST /api/counter) and return it. */
export async function nextCounter(): Promise<number> {
    const res = await apiFetch('/api/counter', { method: 'POST' });
    if (!res.ok) {
        throw new Error(`POST /api/counter: ${res.status} ${res.statusText}`);
    }
    const body = (await res.json()) as { id: number };
    return body.id;
}
