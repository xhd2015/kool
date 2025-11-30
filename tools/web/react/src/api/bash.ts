export interface HistoryResponse {
    list: string[];
    total: number;
}

export interface HistoryParams {
    page: number;
    pageSize: number;
    search?: string;
}

export const bashApi = {
    list: async (params: HistoryParams): Promise<HistoryResponse> => {
        const query = new URLSearchParams({
            page: params.page.toString(),
            pageSize: params.pageSize.toString(),
            search: params.search || '',
        });
        const res = await fetch(`/api/bash/history?${query.toString()}`);
        if (!res.ok) {
            throw new Error(`Failed to fetch history: ${res.statusText}`);
        }
        return res.json();
    },
    delete: async (cmd: string): Promise<void> => {
        const res = await fetch('/api/bash/history/delete', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ cmd }),
        });
        if (!res.ok) {
            throw new Error(`Failed to delete command: ${res.statusText}`);
        }
    },
};

