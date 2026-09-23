import { apiFetch } from './client';

/** One card's server-owned meta: its title, the hint under it, its empty state. */
export interface SectionMeta {
    title: string;
    hint: string;
    empty: string;
}

/** One card of a page document: meta, whether it has rows, and the rows. */
export interface PageSection<Items = unknown> {
    key: string;
    meta: SectionMeta;
    empty: boolean;
    items?: Items;
}

/** The page document the server renders from its card registry. */
export interface PageDoc {
    page: { title: string; url: string };
    sections: PageSection[];
}

/** Rows of the demo counter card. */
export interface CounterItems {
    last: number;
}

async function getJSON<T>(path: string): Promise<T> {
    const res = await apiFetch(path);
    if (!res.ok) {
        throw new Error(`GET ${path}: ${res.status} ${res.statusText}`);
    }
    return (await res.json()) as T;
}

/** The Home page document: cards in web order, each carrying its meta. */
export function getHomePage(): Promise<PageDoc> {
    return getJSON<PageDoc>('/api/pages/home');
}

/**
 * One card of a page document, typed to the rows that card carries. The cast is
 * the price of one document holding heterogeneous cards; it stays here, at the
 * API boundary, so components get a typed section.
 */
export function sectionOf<Items>(doc: PageDoc, key: string): PageSection<Items> | undefined {
    return doc.sections.find((section) => section.key === key) as PageSection<Items> | undefined;
}

/** Every registered card's meta, without walking the pages. */
export function getPageMeta(): Promise<{ sections: Record<string, SectionMeta> }> {
    return getJSON<{ sections: Record<string, SectionMeta> }>('/api/page-meta');
}
